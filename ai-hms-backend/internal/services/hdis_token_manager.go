package services

import (
	"context"
	"errors"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/elliotxin/ai-hms-backend/config"
	"github.com/elliotxin/ai-hms-backend/internal/database"
	"github.com/elliotxin/ai-hms-backend/internal/integrations/hdis"
	"github.com/elliotxin/ai-hms-backend/internal/models"
	"github.com/elliotxin/ai-hms-backend/internal/utils"
	"gorm.io/gorm"
)

const (
	defaultHDISSettingID    = "default"
	defaultRefreshLeadSec   = 1800
	maxRefreshLeadSec       = 86400
	hdisTokenStatusMissing  = "MISSING"
	hdisTokenStatusUnknown  = "UNKNOWN"
	hdisTokenStatusValid    = "VALID"
	hdisTokenStatusExpiring = "EXPIRING"
	hdisTokenStatusExpired  = "EXPIRED"
)

var (
	ErrHDISAuthNotConfigured = errors.New("hdis auth settings are not configured")
	ErrHDISTokenUnavailable  = errors.New("hdis token is unavailable")
)

// HdisRuntimeConfig 运行时配置
type HdisRuntimeConfig struct {
	WebcmdURL  string
	GraphqlURL string
	Token      string
	Timeout    time.Duration
}

// HDISTokenManager 管理 HDIS token 获取与刷新
type HDISTokenManager struct {
	db     *gorm.DB
	cfg    config.HdisConfig
	secret *utils.SecretBox
	mu     sync.Mutex
}

// HdisTokenRefreshResult token 刷新结果
type HdisTokenRefreshResult struct {
	TokenExpiresAt *time.Time `json:"tokenExpiresAt"`
	TokenStatus    string     `json:"tokenStatus"`
}

// NewHDISTokenManager 创建 token manager
func NewHDISTokenManager(cfg config.HdisConfig) *HDISTokenManager {
	secret := strings.TrimSpace(cfg.Secret)
	return &HDISTokenManager{
		db:     database.GetDB(),
		cfg:    cfg,
		secret: utils.NewSecretBox(secret),
	}
}

// GetRuntimeConfig 获取可用于调用 HDIS 的运行时配置（仅来源于数据库 integration_hdis_settings）。
// 会在 token 临期或过期时按需刷新。
func (m *HDISTokenManager) GetRuntimeConfig(ctx context.Context) (*HdisRuntimeConfig, error) {
	timeout := m.timeout()

	setting, exists, err := m.getSetting()
	if err != nil {
		return nil, err
	}

	if !exists || setting == nil {
		return nil, ErrSyncNotConfigured
	}

	webcmdURL := strings.TrimSpace(setting.WebcmdURL)
	graphqlURL := strings.TrimSpace(setting.GraphqlURL)
	if webcmdURL == "" || graphqlURL == "" {
		return nil, ErrSyncNotConfigured
	}

	leadSeconds := normalizeLeadSeconds(setting.RefreshLeadSeconds)
	now := time.Now()

	token := ""
	if strings.TrimSpace(setting.AccessTokenEncrypted) != "" {
		token, err = m.secret.Decrypt(setting.AccessTokenEncrypted)
		if err != nil {
			slog.Error("decrypt hdis token failed", "error", err)
			token = ""
		}
	}

	// token 可用且不在刷新窗口内，直接返回
	if token != "" && !shouldRefreshToken(setting.TokenExpiresAt, leadSeconds, now) {
		return &HdisRuntimeConfig{
			WebcmdURL:  webcmdURL,
			GraphqlURL: graphqlURL,
			Token:      token,
			Timeout:    timeout,
		}, nil
	}

	// 按需刷新
	if setting.AutoRefreshEnabled {
		if _, refreshErr := m.RefreshToken(ctx); refreshErr != nil {
			slog.Error("refresh hdis token failed", "error", refreshErr)
		} else {
			latest, ok, latestErr := m.getSetting()
			if latestErr == nil && ok && latest != nil && strings.TrimSpace(latest.AccessTokenEncrypted) != "" {
				if decrypted, decErr := m.secret.Decrypt(latest.AccessTokenEncrypted); decErr == nil && decrypted != "" {
					return &HdisRuntimeConfig{
						WebcmdURL:  strings.TrimSpace(latest.WebcmdURL),
						GraphqlURL: strings.TrimSpace(latest.GraphqlURL),
						Token:      decrypted,
						Timeout:    timeout,
					}, nil
				}
			}
		}
	}

	if token != "" && setting.TokenExpiresAt == nil {
		return &HdisRuntimeConfig{
			WebcmdURL:  webcmdURL,
			GraphqlURL: graphqlURL,
			Token:      token,
			Timeout:    timeout,
		}, nil
	}

	return nil, ErrHDISTokenUnavailable
}

// RefreshToken 手动触发 token 刷新
func (m *HDISTokenManager) RefreshToken(ctx context.Context) (*HdisTokenRefreshResult, error) {
	if m.db == nil {
		return nil, errors.New("database not available")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	setting, exists, err := m.getSetting()
	if err != nil {
		return nil, err
	}
	if !exists || setting == nil {
		return nil, ErrHDISAuthNotConfigured
	}
	if strings.TrimSpace(setting.AuthURL) == "" ||
		strings.TrimSpace(setting.ClientID) == "" ||
		strings.TrimSpace(setting.ServiceUsername) == "" ||
		strings.TrimSpace(setting.ServicePasswordEncrypted) == "" {
		return nil, ErrHDISAuthNotConfigured
	}

	password, err := m.secret.Decrypt(setting.ServicePasswordEncrypted)
	if err != nil {
		return nil, err
	}

	client := hdis.NewBrowserAuthClient(m.cfg.BrowserHeadless, m.browserTimeout())
	resp, err := client.RefreshToken(ctx, hdis.BrowserTokenRefreshRequest{
		AuthURL:       setting.AuthURL,
		ClientID:      setting.ClientID,
		Username:      setting.ServiceUsername,
		Password:      password,
		GraphqlURL:    strings.TrimSpace(setting.GraphqlURL),
		WebcmdURL:     strings.TrimSpace(setting.WebcmdURL),
		TargetOrganID: m.cfg.TargetOrganID,
	})
	if err != nil {
		setting.LastError = err.Error()
		_ = m.db.Save(setting).Error
		return nil, err
	}

	encryptedToken, err := m.secret.Encrypt(resp.AccessToken)
	if err != nil {
		return nil, err
	}

	setting.AccessTokenEncrypted = encryptedToken
	setting.TokenExpiresAt = &resp.ExpiresAt
	setting.LastError = ""
	if err := m.db.Save(setting).Error; err != nil {
		return nil, err
	}

	return &HdisTokenRefreshResult{
		TokenExpiresAt: setting.TokenExpiresAt,
		TokenStatus:    resolveTokenStatus(true, setting.TokenExpiresAt, normalizeLeadSeconds(setting.RefreshLeadSeconds), time.Now()),
	}, nil
}

func (m *HDISTokenManager) getSetting() (*models.IntegrationHDISSetting, bool, error) {
	if m.db == nil {
		return nil, false, nil
	}

	var setting models.IntegrationHDISSetting
	res := m.db.Where("id = ?", defaultHDISSettingID).Limit(1).Find(&setting)
	if res.Error != nil {
		return nil, false, res.Error
	}
	if res.RowsAffected == 0 {
		return nil, false, nil
	}
	return &setting, true, nil
}

func (m *HDISTokenManager) timeout() time.Duration {
	timeout := time.Duration(m.cfg.TimeoutSeconds) * time.Second
	if timeout <= 0 {
		timeout = 15 * time.Second
	}
	return timeout
}

func (m *HDISTokenManager) browserTimeout() time.Duration {
	timeout := time.Duration(m.cfg.BrowserTimeoutSeconds) * time.Second
	if timeout <= 0 {
		timeout = 120 * time.Second
	}
	return timeout
}

func shouldRefreshToken(expiresAt *time.Time, leadSeconds int, now time.Time) bool {
	if expiresAt == nil {
		return false
	}
	lead := time.Duration(normalizeLeadSeconds(leadSeconds)) * time.Second
	return !now.Before(expiresAt.Add(-lead))
}

func resolveTokenStatus(hasToken bool, expiresAt *time.Time, leadSeconds int, now time.Time) string {
	if !hasToken {
		return hdisTokenStatusMissing
	}
	if expiresAt == nil {
		return hdisTokenStatusUnknown
	}
	if now.After(*expiresAt) {
		return hdisTokenStatusExpired
	}
	if shouldRefreshToken(expiresAt, leadSeconds, now) {
		return hdisTokenStatusExpiring
	}
	return hdisTokenStatusValid
}

func normalizeLeadSeconds(leadSeconds int) int {
	if leadSeconds <= 0 {
		return defaultRefreshLeadSec
	}
	if leadSeconds > maxRefreshLeadSec {
		return maxRefreshLeadSec
	}
	return leadSeconds
}

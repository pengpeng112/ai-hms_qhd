package services

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/elliotxin/ai-hms-backend/config"
	"github.com/elliotxin/ai-hms-backend/internal/database"
	"github.com/elliotxin/ai-hms-backend/internal/models"
	"github.com/elliotxin/ai-hms-backend/internal/utils"
	"gorm.io/gorm"
)

var (
	ErrHDISSettingsInvalidInput = errors.New("invalid hdis settings input")
)

// HdisIntegrationSettingsView 设置页响应
type HdisIntegrationSettingsView struct {
	WebcmdURL  string `json:"webcmdUrl"`
	GraphqlURL string `json:"graphqlUrl"`
	AuthURL    string `json:"authUrl"`
	ClientID   string `json:"clientId"`

	ServiceUsername           string `json:"serviceUsername"`
	ServicePasswordConfigured bool   `json:"servicePasswordConfigured"`

	AutoRefreshEnabled bool `json:"autoRefreshEnabled"`
	RefreshLeadSeconds int  `json:"refreshLeadSeconds"`

	TokenConfigured bool       `json:"tokenConfigured"`
	TokenExpiresAt  *time.Time `json:"tokenExpiresAt"`
	TokenStatus     string     `json:"tokenStatus"`
	LastError       string     `json:"lastError"`
}

// HdisIntegrationSettingsUpdateRequest 更新设置请求
type HdisIntegrationSettingsUpdateRequest struct {
	WebcmdURL  string `json:"webcmdUrl"`
	GraphqlURL string `json:"graphqlUrl"`
	AuthURL    string `json:"authUrl"`
	ClientID   string `json:"clientId"`

	ServiceUsername string `json:"serviceUsername"`
	ServicePassword string `json:"servicePassword,omitempty"`

	AutoRefreshEnabled bool `json:"autoRefreshEnabled"`
	RefreshLeadSeconds int  `json:"refreshLeadSeconds"`
}

// HDISSettingsService HDIS 集成设置服务
type HDISSettingsService struct {
	db           *gorm.DB
	cfg          config.HdisConfig
	tokenManager *HDISTokenManager
	secret       *utils.SecretBox
}

func NewHDISSettingsService(cfg config.HdisConfig) *HDISSettingsService {
	secret := strings.TrimSpace(cfg.Secret)
	return &HDISSettingsService{
		db:           database.GetDB(),
		cfg:          cfg,
		tokenManager: NewHDISTokenManager(cfg),
		secret:       utils.NewSecretBox(secret),
	}
}

func (s *HDISSettingsService) GetSettings() (*HdisIntegrationSettingsView, error) {
	setting, exists, err := s.findSetting()
	if err != nil {
		return nil, err
	}

	now := time.Now()
	if !exists || setting == nil {
		return &HdisIntegrationSettingsView{
			WebcmdURL:                 "",
			GraphqlURL:                "",
			AuthURL:                   strings.TrimSpace(s.cfg.AuthURL),
			ClientID:                  strings.TrimSpace(s.cfg.ClientID),
			ServiceUsername:           strings.TrimSpace(s.cfg.ServiceUser),
			ServicePasswordConfigured: strings.TrimSpace(s.cfg.ServicePass) != "",
			AutoRefreshEnabled:        true,
			RefreshLeadSeconds:        defaultRefreshLeadSec,
			TokenConfigured:           false,
			TokenExpiresAt:            nil,
			TokenStatus:               resolveTokenStatus(false, nil, defaultRefreshLeadSec, now),
			LastError:                 "",
		}, nil
	}

	lead := normalizeLeadSeconds(setting.RefreshLeadSeconds)
	hasToken := strings.TrimSpace(setting.AccessTokenEncrypted) != ""
	return &HdisIntegrationSettingsView{
		WebcmdURL:                 setting.WebcmdURL,
		GraphqlURL:                setting.GraphqlURL,
		AuthURL:                   setting.AuthURL,
		ClientID:                  setting.ClientID,
		ServiceUsername:           setting.ServiceUsername,
		ServicePasswordConfigured: strings.TrimSpace(setting.ServicePasswordEncrypted) != "",
		AutoRefreshEnabled:        setting.AutoRefreshEnabled,
		RefreshLeadSeconds:        lead,
		TokenConfigured:           hasToken,
		TokenExpiresAt:            setting.TokenExpiresAt,
		TokenStatus:               resolveTokenStatus(hasToken, setting.TokenExpiresAt, lead, now),
		LastError:                 setting.LastError,
	}, nil
}

func (s *HDISSettingsService) UpdateSettings(req HdisIntegrationSettingsUpdateRequest, updatedBy string) (*HdisIntegrationSettingsView, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}
	setting, exists, err := s.findSetting()
	if err != nil {
		return nil, err
	}

	missingFields := validateHDISSettingsInput(req, exists && setting != nil && strings.TrimSpace(setting.ServicePasswordEncrypted) != "")
	if len(missingFields) > 0 {
		return nil, fmt.Errorf("%w: missing fields: %s", ErrHDISSettingsInvalidInput, strings.Join(missingFields, ", "))
	}

	if !exists || setting == nil {
		setting = &models.IntegrationHDISSetting{
			ID: defaultHDISSettingID,
		}
	}

	setting.WebcmdURL = strings.TrimSpace(req.WebcmdURL)
	setting.GraphqlURL = strings.TrimSpace(req.GraphqlURL)
	setting.AuthURL = strings.TrimSpace(req.AuthURL)
	setting.ClientID = strings.TrimSpace(req.ClientID)
	setting.ServiceUsername = strings.TrimSpace(req.ServiceUsername)
	setting.AutoRefreshEnabled = req.AutoRefreshEnabled
	setting.RefreshLeadSeconds = normalizeLeadSeconds(req.RefreshLeadSeconds)
	setting.UpdatedBy = strings.TrimSpace(updatedBy)

	if strings.TrimSpace(req.ServicePassword) != "" {
		encrypted, encErr := s.secret.Encrypt(req.ServicePassword)
		if encErr != nil {
			return nil, encErr
		}
		setting.ServicePasswordEncrypted = encrypted
	}
	if strings.TrimSpace(setting.ServicePasswordEncrypted) == "" {
		return nil, ErrHDISSettingsInvalidInput
	}

	if exists {
		if err := s.db.Save(setting).Error; err != nil {
			return nil, err
		}
	} else {
		if err := s.db.Create(setting).Error; err != nil {
			return nil, err
		}
	}

	return s.GetSettings()
}

func (s *HDISSettingsService) RefreshToken(ctx context.Context) (*HdisTokenRefreshResult, error) {
	return s.tokenManager.RefreshToken(ctx)
}

func (s *HDISSettingsService) findSetting() (*models.IntegrationHDISSetting, bool, error) {
	if s.db == nil {
		return nil, false, nil
	}
	var setting models.IntegrationHDISSetting
	res := s.db.Where("id = ?", defaultHDISSettingID).Limit(1).Find(&setting)
	if res.Error != nil {
		return nil, false, res.Error
	}
	if res.RowsAffected == 0 {
		return nil, false, nil
	}
	return &setting, true, nil
}

func validateHDISSettingsInput(req HdisIntegrationSettingsUpdateRequest, hasStoredPassword bool) []string {
	missing := make([]string, 0, 6)
	if strings.TrimSpace(req.WebcmdURL) == "" {
		missing = append(missing, "webcmdUrl")
	}
	if strings.TrimSpace(req.GraphqlURL) == "" {
		missing = append(missing, "graphqlUrl")
	}
	if strings.TrimSpace(req.AuthURL) == "" {
		missing = append(missing, "authUrl")
	}
	if strings.TrimSpace(req.ClientID) == "" {
		missing = append(missing, "clientId")
	}
	if strings.TrimSpace(req.ServiceUsername) == "" {
		missing = append(missing, "serviceUsername")
	}
	if strings.TrimSpace(req.ServicePassword) == "" && !hasStoredPassword {
		missing = append(missing, "servicePassword")
	}
	return missing
}

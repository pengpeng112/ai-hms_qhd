package services

import (
	"strings"
	"sync"
	"time"

	"github.com/elliotxin/ai-hms-backend/internal/config"
	"github.com/elliotxin/ai-hms-backend/internal/database"
	"github.com/elliotxin/ai-hms-backend/internal/models"
	"gorm.io/gorm"
)

const thresholdCacheTTL = 5 * time.Second

var thresholdCache = struct {
	mu   sync.RWMutex
	data *config.MonitoringThresholds
	exp  time.Time
}{}

// InvalidateThresholdCache 在 admin 写入后调用，使下次读取强制回源。
func InvalidateThresholdCache() {
	thresholdCache.mu.Lock()
	thresholdCache.data = nil
	thresholdCache.exp = time.Time{}
	thresholdCache.mu.Unlock()
}

// loadThresholdsCached 带 5s 缓存地取阈值表（DB 优先、分部分回退 JSON）。永不返回 nil。
// 返回值是共享缓存对象，调用方只读、禁止修改。
func loadThresholdsCached() *config.MonitoringThresholds {
	thresholdCache.mu.RLock()
	if thresholdCache.data != nil && time.Now().Before(thresholdCache.exp) {
		d := thresholdCache.data
		thresholdCache.mu.RUnlock()
		return d
	}
	thresholdCache.mu.RUnlock()

	d := buildThresholds(database.GetDB(), config.LoadMonitoringThresholds)

	thresholdCache.mu.Lock()
	thresholdCache.data = d
	thresholdCache.exp = time.Now().Add(thresholdCacheTTL)
	thresholdCache.mu.Unlock()
	return d
}

// jsonLoader 内嵌 JSON 加载函数签名（便于测试注入）。
type jsonLoader func() (*config.MonitoringThresholds, error)

// buildThresholds 组装阈值表：DB 优先 + 分部分回退 fallback()。db 为 nil 或表缺失时全量回退。
func buildThresholds(db *gorm.DB, fallback jsonLoader) *config.MonitoringThresholds {
	def, _ := fallback()
	if def == nil {
		def = &config.MonitoringThresholds{Fixed: map[string]config.FixedThreshold{}}
	}
	if db == nil {
		return def
	}

	var fixedRows []models.MonitoringThreshold
	if err := db.Where("tenant_id = ? AND scope = ?", LegacyTenantID, models.ScopeGlobal).Find(&fixedRows).Error; err != nil {
		return def
	}

	out := &config.MonitoringThresholds{
		Fixed:             map[string]config.FixedThreshold{},
		DialysateNaFactor: def.DialysateNaFactor,
	}

	// 固定阈值：DB 有则用 DB，空则回退默认。
	if len(fixedRows) == 0 {
		out.Fixed = def.Fixed
	} else {
		for _, r := range fixedRows {
			out.Fixed[r.MetricKey] = config.FixedThreshold{
				Label:      r.Label,
				Enabled:    r.Enabled,
				Unit:       r.Unit,
				DangerLow:  r.DangerLow,
				WarnLow:    r.WarnLow,
				WarnHigh:   r.WarnHigh,
				DangerHigh: r.DangerHigh,
			}
		}
	}

	// VP 分层：DB 有则用 DB，空则回退默认。
	var vpRows []models.MonitoringVPStratum
	if err := db.Where("tenant_id = ?", LegacyTenantID).Order("access asc, bf_min asc").Find(&vpRows).Error; err != nil || len(vpRows) == 0 {
		out.VPReference = def.VPReference
	} else {
		for _, r := range vpRows {
			if !r.Enabled {
				continue
			}
			out.VPReference = append(out.VPReference, config.VPStratum{
				Access:     strings.ToUpper(r.Access),
				BFMin:      r.BFMin,
				BFMax:      r.BFMax,
				NormalLow:  r.NormalLow,
				WarnHigh:   r.WarnHigh,
				DangerHigh: r.DangerHigh,
			})
		}
		if len(out.VPReference) == 0 {
			out.VPReference = def.VPReference
		}
	}

	// naFactor：setting 有则用，否则保留默认（NaFactor() 再兜底 9.9）。
	var setting models.MonitoringSetting
	if err := db.Where("tenant_id = ? AND setting_key = ?", LegacyTenantID, models.SettingKeyDialysateNaFactor).First(&setting).Error; err == nil && setting.ValueNum != nil {
		out.DialysateNaFactor = *setting.ValueNum
	}

	return out
}

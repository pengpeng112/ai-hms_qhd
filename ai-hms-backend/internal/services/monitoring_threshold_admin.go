package services

import (
	"errors"
	"sort"
	"time"

	"github.com/elliotxin/ai-hms-backend/internal/config"
	"github.com/elliotxin/ai-hms-backend/internal/database"
	"github.com/elliotxin/ai-hms-backend/internal/models"
	"gorm.io/gorm"
)

// ThresholdAdminPayload admin 页整体读写的载荷（与前端契约一致）。
type ThresholdAdminPayload struct {
	Fixed       []FixedThresholdDTO `json:"fixed"`
	VPReference []VPStratumDTO      `json:"vpReference"`
	NaFactor    float64             `json:"naFactor"`
}

type FixedThresholdDTO struct {
	MetricKey  string   `json:"metricKey"`
	Label      string   `json:"label"`
	Unit       string   `json:"unit"`
	DangerLow  *float64 `json:"dangerLow"`
	WarnLow    *float64 `json:"warnLow"`
	WarnHigh   *float64 `json:"warnHigh"`
	DangerHigh *float64 `json:"dangerHigh"`
	Basis      string   `json:"basis"`
	Enabled    bool     `json:"enabled"`
	SortOrder  int      `json:"sortOrder"`
}

type VPStratumDTO struct {
	Access     string  `json:"access"`
	BFMin      float64 `json:"bfMin"`
	BFMax      float64 `json:"bfMax"`
	NormalLow  float64 `json:"normalLow"`
	WarnHigh   float64 `json:"warnHigh"`
	DangerHigh float64 `json:"dangerHigh"`
	Basis      string  `json:"basis"`
	Enabled    bool    `json:"enabled"`
}

var validVPAccess = map[string]bool{"AVF": true, "AVG": true, "TCC": true, "NCC": true}

// ErrThresholdTablesMissing 表缺失（部署未建表）时返回，handler 映射为 503。
var ErrThresholdTablesMissing = errors.New("阈值表未部署，请先执行 deploy_new_tables.sql 建表")

// thresholdTablesReady 检查三张监控阈值表是否全部存在。
func thresholdTablesReady(db *gorm.DB) bool {
	return db.Migrator().HasTable(&models.MonitoringThreshold{}) &&
		db.Migrator().HasTable(&models.MonitoringVPStratum{}) &&
		db.Migrator().HasTable(&models.MonitoringSetting{})
}

// ValidatePayload 校验单调性与区间合法性。返回首个错误。
func ValidatePayload(p ThresholdAdminPayload) error {
	if p.NaFactor <= 0 {
		return errors.New("透析液钠电导率系数必须大于 0")
	}
	for _, f := range p.Fixed {
		if f.MetricKey == "" {
			return errors.New("固定阈值缺少 metricKey")
		}
		seq := []*float64{f.DangerLow, f.WarnLow, f.WarnHigh, f.DangerHigh}
		var prev *float64
		for _, cur := range seq {
			if cur == nil {
				continue
			}
			if prev != nil && *cur < *prev {
				return errors.New("固定阈值[" + f.MetricKey + "]四档需单调不降：危险低≤警戒低≤警戒高≤危险高")
			}
			prev = cur
		}
	}
	for _, v := range p.VPReference {
		if !validVPAccess[v.Access] {
			return errors.New("VP 分层通路类型非法（仅 AVF/AVG/TCC/NCC）：" + v.Access)
		}
		if v.BFMin >= v.BFMax {
			return errors.New("VP 分层 bfMin 必须小于 bfMax")
		}
		if !(v.NormalLow <= v.WarnHigh && v.WarnHigh <= v.DangerHigh) {
			return errors.New("VP 分层需满足 normalLow≤warnHigh≤dangerHigh")
		}
	}
	return nil
}

// GetThresholdAdmin 读取当前阈值表供 admin 展示与编辑。
func GetThresholdAdmin() ThresholdAdminPayload {
	db := database.GetDB()
	if db != nil && thresholdTablesReady(db) {
		var fixedRows []models.MonitoringThreshold
		if err := db.Where("tenant_id = ? AND scope = ?", LegacyTenantID, models.ScopeGlobal).
			Order("sort_order asc, metric_key asc").Find(&fixedRows).Error; err == nil && len(fixedRows) > 0 {
			return getThresholdAdminFromDB(db, fixedRows)
		}
	}
	return configToPayload(buildThresholds(db, config.LoadMonitoringThresholds))
}

func getThresholdAdminFromDB(db *gorm.DB, fixedRows []models.MonitoringThreshold) ThresholdAdminPayload {
	out := ThresholdAdminPayload{NaFactor: 9.9}
	for _, r := range fixedRows {
		out.Fixed = append(out.Fixed, FixedThresholdDTO{
			MetricKey: r.MetricKey, Label: r.Label, Unit: r.Unit,
			DangerLow: r.DangerLow, WarnLow: r.WarnLow, WarnHigh: r.WarnHigh, DangerHigh: r.DangerHigh,
			Basis: r.Basis, Enabled: r.Enabled, SortOrder: r.SortOrder,
		})
	}
	var vpRows []models.MonitoringVPStratum
	db.Where("tenant_id = ?", LegacyTenantID).Order("access asc, bf_min asc").Find(&vpRows)
	for _, r := range vpRows {
		out.VPReference = append(out.VPReference, VPStratumDTO{
			Access: r.Access, BFMin: r.BFMin, BFMax: r.BFMax,
			NormalLow: r.NormalLow, WarnHigh: r.WarnHigh, DangerHigh: r.DangerHigh,
			Basis: r.Basis, Enabled: r.Enabled,
		})
	}
	var setting models.MonitoringSetting
	if err := db.Where("tenant_id = ? AND setting_key = ?", LegacyTenantID, models.SettingKeyDialysateNaFactor).First(&setting).Error; err == nil && setting.ValueNum != nil {
		out.NaFactor = *setting.ValueNum
	}
	return out
}

func configToPayload(t *config.MonitoringThresholds) ThresholdAdminPayload {
	out := ThresholdAdminPayload{NaFactor: t.NaFactor()}
	keys := make([]string, 0, len(t.Fixed))
	for k := range t.Fixed {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, key := range keys {
		f := t.Fixed[key]
		out.Fixed = append(out.Fixed, FixedThresholdDTO{
			MetricKey: key, Label: f.Label, Unit: f.Unit,
			DangerLow: f.DangerLow, WarnLow: f.WarnLow, WarnHigh: f.WarnHigh, DangerHigh: f.DangerHigh,
			Enabled: f.Enabled,
		})
	}
	for _, v := range t.VPReference {
		out.VPReference = append(out.VPReference, VPStratumDTO{
			Access: v.Access, BFMin: v.BFMin, BFMax: v.BFMax,
			NormalLow: v.NormalLow, WarnHigh: v.WarnHigh, DangerHigh: v.DangerHigh, Enabled: true,
		})
	}
	return out
}

// SaveThresholdAdmin 校验后事务性整体保存，并失缓存。
func SaveThresholdAdmin(p ThresholdAdminPayload, operatorID int64) error {
	if err := ValidatePayload(p); err != nil {
		return err
	}
	db := database.GetDB()
	if db == nil {
		return errors.New("database not available")
	}
	if !thresholdTablesReady(db) {
		return ErrThresholdTablesMissing
	}
	err := db.Transaction(func(tx *gorm.DB) error { return writeAllThresholds(tx, p, operatorID) })
	if err != nil {
		return err
	}
	InvalidateThresholdCache()
	return nil
}

// ResetThresholdAdmin 用内嵌 JSON 默认重新种子写回三表，并失缓存。
func ResetThresholdAdmin(operatorID int64) error {
	def, err := config.LoadMonitoringThresholds()
	if err != nil {
		return err
	}
	return SaveThresholdAdmin(configToPayload(def), operatorID)
}

func writeAllThresholds(tx *gorm.DB, p ThresholdAdminPayload, operatorID int64) error {
	now := time.Now()
	op := &operatorID

	if err := tx.Where("tenant_id = ? AND scope = ?", LegacyTenantID, models.ScopeGlobal).Delete(&models.MonitoringThreshold{}).Error; err != nil {
		return err
	}
	for _, f := range p.Fixed {
		row := models.MonitoringThreshold{
			TenantID: LegacyTenantID, MetricKey: f.MetricKey, Label: f.Label, Unit: f.Unit, Scope: models.ScopeGlobal,
			DangerLow: f.DangerLow, WarnLow: f.WarnLow, WarnHigh: f.WarnHigh, DangerHigh: f.DangerHigh,
			Basis: f.Basis, Enabled: f.Enabled, SortOrder: f.SortOrder,
			CreatedAt: now, UpdatedAt: now, LastModifyBy: op,
		}
		if err := tx.Create(&row).Error; err != nil {
			return err
		}
	}

	if err := tx.Where("tenant_id = ?", LegacyTenantID).Delete(&models.MonitoringVPStratum{}).Error; err != nil {
		return err
	}
	for _, v := range p.VPReference {
		row := models.MonitoringVPStratum{
			TenantID: LegacyTenantID, Access: v.Access, BFMin: v.BFMin, BFMax: v.BFMax,
			NormalLow: v.NormalLow, WarnHigh: v.WarnHigh, DangerHigh: v.DangerHigh,
			Basis: v.Basis, Enabled: v.Enabled, CreatedAt: now, UpdatedAt: now, LastModifyBy: op,
		}
		if err := tx.Create(&row).Error; err != nil {
			return err
		}
	}

	naVal := p.NaFactor
	setting := models.MonitoringSetting{
		TenantID: LegacyTenantID, SettingKey: models.SettingKeyDialysateNaFactor,
		ValueNum: &naVal, UpdatedAt: now, LastModifyBy: op,
	}
	if err := tx.Where("tenant_id = ? AND setting_key = ?", LegacyTenantID, models.SettingKeyDialysateNaFactor).
		Delete(&models.MonitoringSetting{}).Error; err != nil {
		return err
	}
	return tx.Create(&setting).Error
}

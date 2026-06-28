package models

import "time"

// MonitoringThreshold 实时监控固定阈值（五档），每指标一行。snake_case 列对齐既有新表。
type MonitoringThreshold struct {
	ID           int64     `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	TenantID     int64     `gorm:"column:tenant_id;not null" json:"tenantId"`
	MetricKey    string    `gorm:"column:metric_key;type:varchar(32);not null" json:"metricKey"`
	Label        string    `gorm:"column:label;type:varchar(64)" json:"label"`
	Unit         string    `gorm:"column:unit;type:varchar(32)" json:"unit"`
	Scope        string    `gorm:"column:scope;type:varchar(16);not null;default:global" json:"scope"`
	DangerLow    *float64  `gorm:"column:danger_low" json:"dangerLow"`
	WarnLow      *float64  `gorm:"column:warn_low" json:"warnLow"`
	WarnHigh     *float64  `gorm:"column:warn_high" json:"warnHigh"`
	DangerHigh   *float64  `gorm:"column:danger_high" json:"dangerHigh"`
	Basis        string    `gorm:"column:basis;type:text" json:"basis"`
	Enabled      bool      `gorm:"column:enabled;not null;default:true" json:"enabled"`
	SortOrder    int       `gorm:"column:sort_order;not null;default:0" json:"sortOrder"`
	CreatedAt    time.Time `gorm:"column:created_at" json:"createdAt"`
	UpdatedAt    time.Time `gorm:"column:updated_at" json:"updatedAt"`
	LastModifyBy *int64    `gorm:"column:last_modify_by" json:"lastModifyBy"`
}

func (MonitoringThreshold) TableName() string { return "monitoring_threshold" }

// MonitoringVPStratum 静脉压分层阈值（通路×BF），每段一行。
type MonitoringVPStratum struct {
	ID           int64     `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	TenantID     int64     `gorm:"column:tenant_id;not null" json:"tenantId"`
	Access       string    `gorm:"column:access;type:varchar(8);not null" json:"access"`
	BFMin        float64   `gorm:"column:bf_min;not null" json:"bfMin"`
	BFMax        float64   `gorm:"column:bf_max;not null" json:"bfMax"`
	NormalLow    float64   `gorm:"column:normal_low;not null" json:"normalLow"`
	WarnHigh     float64   `gorm:"column:warn_high;not null" json:"warnHigh"`
	DangerHigh   float64   `gorm:"column:danger_high;not null" json:"dangerHigh"`
	Basis        string    `gorm:"column:basis;type:text" json:"basis"`
	Enabled      bool      `gorm:"column:enabled;not null;default:true" json:"enabled"`
	CreatedAt    time.Time `gorm:"column:created_at" json:"createdAt"`
	UpdatedAt    time.Time `gorm:"column:updated_at" json:"updatedAt"`
	LastModifyBy *int64    `gorm:"column:last_modify_by" json:"lastModifyBy"`
}

func (MonitoringVPStratum) TableName() string { return "monitoring_vp_stratum" }

// MonitoringSetting 标量配置 key-value。
type MonitoringSetting struct {
	ID           int64     `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	TenantID     int64     `gorm:"column:tenant_id;not null" json:"tenantId"`
	SettingKey   string    `gorm:"column:setting_key;type:varchar(64);not null" json:"settingKey"`
	ValueNum     *float64  `gorm:"column:value_num" json:"valueNum"`
	ValueText    *string   `gorm:"column:value_text" json:"valueText"`
	UpdatedAt    time.Time `gorm:"column:updated_at" json:"updatedAt"`
	LastModifyBy *int64    `gorm:"column:last_modify_by" json:"lastModifyBy"`
}

func (MonitoringSetting) TableName() string { return "monitoring_setting" }

// SettingKeyDialysateNaFactor 透析液钠电导率系数的 setting_key。
const SettingKeyDialysateNaFactor = "dialysateNaFactor"

// ScopeGlobal 阈值作用域：全局（本期唯一作用域，ward/patient 预留）。
const ScopeGlobal = "global"

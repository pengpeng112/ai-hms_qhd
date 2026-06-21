package models

import "time"

// DisinfectionCompliance 消毒监管伴生表：挂老库 Auxiliary_EquipmentDisinfection.Id，
// 补残留检测/结果/来源/文档/浓度（老表无此列，不 ALTER 老表）。
type DisinfectionCompliance struct {
	ID             string    `gorm:"column:id;type:varchar(36);primaryKey" json:"id"`
	TenantID       int64     `gorm:"column:tenant_id;index:idx_dc_tenant_device;not null" json:"tenantId"`
	DisinfectionID int64     `gorm:"column:disinfection_id;uniqueIndex;not null" json:"disinfectionId"`
	DeviceID       int64     `gorm:"column:device_id;index:idx_dc_tenant_device" json:"deviceId"`
	Concentration  string    `gorm:"column:concentration;type:varchar(32)" json:"concentration"`
	ResidualCheck  string    `gorm:"column:residual_check;type:varchar(8)" json:"residualCheck"`
	Result         string    `gorm:"column:result;type:varchar(8)" json:"result"`
	Source         string    `gorm:"column:source;type:varchar(12)" json:"source"`
	DocRef         string    `gorm:"column:doc_ref;type:varchar(256)" json:"docRef"`
	CreatedAt      time.Time `gorm:"column:created_at;autoCreateTime" json:"createdAt"`
	UpdatedAt      time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updatedAt"`
}

func (DisinfectionCompliance) TableName() string { return "disinfection_compliance" }

const (
	DisinfectTypeHeat     = "heat"
	DisinfectTypeTerminal = "terminal"
	DisinfectTypeDecalc   = "decalc"
	DisinfectTypeEnhanced = "enhanced"
)

const (
	DisinfectResultPass = "pass"
	DisinfectResultFail = "fail"
)

package models

import "time"

// DeviceStatus 设备状态
const (
	DeviceStatusNormal      = "normal"
	DeviceStatusWarning     = "warning"
	DeviceStatusAlarm       = "alarm"
	DeviceStatusOffline     = "offline"
	DeviceStatusMaintenance = "maintenance"
)

// Device 透析设备
type Device struct {
	ID             string    `gorm:"type:varchar(36);primaryKey" json:"id"`
	TenantId       int64     `gorm:"type:bigint;index" json:"tenantId"`
	Name           string    `gorm:"type:varchar(100);not null" json:"name"`         // 设备名称
	SerialNo       string    `gorm:"type:varchar(100)" json:"serialNo"`              // 序列号
	Model          string    `gorm:"type:varchar(100)" json:"model"`                 // 型号
	Manufacturer   string    `gorm:"type:varchar(100)" json:"manufacturer"`          // 厂商
	BedNumber      string    `gorm:"type:varchar(20);index" json:"bedNumber"`        // 对应床号
	WardId         *int64    `gorm:"type:bigint;index" json:"wardId"`                // 所属病区
	Status         string    `gorm:"type:varchar(20);default:offline" json:"status"` // 设备状态
	PurchaseDate   *time.Time `gorm:"type:date" json:"purchaseDate"`                 // 购入日期
	LastMaintained *time.Time `gorm:"type:date" json:"lastMaintained"`               // 最近维保日期
	Notes          string    `gorm:"type:text" json:"notes"`
	IsDisabled     bool      `gorm:"default:false" json:"isDisabled"`
	CreatorId      int64     `gorm:"type:bigint" json:"creatorId"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
}

// TableName 指定表名
func (Device) TableName() string {
	return "devices"
}

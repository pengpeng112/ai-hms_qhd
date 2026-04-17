// DEPRECATED: legacy new-db model, will be rewritten to map legacy hemodialysis DB in Phase 1~5.
package models

import (
	"time"

	modeltypes "github.com/elliotxin/ai-hms-backend/internal/models/types"
)

// Ward 病房/病区
type Ward struct {
	Id             int64     `gorm:"column:Id;type:bigint;primaryKey" json:"id"`
	TenantId       int64     `gorm:"column:TenantId;type:bigint;index" json:"tenantId"`
	Name           string    `gorm:"column:Name;type:varchar(256);not null" json:"name"`  // 病房名称
	WardType       string    `gorm:"column:PatientType;type:varchar(64)" json:"wardType"` // 病房类型（映射自 PatientType）
	Department     string    `gorm:"-" json:"department"`                                 // 老表无直接字段
	Floor          *int      `gorm:"-" json:"floor"`                                      // 老表无直接字段
	IsDisabled     bool      `gorm:"column:IsDisabled;default:false" json:"isDisabled"`
	Sort           int       `gorm:"column:Sort" json:"sort"`
	Notes          string    `gorm:"column:Note;type:varchar(512)" json:"notes"`
	CreatorId      int64     `gorm:"column:CreatorId;type:bigint" json:"creatorId"`
	CreateTime     time.Time `gorm:"column:CreateTime;autoCreateTime" json:"createTime"`
	LastModifyTime time.Time `gorm:"column:LastModifyTime;autoUpdateTime" json:"lastModifyTime"`

	// 关联
	Beds []Bed `gorm:"foreignKey:WardId" json:"beds,omitempty"`
}

// TableName 指定表名
func (Ward) TableName() string {
	return "Schedule_Ward"
}

// WardType 病房类型常量
const (
	WardTypeHD        = "HD"        // 血液透析
	WardTypeHDF       = "HDF"       // 血液透析滤过
	WardTypeIsolation = "Isolation" // 隔离病房
	WardTypeVIP       = "VIP"       // VIP病房
)

// Bed 床位管理
type Bed struct {
	Id             int64     `gorm:"column:Id;type:bigint;primaryKey" json:"id"`
	TenantId       int64     `gorm:"column:TenantId;type:bigint;index" json:"tenantId"`
	WardId         *int64    `gorm:"column:WardId;type:bigint;index" json:"wardId"` // 所属病房
	Name           string    `gorm:"column:Name;type:varchar(256);not null" json:"name"`
	BedType        string    `gorm:"-" json:"bedType"` // 老表无直接字段
	Status         string    `gorm:"-" json:"status"`  // 老表无直接字段
	IsDisabled     bool      `gorm:"column:IsDisabled;default:false" json:"isDisabled"`
	Sort           int       `gorm:"column:Sort" json:"sort"`
	Notes          string    `gorm:"column:Note;type:varchar(512)" json:"notes"`
	CreatorId      int64     `gorm:"column:CreatorId;type:bigint" json:"creatorId"`
	CreateTime     time.Time `gorm:"column:CreateTime;autoCreateTime" json:"createTime"`
	LastModifyTime time.Time `gorm:"column:LastModifyTime;autoUpdateTime" json:"lastModifyTime"`

	// 关联
	Ward *Ward `gorm:"foreignKey:WardId" json:"ward,omitempty"`
}

// TableName 指定表名
func (Bed) TableName() string {
	return "Schedule_Bed"
}

// BedType 床位类型常量
const (
	BedTypeRegular   = "Regular"   // 普通床
	BedTypeICU       = "ICU"       // ICU床位
	BedTypeVIP       = "VIP"       // VIP床位
	BedTypeIsolation = "Isolation" // 隔离床
)

// BedStatus 床位状态常量
const (
	BedStatusAvailable   = "available"   // 可用
	BedStatusOccupied    = "occupied"    // 占用中
	BedStatusReserved    = "reserved"    // 预留
	BedStatusMaintenance = "maintenance" // 维护中
)

// Shift 班次定义
type Shift struct {
	Id             int64     `gorm:"column:Id;type:bigint;primaryKey" json:"id"`
	TenantId       int64     `gorm:"column:TenantId;type:bigint;index" json:"tenantId"`
	Name           string    `gorm:"column:Name;type:varchar(256);not null" json:"name"`          // 班次名称：早班/中班/晚班
	StartTime      string    `gorm:"column:StartTime;type:varchar(32);not null" json:"startTime"` // 开始时间
	EndTime        string    `gorm:"column:EndTime;type:varchar(32);not null" json:"endTime"`     // 结束时间
	Type           string    `gorm:"column:Type;type:varchar(64)" json:"type"`                    // 班次类型
	IsDisabled     bool      `gorm:"column:IsDisabled;default:false" json:"isDisabled"`
	Sort           int       `gorm:"column:Sort" json:"sort"`
	Notes          string    `gorm:"column:Note;type:varchar(512)" json:"notes"`
	CreatorId      int64     `gorm:"column:CreatorId;type:bigint" json:"creatorId"`
	CreateTime     time.Time `gorm:"column:CreateTime;autoCreateTime" json:"createTime"`
	LastModifyTime time.Time `gorm:"column:LastModifyTime;autoUpdateTime" json:"lastModifyTime"`

	// 关联
	PatientShifts []PatientShift `gorm:"foreignKey:ShiftId" json:"patientShifts,omitempty"`
}

// TableName 指定表名
func (Shift) TableName() string {
	return "Schedule_Shift"
}

// ShiftType 班次类型常量
const (
	ShiftTypeMorning   = "Morning"   // 早班
	ShiftTypeAfternoon = "Afternoon" // 中班
	ShiftTypeNight     = "Night"     // 晚班
	ShiftTypeOvertime  = "Overtime"  // 加班
)

// PatientShift 患者排班
type PatientShift struct {
	Id             int64               `gorm:"column:Id;type:bigint;primaryKey" json:"id"`
	TenantId       int64               `gorm:"column:TenantId;type:bigint;index" json:"tenantId"`
	PatientId      modeltypes.LegacyID `gorm:"column:PatientId;type:bigint;not null;index" json:"patientId"`
	ScheduleDate   time.Time           `gorm:"column:TreatmentTime;type:timestamp;not null;index" json:"scheduleDate"` // 映射到老表 TreatmentTime
	ShiftId        int64               `gorm:"column:ShiftId;type:bigint;not null;index" json:"shiftId"`
	BedId          *int64              `gorm:"column:BedId;type:bigint;index" json:"bedId"`
	WardId         *int64              `gorm:"column:WardId;type:bigint;index" json:"wardId"`
	Status         int                 `gorm:"column:Status;type:int" json:"status"`
	IsDisabled     bool                `gorm:"-" json:"isDisabled"`
	Notes          string              `gorm:"-" json:"notes"`
	CreatorId      int64               `gorm:"column:CreatorId;type:bigint" json:"creatorId"`
	CreateTime     time.Time           `gorm:"column:CreateTime;autoCreateTime" json:"createTime"`
	LastModifyTime time.Time           `gorm:"column:LastModifyTime;autoUpdateTime" json:"lastModifyTime"`

	// 关联
	Patient *Patient `gorm:"foreignKey:PatientId" json:"patient,omitempty"`
	Shift   *Shift   `gorm:"foreignKey:ShiftId" json:"shift,omitempty"`
	Bed     *Bed     `gorm:"foreignKey:BedId" json:"bed,omitempty"`
	Ward    *Ward    `gorm:"foreignKey:WardId" json:"ward,omitempty"`
}

// TableName 指定表名
func (PatientShift) TableName() string {
	return "Schedule_PatientShift"
}

// PatientShiftStatus 患者排班状态常量
const (
	PatientShiftStatusPending    = 0 // 待执行
	PatientShiftStatusConfirmed  = 1 // 已确认
	PatientShiftStatusInProgress = 2 // 进行中
	PatientShiftStatusCompleted  = 3 // 已完成
	PatientShiftStatusCancelled  = 4 // 已取消
)

package models

import (
	"time"
)

// Ward 病房/病区
type Ward struct {
	Id           int64     `gorm:"type:bigint;primaryKey" json:"id"`
	TenantId     int64     `gorm:"type:bigint;index" json:"tenantId"`
	Name         string    `gorm:"type:varchar(128);not null" json:"name"`         // 病房名称
	WardType     string    `gorm:"type:varchar(64)" json:"wardType"`           // 病房类型
	Department   string    `gorm:"type:varchar(128)" json:"department"`         // 科室
	Floor        *int      `json:"floor"`                                      // 楼层
	IsDisabled   bool      `gorm:"default:false" json:"isDisabled"`
	Sort         int       `json:"sort"`
	Notes          string    `gorm:"type:text" json:"notes"`
	CreatorId      int64     `gorm:"type:bigint" json:"creatorId"`
	CreateTime     time.Time `json:"createTime"`
	LastModifyTime time.Time `json:"lastModifyTime"`

	// 关联
	Beds []Bed `gorm:"foreignKey:WardId" json:"beds,omitempty"`
}

// TableName 指定表名
func (Ward) TableName() string {
	return "wards"
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
	Id          int64     `gorm:"type:bigint;primaryKey" json:"id"`
	TenantId    int64     `gorm:"type:bigint;index" json:"tenantId"`
	WardId      *int64    `gorm:"type:bigint;index" json:"wardId"`           // 所属病房
	Name        string    `gorm:"type:varchar(64);not null" json:"name"`        // 床位号
	BedType     string    `gorm:"type:varchar(64)" json:"bedType"`            // 床位类型
	Status      string    `gorm:"type:varchar(20);default:available" json:"status"` // 状态
	IsDisabled  bool      `gorm:"default:false" json:"isDisabled"`
	Sort        int       `json:"sort"`
	Notes          string    `gorm:"type:text" json:"notes"`
	CreatorId      int64     `gorm:"type:bigint" json:"creatorId"`
	CreateTime     time.Time `json:"createTime"`
	LastModifyTime time.Time `json:"lastModifyTime"`

	// 关联
	Ward *Ward `gorm:"foreignKey:WardId" json:"ward,omitempty"`
}

// TableName 指定表名
func (Bed) TableName() string {
	return "beds"
}

// BedType 床位类型常量
const (
	BedTypeRegular  = "Regular"  // 普通床
	BedTypeICU      = "ICU"      // ICU床位
	BedTypeVIP      = "VIP"      // VIP床位
	BedTypeIsolation = "Isolation" // 隔离床
)

// BedStatus 床位状态常量
const (
	BedStatusAvailable = "available" // 可用
	BedStatusOccupied  = "occupied"  // 占用中
	BedStatusReserved  = "reserved"  // 预留
	BedStatusMaintenance = "maintenance" // 维护中
)

// Shift 班次定义
type Shift struct {
	Id          int64     `gorm:"type:bigint;primaryKey" json:"id"`
	TenantId    int64     `gorm:"type:bigint;index" json:"tenantId"`
	Name        string    `gorm:"type:varchar(64);not null" json:"name"`        // 班次名称：早班/中班/晚班
	StartTime  string    `gorm:"type:varchar(10);not null" json:"startTime"`  // 开始时间 HH:MM
	EndTime    string    `gorm:"type:varchar(10);not null" json:"endTime"`    // 结束时间 HH:MM
	Type        string    `gorm:"type:varchar(64)" json:"type"`               // 班次类型
	IsDisabled bool      `gorm:"default:false" json:"isDisabled"`
	Sort        int       `json:"sort"`
	Notes          string    `gorm:"type:text" json:"notes"`
	CreatorId      int64     `gorm:"type:bigint" json:"creatorId"`
	CreateTime     time.Time `json:"createTime"`
	LastModifyTime time.Time `json:"lastModifyTime"`

	// 关联
	PatientShifts []PatientShift `gorm:"foreignKey:ShiftId" json:"patientShifts,omitempty"`
}

// TableName 指定表名
func (Shift) TableName() string {
	return "shifts"
}

// ShiftType 班次类型常量
const (
	ShiftTypeMorning = "Morning" // 早班
	ShiftTypeAfternoon = "Afternoon" // 中班
	ShiftTypeNight = "Night"   // 晚班
	ShiftTypeOvertime = "Overtime" // 加班
)

// PatientShift 患者排班
type PatientShift struct {
	Id            int64      `gorm:"type:bigint;primaryKey" json:"id"`
	TenantId      int64      `gorm:"type:bigint;index" json:"tenantId"`
	PatientId     int64     `gorm:"type:bigint;not null;uniqueIndex:unique_patient_date" json:"patientId"`
	ScheduleDate  time.Time `gorm:"type:timestamp;not null;uniqueIndex:unique_patient_date" json:"scheduleDate"` // 排班日期
	ShiftId       int64     `gorm:"type:bigint;not null;index" json:"shiftId"`     // 班次ID
	BedId         *int64    `gorm:"type:bigint;index" json:"bedId"`            // 床位ID
	WardId        *int64    `gorm:"type:bigint;index" json:"wardId"`           // 病房ID
	Status        int       `gorm:"type:int;default:0" json:"status"`          // 排班状态
	IsDisabled    bool      `gorm:"default:false" json:"isDisabled"`
	Notes          string    `gorm:"type:text" json:"notes"`
	CreatorId      int64     `gorm:"type:bigint" json:"creatorId"`
	CreateTime     time.Time `json:"createTime"`
	LastModifyTime time.Time `json:"lastModifyTime"`

	// 关联
	Patient *Patient `gorm:"foreignKey:PatientId" json:"patient,omitempty"`
	Shift   *Shift   `gorm:"foreignKey:ShiftId" json:"shift,omitempty"`
	Bed     *Bed     `gorm:"foreignKey:BedId" json:"bed,omitempty"`
	Ward    *Ward    `gorm:"foreignKey:WardId" json:"ward,omitempty"`
}

// TableName 指定表名
func (PatientShift) TableName() string {
	return "patient_shifts"
}

// PatientShiftStatus 患者排班状态常量
const (
	PatientShiftStatusPending    = 0 // 待执行
	PatientShiftStatusConfirmed  = 1 // 已确认
	PatientShiftStatusInProgress = 2 // 进行中
	PatientShiftStatusCompleted  = 3 // 已完成
	PatientShiftStatusCancelled  = 4 // 已取消
)

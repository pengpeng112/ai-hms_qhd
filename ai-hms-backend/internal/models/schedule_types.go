package models

import "time"

type ScheduleWard struct {
	Id             int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	TenantId       int64     `gorm:"index:idx_ward_tenant;not null" json:"tenantId"`
	Name           string    `gorm:"size:256;not null" json:"name"`
	PatientType    string    `gorm:"column:PatientType;size:64" json:"patientType"`
	IsDisabled     bool      `gorm:"default:false" json:"isDisabled"`
	Sort           int       `json:"sort"`
	Note           string    `gorm:"size:512" json:"note"`
	CreatorId      int64     `gorm:"not null" json:"creatorId"`
	CreateTime     time.Time `gorm:"autoCreateTime" json:"createTime"`
	LastModifyTime time.Time `gorm:"autoUpdateTime" json:"lastModifyTime"`
}

func (ScheduleWard) TableName() string { return "Schedule_Ward" }

type ScheduleBed struct {
	Id             int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	TenantId       int64     `gorm:"index:idx_bed_tenant;not null" json:"tenantId"`
	WardId         *int64    `gorm:"index:idx_bed_ward" json:"wardId"`
	Name           string    `gorm:"size:256;not null" json:"name"`
	IsDisabled     bool      `gorm:"default:false" json:"isDisabled"`
	Sort           int       `json:"sort"`
	Note           string    `gorm:"size:512" json:"note"`
	CreatorId      int64     `gorm:"not null" json:"creatorId"`
	CreateTime     time.Time `gorm:"autoCreateTime" json:"createTime"`
	LastModifyTime time.Time `gorm:"autoUpdateTime" json:"lastModifyTime"`
}

func (ScheduleBed) TableName() string { return "Schedule_Bed" }

type ScheduleShift struct {
	Id             int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	TenantId       int64     `gorm:"index:idx_shift_tenant;not null" json:"tenantId"`
	Name           string    `gorm:"size:256;not null" json:"name"`
	StartTime      string    `gorm:"size:32;not null" json:"startTime"`
	EndTime        string    `gorm:"size:32;not null" json:"endTime"`
	Type           string    `gorm:"size:64" json:"type"`
	IsDisabled     bool      `gorm:"default:false" json:"isDisabled"`
	Sort           int       `json:"sort"`
	Note           string    `gorm:"size:512" json:"note"`
	CreatorId      int64     `gorm:"not null" json:"creatorId"`
	CreateTime     time.Time `gorm:"autoCreateTime" json:"createTime"`
	LastModifyTime time.Time `gorm:"autoUpdateTime" json:"lastModifyTime"`
}

func (ScheduleShift) TableName() string { return "Schedule_Shift" }

type SchedulePatientShift struct {
	Id              int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	TenantId        int64     `gorm:"index:idx_ps_tenant;not null" json:"tenantId"`
	PatientId       int64     `gorm:"type:bigint;not null;index:idx_ps_patient" json:"patientId"`
	ScheduleDate    time.Time `gorm:"column:TreatmentTime;not null;index:idx_ps_date" json:"scheduleDate"`
	ShiftId         int64     `gorm:"not null;index:idx_ps_shift" json:"shiftId"`
	BedId           *int64    `gorm:"index:idx_ps_bed" json:"bedId"`
	WardId          *int64    `gorm:"index:idx_ps_ward" json:"wardId"`
	PatientPlanId   *int64    `gorm:"index:idx_ps_plan" json:"patientPlanId"`
	ShiftTiming     *int      `json:"shiftTiming"`
	Status          int       `gorm:"not null;default:0" json:"status"`
	CreatorId       int64     `gorm:"not null" json:"creatorId"`
	CreateTime      time.Time `gorm:"autoCreateTime" json:"createTime"`
	LastModifyTime  time.Time `gorm:"autoUpdateTime" json:"lastModifyTime"`
}

func (SchedulePatientShift) TableName() string { return "Schedule_PatientShift" }


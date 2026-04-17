// DEPRECATED: legacy new-db model, will be rewritten to map legacy hemodialysis DB in Phase 1~5.
package models

import (
	"time"
)

// Hospitalization 住院信息
type Hospitalization struct {
	Id              int64      `gorm:"column:Id;type:bigint;primaryKey" json:"id"`
	TenantId        int64      `gorm:"column:TenantId;type:bigint;index" json:"tenantId"`
	PatientId       int64      `gorm:"column:PatientId;type:bigint;not null;index:idx_hospitalization_patient" json:"patientId"`
	CaseNo          string     `gorm:"column:CaseNo;type:varchar(64)" json:"caseNo"`
	HospNo          string     `gorm:"column:HospNo;type:varchar(64)" json:"hospNo"`
	BarCode         string     `gorm:"column:BarCode;type:varchar(64)" json:"barCode"`
	HospPatientType string     `gorm:"column:HospPatientType;type:varchar(64)" json:"hospPatientType"`
	HospReceiveDept string     `gorm:"column:HospReceiveDept;type:varchar(64)" json:"hospReceiveDept"`
	HospWard        string     `gorm:"column:HospWard;type:varchar(64)" json:"hospWard"`
	HospBed         string     `gorm:"column:HospBed;type:varchar(64)" json:"hospBed"`
	AttendDr        string     `gorm:"column:AttendDr;type:varchar(64)" json:"attendDr"`
	ReceptionDr     string     `gorm:"column:ReceptionDr;type:varchar(64)" json:"receptionDr"`
	Status          int        `gorm:"-" json:"status"`
	AdmissionDate   *time.Time `gorm:"-" json:"admissionDate"`
	DischargeDate   *time.Time `gorm:"-" json:"dischargeDate"`
	Notes           string     `gorm:"-" json:"notes"`
	CreatorId       int64      `gorm:"column:CreatorId;type:bigint" json:"creatorId"`
	CreateTime      time.Time  `gorm:"column:CreateTime;autoCreateTime" json:"createTime"`
	LastModifyTime  time.Time  `gorm:"column:LastModifyTime;autoUpdateTime" json:"lastModifyTime"`

	// 关联
	Patient *Patient `gorm:"foreignKey:PatientId" json:"patient,omitempty"`
}

// TableName 指定表名
func (Hospitalization) TableName() string {
	return "Register_Hospitalization"
}

// HospitalizationStatus 住院状态常量
const (
	HospitalizationStatusInPatient  = 1 // 在院
	HospitalizationStatusDischarged = 0 // 出院
)

// HospPatientType 住院患者类型常量
const (
	HospPatientTypeOutpatient = "门诊" // 门诊透析
	HospPatientTypeInpatient  = "住院" // 住院透析
	HospPatientTypeEmergency  = "急诊" // 急诊
)

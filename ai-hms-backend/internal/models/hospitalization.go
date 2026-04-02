package models

import (
	"time"
)

// Hospitalization 住院信息
type Hospitalization struct {
	Id              int64     `gorm:"type:bigint;primaryKey" json:"id"`
	TenantId        int64     `gorm:"type:bigint;index" json:"tenantId"`
	PatientId       int64     `gorm:"type:bigint;not null;index:idx_patient_id" json:"patientId"`
	CaseNo          string    `gorm:"type:varchar(64)" json:"caseNo"`           // 病案号
	HospNo          string    `gorm:"type:varchar(64)" json:"hospNo"`           // 住院号
	BarCode         string    `gorm:"type:varchar(64)" json:"barCode"`          // 条码
	HospPatientType string    `gorm:"type:varchar(64)" json:"hospPatientType"`  // 住院患者类型
	HospReceiveDept string    `gorm:"type:varchar(64)" json:"hospReceiveDept"` // 接收科室
	HospWard        string    `gorm:"type:varchar(64)" json:"hospWard"`        // 病房
	HospBed         string    `gorm:"type:varchar(64)" json:"hospBed"`         // 床位
	AttendDr        string    `gorm:"type:varchar(64)" json:"attendDr"`        // 主治医生
	ReceptionDr     string    `gorm:"type:varchar(64)" json:"receptionDr"`     // 接诊医生
	Status          int       `gorm:"type:int;default:1" json:"status"`        // 状态：1-在院，0-出院
	AdmissionDate   *time.Time `json:"admissionDate"`                        // 入院日期
	DischargeDate   *time.Time `json:"dischargeDate"`                        // 出院日期
	Notes          string    `gorm:"type:text" json:"notes"`                // 备注
	CreatorId      int64     `gorm:"type:bigint" json:"creatorId"`
	CreateTime     time.Time `json:"createTime"`
	LastModifyTime time.Time `json:"lastModifyTime"`

	// 关联
	Patient *Patient `gorm:"foreignKey:PatientId" json:"patient,omitempty"`
}

// TableName 指定表名
func (Hospitalization) TableName() string {
	return "hospitalizations"
}

// HospitalizationStatus 住院状态常量
const (
	HospitalizationStatusInPatient = 1 // 在院
	HospitalizationStatusDischarged = 0 // 出院
)

// HospPatientType 住院患者类型常量
const (
	HospPatientTypeOutpatient = "门诊" // 门诊透析
	HospPatientTypeInpatient  = "住院" // 住院透析
	HospPatientTypeEmergency  = "急诊" // 急诊
)

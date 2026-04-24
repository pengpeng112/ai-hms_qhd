// DEPRECATED: legacy new-db model, will be rewritten to map legacy hemodialysis DB in Phase 1~5.
package models

import (
	"time"

	modeltypes "github.com/elliotxin/ai-hms-backend/internal/models/types"
)

// PatientBasicInfo 患者基本信息档案扩展表
type PatientBasicInfo struct {
	ID        string              `gorm:"type:varchar(36);primaryKey" json:"id"`
	PatientID modeltypes.LegacyID `gorm:"type:bigint;not null;uniqueIndex" json:"patientId"`

	// 身份核心信息
	Pinyin    *string    `gorm:"type:varchar(100)" json:"pinyin"`            // 姓名拼音
	Birthday  *time.Time `json:"birthday"`                                   // 出生日期
	Ethnicity *string    `gorm:"type:varchar(20)" json:"ethnicity"`          // 民族
	IDType    string     `gorm:"type:varchar(20);default:身份证" json:"idType"` // 身份证件类型
	IDNumber  *string    `gorm:"type:varchar(50)" json:"idNumber"`           // 身份证号

	// 医疗登记信息
	VisitCategory         *string    `gorm:"type:varchar(20)" json:"visitCategory"`          // 就诊类别
	AdmissionNo           *string    `gorm:"type:varchar(50)" json:"admissionNo"`            // 住院号
	VisitNo               *string    `gorm:"type:varchar(50)" json:"visitNo"`                // 就诊号
	MedicalRecordNo       *string    `gorm:"type:varchar(50)" json:"medicalRecordNo"`        // 病历号
	InsuranceNo           *string    `gorm:"type:varchar(50)" json:"insuranceNo"`            // 医保号
	HdisPatientID         *int       `gorm:"uniqueIndex" json:"hdisPatientId,omitempty"`     // HDIS/LIS 患者数字 ID
	DialysisNo            *string    `gorm:"type:varchar(50)" json:"dialysisNo"`             // 透析号
	NurseName             *string    `gorm:"type:varchar(50)" json:"nurseName"`              // 责任护士
	FirstDialysisDate     *time.Time `json:"firstDialysisDate"`                              // 首次透析日期
	FirstHospitalDate     *time.Time `json:"firstHospitalDate"`                              // 首次在本院透析日期
	FirstDialysisHospital *string    `gorm:"type:varchar(100)" json:"firstDialysisHospital"` // 首次透析医院

	// 生命体征与社会信息
	Height         *string `gorm:"type:varchar(10)" json:"height"`         // 身高 (cm)
	ABOBloodType   *string `gorm:"type:varchar(10)" json:"aboBloodType"`   // ABO血型
	RhBloodType    *string `gorm:"type:varchar(10)" json:"rhBloodType"`    // Rh血型
	EducationLevel *string `gorm:"type:varchar(20)" json:"educationLevel"` // 文化程度
	Occupation     *string `gorm:"type:varchar(50)" json:"occupation"`     // 职业
	MaritalStatus  *string `gorm:"type:varchar(20)" json:"maritalStatus"`  // 婚姻状况
	Workplace      *string `gorm:"type:varchar(100)" json:"workplace"`     // 工作单位

	// 联系信息
	Phone        *string `gorm:"type:varchar(20)" json:"phone"`        // 手机号码
	Wechat       *string `gorm:"type:varchar(50)" json:"wechat"`       // 微信号
	Landline     *string `gorm:"type:varchar(20)" json:"landline"`     // 固定电话
	Address      *string `gorm:"type:text" json:"address"`             // 地址
	District     *string `gorm:"type:varchar(100)" json:"district"`    // 区域
	ContactName  *string `gorm:"type:varchar(50)" json:"contactName"`  // 紧急联系人
	ContactPhone *string `gorm:"type:varchar(20)" json:"contactPhone"` // 紧急联系电话

	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`

	// 关联
	Patient *Patient `gorm:"foreignKey:PatientID" json:"patient,omitempty"`
}

// TableName 指定表名
func (PatientBasicInfo) TableName() string {
	return "patient_basic_infos"
}

// IDType 身份证件类型常量
const (
	IDTypeIDCard   = "身份证"
	IDTypePassport = "护照"
	IDTypeOther    = "其他"
)

// VisitCategory 就诊类别常量
const (
	VisitCategoryOutpatient = "门诊"
	VisitCategoryInpatient  = "住院"
	VisitCategoryEmergency  = "急诊"
)

// ABOBloodType ABO血型常量
const (
	BloodTypeA  = "A"
	BloodTypeB  = "B"
	BloodTypeAB = "AB"
	BloodTypeO  = "O"
)

// RhBloodType Rh血型常量
const (
	RhPositive = "Rh+"
	RhNegative = "Rh-"
)

// EducationLevel 文化程度常量
const (
	EducationPrimary  = "小学"
	EducationMiddle   = "初中"
	EducationHigh     = "高中"
	EducationCollege  = "大专"
	EducationBachelor = "本科"
	EducationMaster   = "硕士"
	EducationDoctor   = "博士"
	EducationOther    = "其他"
)

// MaritalStatus 婚姻状况常量
const (
	MaritalSingle   = "未婚"
	MaritalMarried  = "已婚"
	MaritalDivorced = "离异"
	MaritalWidowed  = "丧偶"
)

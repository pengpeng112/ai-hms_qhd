// DEPRECATED: legacy new-db model, will be rewritten to map legacy hemodialysis DB in Phase 1~5.
package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	modeltypes "github.com/elliotxin/ai-hms-backend/internal/models/types"
)

// StringSlice JSON 数组类型，用于存储字符串数组到数据库
type StringSlice []string

// Value 实现 driver.Valuer 接口
func (s StringSlice) Value() (driver.Value, error) {
	if s == nil {
		return "[]", nil
	}
	return json.Marshal(s)
}

// Scan 实现 sql.Scanner 接口
func (s *StringSlice) Scan(value interface{}) error {
	if value == nil {
		*s = []string{}
		return nil
	}

	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return errors.New("unsupported type for StringSlice")
	}

	return json.Unmarshal(bytes, s)
}

// Patient 患者基本信息 — 映射老血透库 Register_PatientInfomation
type Patient struct {
	ID                 modeltypes.LegacyID `gorm:"column:Id;primaryKey" json:"id"`
	TenantID           int64               `gorm:"column:TenantId" json:"-"`
	Name               string              `gorm:"column:Name" json:"name"`
	Gender             string              `gorm:"column:Gender" json:"gender"`
	BirthDate          *time.Time          `gorm:"column:BirthDate" json:"birthDate"`
	PatientType        string              `gorm:"column:PatientType" json:"patientType"`
	InsuranceType      string              `gorm:"column:ExpenseType" json:"insuranceType"`
	Status             string              `gorm:"column:TreatmentStatus" json:"status"`
	DialysisNo         string              `gorm:"column:DialysisNo" json:"dialysisNo"`
	PhoneNo            string              `gorm:"column:PhoneNo" json:"phoneNo"`
	ImageBase64String  string              `gorm:"column:ImageBase64String" json:"-"`
	DoctorID      *string    `gorm:"-" json:"doctorId"` // 老库为 bigint，由服务层从 ResponsibilityDrId 转换
	AdmissionDate *time.Time `gorm:"column:FirstDialysisDate" json:"admissionDate"`
	CreatedAt     time.Time  `gorm:"column:CreateTime" json:"createdAt"`
	UpdatedAt     time.Time  `gorm:"column:LastModifyTime" json:"updatedAt"`
	Diagnosis          string              `gorm:"-" json:"diagnosis"`  // 从 Register_Diagnosis.DiagnosisDesc 获取，由服务层填充
	DryWeight          float64             `gorm:"-" json:"dryWeight"`  // 从 Plan_PatientPlan.DryWeight 获取，由服务层填充

	// 以下字段老库无直接对应列，由服务层计算/联查后填充
	Age           int        `gorm:"-" json:"age"`
	BedNumber     string     `gorm:"-" json:"bedNumber"`
	RiskLevel     string     `gorm:"-" json:"riskLevel"`
	DoctorName    string     `gorm:"-" json:"doctorName"`
	DefaultMode   string     `gorm:"-" json:"defaultMode"`
	DischargeDate *time.Time `gorm:"-" json:"dischargeDate"`

	// 关联（保留接口兼容）
	VascularAccesses []VascularAccess `gorm:"foreignKey:PatientID" json:"vascularAccesses,omitempty"`
	MedicalHistory   *MedicalHistory  `gorm:"foreignKey:PatientID" json:"medicalHistory,omitempty"`
	TreatmentPlan    *TreatmentPlan   `gorm:"foreignKey:PatientID" json:"treatmentPlan,omitempty"`
}

// TableName 指定表名 — 老血透库
func (Patient) TableName() string {
	return "Register_PatientInfomation"
}

// Gender 常量
const (
	GenderMale   = "男"
	GenderFemale = "女"
)

// RiskLevel 常量
const (
	RiskLevelHigh   = "高危"
	RiskLevelMedium = "中危"
	RiskLevelLow    = "低危"
)

// PatientStatus 常量
const (
	PatientStatusActive     = "active"
	PatientStatusInactive   = "inactive"
	PatientStatusDischarged = "discharged"
)

// PatientType 常量
const (
	PatientTypeOutpatient = "门诊"
	PatientTypeInpatient  = "住院"
)

// VascularAccess 血管通路 — 映射老血透库 Register_VascularAccess
type VascularAccess struct {
	ID                modeltypes.LegacyID `gorm:"column:Id;primaryKey" json:"id"`
	TenantID          int64               `gorm:"column:TenantId" json:"-"`
	PatientID         modeltypes.LegacyID `gorm:"column:PatientId;not null;index" json:"patientId"`
	AccessType        string              `gorm:"column:AccessType" json:"accessType"`
	Site              string              `gorm:"column:AccessPosition" json:"site"`
	Artery            string              `gorm:"column:Artery" json:"artery"`
	Vein              string              `gorm:"column:Venous" json:"vein"`
	Side              string              `gorm:"column:LeftAndRight" json:"side"`
	Hospital          string              `gorm:"column:OperationHospital" json:"hospital"`
	Surgeon           string              `gorm:"column:OperationDr" json:"surgeon"`
	SurgeryDate       *time.Time          `gorm:"column:OperationTime" json:"surgeryDate"`
	FirstUseDate      *time.Time          `gorm:"column:FirstUseTime" json:"firstUseDate"`
	AccessNumber      int64               `gorm:"column:AccessCount" json:"accessNumber"`
	InterventionCount int64               `gorm:"column:InterveneCount" json:"interventionCount"`
	InterventionDate  *time.Time          `gorm:"column:InterveneTime" json:"interventionDate"`
	CatheterMethod    string              `gorm:"column:CatheterizeMethod" json:"catheterMethod"`
	CatheterDepth     float64             `gorm:"column:CatheterDepth" json:"catheterDepth"`
	VPuncturePosition string              `gorm:"column:VSidePointCount" json:"vPuncturePosition"`
	APuncturePosition string              `gorm:"column:ASidePointCount" json:"aPuncturePosition"`
	Notes             string              `gorm:"column:Note" json:"notes"`
	PictureID         *modeltypes.LegacyID `gorm:"column:PictureIds" json:"-"`
	IsDefault         bool                `gorm:"column:IsDefault" json:"isDefault"`
	IsDisabled        bool                `gorm:"column:IsDisabled" json:"isDisabled"`
	CreatorID         int64               `gorm:"column:CreatorId" json:"-"`
	CreatedAt      time.Time `gorm:"column:CreateTime" json:"createdAt"`
	UpdatedAt      time.Time `gorm:"column:LastModifyTime" json:"updatedAt"`
}

// TableName 指定表名 — 老血透库
func (VascularAccess) TableName() string {
	return "Register_VascularAccess"
}

// VascularAccessImage 血管通路图片 — 映射老血透库 Register_VascularAccessImage
type VascularAccessImage struct {
	ID                modeltypes.LegacyID `gorm:"column:Id;primaryKey" json:"id"`
	TenantID          int64               `gorm:"column:TenantId" json:"-"`
	VascularAccessID  modeltypes.LegacyID `gorm:"column:VascularAccessId;not null;index" json:"vascularAccessId"`
	ImageName         string              `gorm:"column:ImageName" json:"imageName"`
	ImageBase64String string              `gorm:"column:ImageBase64String" json:"imageBase64String"`
	Note              string              `gorm:"column:Note" json:"note"`
	Sort              int                 `gorm:"column:Sort" json:"sort"`
	CreatorID         int64               `gorm:"column:CreatorId" json:"-"`
	CreatedAt         time.Time           `gorm:"column:CreateTime" json:"createdAt"`
	UpdatedAt         time.Time           `gorm:"column:LastModifyTime" json:"updatedAt"`
}

// TableName 指定表名 — 老血透库
func (VascularAccessImage) TableName() string {
	return "Register_VascularAccessImage"
}

// VascularAccessIntervention 血管通路干预 — 映射老血透库 Register_VascularAccessChange
type VascularAccessIntervention struct {
	ID               modeltypes.LegacyID `gorm:"column:Id;primaryKey" json:"id"`
	TenantID         int64               `gorm:"column:TenantId" json:"-"`
	PatientID        modeltypes.LegacyID `gorm:"column:PatientId;not null;index" json:"patientId"`
	VascularAccessID modeltypes.LegacyID `gorm:"column:VascularAccessId;not null;index" json:"vascularAccessId"`
	UsageDays        int                 `gorm:"column:UseDuration" json:"usageDays"`
	SurgeryType      string              `gorm:"column:ChangeDesc" json:"surgeryType"`
	InterventionReason string            `gorm:"column:ChangeReason" json:"interventionReason"`
	AvgBloodFlow     float64             `gorm:"column:AvgBF" json:"avgBloodFlow"`
	InterventionDate time.Time           `gorm:"column:ChangeTime" json:"interventionDate"`
	Description      string              `gorm:"-" json:"description"`
	CreatedAt        time.Time           `gorm:"column:CreateTime" json:"createdAt"`
	UpdatedAt        time.Time           `gorm:"column:LastModifyTime" json:"updatedAt"`

	// 老库无对应列，保留兼容
	AccessType string `gorm:"-" json:"accessType"`
	Doctor     string `gorm:"-" json:"doctor"`

	// 关联
	VascularAccess *VascularAccess `gorm:"foreignKey:VascularAccessID" json:"vascularAccess,omitempty"`
}

// TableName 指定表名 — 老血透库
func (VascularAccessIntervention) TableName() string {
	return "Register_VascularAccessChange"
}

// VascularAccessType 常量
const (
	VascularAccessAVF = "自体动静脉内瘘AVF"
	VascularAccessAVG = "移植物动静脉内瘘AVG"
	VascularAccessTCC = "带隧道和涤纶套的透析导管TCC"
	VascularAccessNCC = "无隧道和涤纶套的透析导管NCC"
)

// VascularAccessStatus 常量
const (
	VascularAccessStatusNormal    = "正常"
	VascularAccessStatusThrombus  = "血栓"
	VascularAccessStatusStenosis  = "狭窄"
	VascularAccessStatusInfection = "感染"
)

// MedicalHistory 临床病史档案（一个患者一条记录）
type MedicalHistory struct {
	ID        modeltypes.LegacyID `gorm:"type:bigint;primaryKey" json:"id"`
	PatientID modeltypes.LegacyID `gorm:"type:bigint;not null;uniqueIndex" json:"patientId"`

	// 基础临床病史
	CurrentIllness     string `gorm:"type:text" json:"currentIllness"`     // 现病史
	PastHistory        string `gorm:"type:text" json:"pastHistory"`        // 既往史
	TransfusionHistory string `gorm:"type:text" json:"transfusionHistory"` // 输血史
	MaritalHistory     string `gorm:"type:text" json:"maritalHistory"`     // 婚育史
	FamilyHistory      string `gorm:"type:text" json:"familyHistory"`      // 家族史
	DiseaseDiagnosis   string `gorm:"type:text" json:"diseaseDiagnosis"`   // 疾病诊断

	// 专科记录
	PrimaryDiseaseName      string `gorm:"type:varchar(255)" json:"primaryDiseaseName"`     // 原发病名称
	PrimaryDiseaseContent   string `gorm:"type:text" json:"primaryDiseaseContent"`          // 原发病详情
	PrimaryDiseaseType      string `gorm:"type:varchar(255)" json:"primaryDiseaseType"`     // 原发病分类
	PrimaryDiseaseCheckTime string `gorm:"type:varchar(32)" json:"primaryDiseaseCheckTime"` // 原发病检查时间
	PrimaryDiseaseCheckDoc  string `gorm:"type:varchar(100)" json:"primaryDiseaseCheckDoc"` // 原发病检查医生
	PathologyName           string `gorm:"type:varchar(255)" json:"pathologyName"`          // 病理诊断名称
	PathologyContent        string `gorm:"type:text" json:"pathologyContent"`               // 病理诊断详情
	PathologyType           string `gorm:"type:varchar(255)" json:"pathologyType"`          // 病理诊断分类
	PathologyCheckTime      string `gorm:"type:varchar(32)" json:"pathologyCheckTime"`      // 病理检查时间
	PathologyCheckDoc       string `gorm:"type:varchar(100)" json:"pathologyCheckDoc"`      // 病理检查医生
	AllergenName            string `gorm:"type:varchar(255)" json:"allergenName"`           // 过敏信息名称
	AllergenContent         string `gorm:"type:text" json:"allergenContent"`                // 过敏信息详情
	AllergenType            string `gorm:"type:varchar(255)" json:"allergenType"`           // 过敏原分类
	AllergenCheckTime       string `gorm:"type:varchar(32)" json:"allergenCheckTime"`       // 过敏检查时间
	AllergenCheckDoc        string `gorm:"type:varchar(100)" json:"allergenCheckDoc"`       // 过敏检查医生
	TumorHistoryName        string `gorm:"type:varchar(255)" json:"tumorHistoryName"`       // 肿瘤病史名称
	TumorHistoryContent     string `gorm:"type:text" json:"tumorHistoryContent"`            // 肿瘤病史详情
	TumorHistoryType        string `gorm:"type:varchar(255)" json:"tumorHistoryType"`       // 肿瘤分类
	TumorHistoryCheckTime   string `gorm:"type:varchar(32)" json:"tumorHistoryCheckTime"`   // 肿瘤检查时间
	TumorHistoryCheckDoc    string `gorm:"type:varchar(100)" json:"tumorHistoryCheckDoc"`   // 肿瘤检查医生
	ComplicationName        string `gorm:"type:varchar(255)" json:"complicationName"`       // 并发症名称
	ComplicationContent     string `gorm:"type:text" json:"complicationContent"`            // 并发症详情
	ComplicationType        string `gorm:"type:varchar(255)" json:"complicationType"`       // 并发症分类
	ComplicationCheckTime   string `gorm:"type:varchar(32)" json:"complicationCheckTime"`   // 并发症检查时间
	ComplicationCheckDoc    string `gorm:"type:varchar(100)" json:"complicationCheckDoc"`   // 并发症检查医生

	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// TableName 指定表名
func (MedicalHistory) TableName() string {
	return "medical_histories"
}

// OutcomeRecord 治疗转归 — 映射老血透库 Register_OutCome
type OutcomeRecord struct {
	ID             modeltypes.LegacyID `gorm:"column:Id;primaryKey" json:"id"`
	TenantID       int64               `gorm:"column:TenantId" json:"-"`
	PatientID      modeltypes.LegacyID `gorm:"column:PatientId;not null;index" json:"patientId"`
	Type           string              `gorm:"column:Type" json:"type"`
	Reason         string              `gorm:"column:Reason" json:"reason"`
	Time           time.Time           `gorm:"column:OutComeTime" json:"time"`
	Remarks        string              `gorm:"column:Note" json:"remarks"`
	CreatedAt  time.Time `gorm:"column:CreateTime" json:"createdAt"`
	UpdatedAt  time.Time `gorm:"column:LastModifyTime" json:"updatedAt"`

	// 老库无以下字段，保留供服务层使用（不写入 DB）
	Registrar        string    `gorm:"-" json:"registrar"`
	RegistrationTime time.Time `gorm:"-" json:"registrationTime"`
	IsDoorRule       bool      `gorm:"-" json:"isDoorRule"`
}

// TableName 指定表名 — 老血透库
func (OutcomeRecord) TableName() string {
	return "Register_OutCome"
}

// OutcomeType 转归类型常量
const (
	OutcomeTypeIn  = "转入"
	OutcomeTypeOut = "转出"
)

// InfectionInfo 传染病筛查 — 映射老血透库 Register_Infection
// 注意：老库 InfectionDesc 字段存储感染描述文本，无独立 HbsAg/HcvAb 等字段
type InfectionInfo struct {
	ID             modeltypes.LegacyID `gorm:"column:Id;primaryKey" json:"id"`
	TenantID       int64               `gorm:"column:TenantId" json:"-"`
	PatientID      modeltypes.LegacyID `gorm:"column:PatientId;not null;uniqueIndex" json:"patientId"`
	InfectionDesc  string              `gorm:"column:InfectionDesc" json:"infectionDesc"`
	OtherDesc      string              `gorm:"column:OtherDesc" json:"otherDesc"`
	Note           string              `gorm:"column:Note" json:"note"`
	CreateTime     time.Time           `gorm:"column:CreateTime" json:"createdAt"`
	LastModifyTime time.Time           `gorm:"column:LastModifyTime" json:"updatedAt"`

	// 以下字段由 InfectionDesc 解析或来自其他表，暂以空值兼容前端
	HbsAg      string    `gorm:"-" json:"hbsag"`
	HcvAb      string    `gorm:"-" json:"hcvab"`
	HivAb      string    `gorm:"-" json:"hivab"`
	TpaB       string    `gorm:"-" json:"tpab"`
	UpdateDate time.Time `gorm:"-" json:"updateDate"`
}

// TableName 指定表名 — 老血透库
func (InfectionInfo) TableName() string {
	return "Register_Infection"
}

// InfectionStatus 常量
const (
	InfectionNegative = "阴性"
	InfectionPositive = "阳性"
)

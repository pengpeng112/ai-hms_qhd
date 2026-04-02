package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
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

// Patient 患者基本信息
type Patient struct {
	ID             string         `gorm:"type:varchar(36);primaryKey" json:"id"`
	Name           string         `gorm:"type:varchar(50);not null" json:"name"`
	Age            int            `gorm:"not null" json:"age"`
	Gender         string         `gorm:"type:varchar(10);not null" json:"gender"` // M, F (ISO 5218)
	BedNumber      string         `gorm:"type:varchar(20)" json:"bedNumber"`
	Diagnosis      string         `gorm:"type:text" json:"diagnosis"`
	RiskLevel      string         `gorm:"type:varchar(20);default:低危" json:"riskLevel"` // 高危, 中危, 低危
	Status         string         `gorm:"type:varchar(20);default:active" json:"status"` // active, inactive, discharged
	PatientType    string         `gorm:"type:varchar(50)" json:"patientType"`           // 门诊, 住院
	InsuranceType  string         `gorm:"type:varchar(50)" json:"insuranceType"`
	DryWeight      float64        `gorm:"type:decimal(5,2)" json:"dryWeight"`
	DefaultMode    string         `gorm:"type:varchar(50)" json:"defaultMode"`
	DoctorID      *string    `gorm:"type:varchar(36)" json:"doctorId"`
	DoctorName    string     `gorm:"type:varchar(50)" json:"doctorName"`
	AdmissionDate *time.Time `json:"admissionDate"`
	DischargeDate *time.Time `json:"dischargeDate"`
	CreatedAt     time.Time  `json:"createdAt"`
	UpdatedAt     time.Time  `json:"updatedAt"`

	// 关联
	VascularAccesses []VascularAccess `gorm:"foreignKey:PatientID" json:"vascularAccesses,omitempty"`
	MedicalHistory   *MedicalHistory  `gorm:"foreignKey:PatientID" json:"medicalHistory,omitempty"`
	TreatmentPlan    *TreatmentPlan   `gorm:"foreignKey:PatientID" json:"treatmentPlan,omitempty"`
}

// TableName 指定表名
func (Patient) TableName() string {
	return "patients"
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

// VascularAccess 血管通路（一个患者多条记录）
type VascularAccess struct {
	ID                string      `gorm:"type:varchar(36);primaryKey" json:"id"`
	PatientID         string      `gorm:"type:varchar(36);not null;index" json:"patientId"`
	AccessType        string      `gorm:"type:varchar(50);not null" json:"accessType"`         // AVF/AVG/TCC/NCC
	Site              string      `gorm:"type:varchar(100)" json:"site"`                       // 通路部位
	Artery            StringSlice `gorm:"type:text" json:"artery"`                             // 动脉（JSON 数组）
	Vein              StringSlice `gorm:"type:text" json:"vein"`                               // 静脉（JSON 数组）
	Side              string      `gorm:"type:varchar(10)" json:"side"`                        // L/R
	Hospital          string      `gorm:"type:varchar(200)" json:"hospital"`                   // 手术医院
	Surgeon           string      `gorm:"type:varchar(100)" json:"surgeon"`                    // 手术医生
	SurgeryDate       *time.Time  `json:"surgeryDate"`                                         // 手术时间
	FirstUseDate      *time.Time  `json:"firstUseDate"`                                        // 首次使用时间
	AccessNumber      int         `gorm:"default:1" json:"accessNumber"`                       // 第几次血管通路
	InterventionCount int         `gorm:"default:0" json:"interventionCount"`                  // 干预次数
	InterventionDate  *time.Time  `json:"interventionDate"`                                    // 干预日期
	CatheterMethod    *string     `gorm:"type:varchar(50)" json:"catheterMethod"`              // 置管方法（导管）
	CatheterDepth     *string     `gorm:"type:varchar(20)" json:"catheterDepth"`               // 导管深度
	VPuncturePosition StringSlice `gorm:"type:text" json:"vPuncturePosition"`                  // V侧穿刺位置
	APuncturePosition StringSlice `gorm:"type:text" json:"aPuncturePosition"`                  // A侧穿刺位置
	Notes             string      `gorm:"type:text" json:"notes"`                              // 备注
	Images            StringSlice `gorm:"type:text" json:"images"`                             // 图片URLs（JSON 数组）
	IsDefault         bool                       `gorm:"default:false" json:"isDefault"`                      // 是否默认
	IsDisabled        bool                       `gorm:"default:false" json:"isDisabled"`                     // 是否禁用
	CreatedAt         time.Time                  `json:"createdAt"`
	UpdatedAt         time.Time                  `json:"updatedAt"`
}

// TableName 指定表名
func (VascularAccess) TableName() string {
	return "vascular_accesses"
}

// VascularAccessIntervention 血管通路干预记录（一条血管通路有多条干预记录）
type VascularAccessIntervention struct {
	ID                 string     `gorm:"type:varchar(36);primaryKey" json:"id"`
	VascularAccessID   string     `gorm:"type:varchar(36);not null;index" json:"vascularAccessId"` // 关联的血管通路ID
	PatientID          string     `gorm:"type:varchar(36);not null;index" json:"patientId"`         // 冗余患者ID便于查询
	AccessType         string     `gorm:"type:varchar(50)" json:"accessType"`                        // 通路类型（冗余方便显示）
	AvgBloodFlow       int        `gorm:"default:0" json:"avgBloodFlow"`                             // 平均血流量
	UsageDays          int        `gorm:"default:0" json:"usageDays"`                                // 使用时间（天）
	SurgeryType        string     `gorm:"type:varchar(50);not null" json:"surgeryType"`             // 手术类型（必填）
	InterventionReason string     `gorm:"type:text;not null" json:"interventionReason"`             // 干预原因（必填）
	Doctor             string     `gorm:"type:varchar(50)" json:"doctor"`                            // 干预医生
	InterventionDate   time.Time  `gorm:"not null" json:"interventionDate"`                          // 干预时间
	Description        string     `gorm:"type:text" json:"description"`                              // 干预描述
	CreatedAt          time.Time  `json:"createdAt"`
	UpdatedAt          time.Time  `json:"updatedAt"`

	// 关联
	VascularAccess *VascularAccess `gorm:"foreignKey:VascularAccessID" json:"vascularAccess,omitempty"`
}

// TableName 指定表名
func (VascularAccessIntervention) TableName() string {
	return "vascular_access_interventions"
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
	VascularAccessStatusNormal   = "正常"
	VascularAccessStatusThrombus = "血栓"
	VascularAccessStatusStenosis = "狭窄"
	VascularAccessStatusInfection = "感染"
)

// MedicalHistory 临床病史档案（一个患者一条记录）
type MedicalHistory struct {
	ID        string `gorm:"type:varchar(36);primaryKey" json:"id"`
	PatientID string `gorm:"type:varchar(36);not null;uniqueIndex" json:"patientId"`

	// 基础临床病史
	CurrentIllness     string `gorm:"type:text" json:"currentIllness"`     // 现病史
	PastHistory        string `gorm:"type:text" json:"pastHistory"`        // 既往史
	TransfusionHistory string `gorm:"type:text" json:"transfusionHistory"` // 输血史
	MaritalHistory     string `gorm:"type:text" json:"maritalHistory"`     // 婚育史
	FamilyHistory      string `gorm:"type:text" json:"familyHistory"`      // 家族史
	DiseaseDiagnosis   string `gorm:"type:text" json:"diseaseDiagnosis"`   // 疾病诊断

	// 专科记录
	PrimaryDiseaseName      string `gorm:"type:varchar(255)" json:"primaryDiseaseName"`      // 原发病名称
	PrimaryDiseaseContent   string `gorm:"type:text" json:"primaryDiseaseContent"`           // 原发病详情
	PrimaryDiseaseType      string `gorm:"type:varchar(255)" json:"primaryDiseaseType"`      // 原发病分类
	PrimaryDiseaseCheckTime string `gorm:"type:varchar(32)" json:"primaryDiseaseCheckTime"`  // 原发病检查时间
	PrimaryDiseaseCheckDoc  string `gorm:"type:varchar(100)" json:"primaryDiseaseCheckDoc"`  // 原发病检查医生
	PathologyName           string `gorm:"type:varchar(255)" json:"pathologyName"`           // 病理诊断名称
	PathologyContent        string `gorm:"type:text" json:"pathologyContent"`                // 病理诊断详情
	PathologyType           string `gorm:"type:varchar(255)" json:"pathologyType"`           // 病理诊断分类
	PathologyCheckTime      string `gorm:"type:varchar(32)" json:"pathologyCheckTime"`       // 病理检查时间
	PathologyCheckDoc       string `gorm:"type:varchar(100)" json:"pathologyCheckDoc"`       // 病理检查医生
	AllergenName            string `gorm:"type:varchar(255)" json:"allergenName"`            // 过敏信息名称
	AllergenContent         string `gorm:"type:text" json:"allergenContent"`                 // 过敏信息详情
	AllergenType            string `gorm:"type:varchar(255)" json:"allergenType"`            // 过敏原分类
	AllergenCheckTime       string `gorm:"type:varchar(32)" json:"allergenCheckTime"`        // 过敏检查时间
	AllergenCheckDoc        string `gorm:"type:varchar(100)" json:"allergenCheckDoc"`        // 过敏检查医生
	TumorHistoryName        string `gorm:"type:varchar(255)" json:"tumorHistoryName"`        // 肿瘤病史名称
	TumorHistoryContent     string `gorm:"type:text" json:"tumorHistoryContent"`             // 肿瘤病史详情
	TumorHistoryType        string `gorm:"type:varchar(255)" json:"tumorHistoryType"`        // 肿瘤分类
	TumorHistoryCheckTime   string `gorm:"type:varchar(32)" json:"tumorHistoryCheckTime"`    // 肿瘤检查时间
	TumorHistoryCheckDoc    string `gorm:"type:varchar(100)" json:"tumorHistoryCheckDoc"`    // 肿瘤检查医生
	ComplicationName        string `gorm:"type:varchar(255)" json:"complicationName"`        // 并发症名称
	ComplicationContent     string `gorm:"type:text" json:"complicationContent"`             // 并发症详情
	ComplicationType        string `gorm:"type:varchar(255)" json:"complicationType"`        // 并发症分类
	ComplicationCheckTime   string `gorm:"type:varchar(32)" json:"complicationCheckTime"`    // 并发症检查时间
	ComplicationCheckDoc    string `gorm:"type:varchar(100)" json:"complicationCheckDoc"`    // 并发症检查医生

	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// TableName 指定表名
func (MedicalHistory) TableName() string {
	return "medical_histories"
}

// OutcomeRecord 治疗转归记录（一个患者多条记录）
type OutcomeRecord struct {
	ID               string    `gorm:"type:varchar(36);primaryKey" json:"id"`
	PatientID        string    `gorm:"type:varchar(36);not null;index" json:"patientId"`
	Type             string    `gorm:"type:varchar(20);not null" json:"type"` // 转入/转出
	Reason           string    `gorm:"type:varchar(255)" json:"reason"`       // 原因
	Time             time.Time `json:"time"`                                  // 转归时间
	Remarks          string    `gorm:"type:text" json:"remarks"`              // 备注
	Registrar        string    `gorm:"type:varchar(50)" json:"registrar"`     // 登记人
	RegistrationTime time.Time `json:"registrationTime"`                      // 登记时间
	IsDoorRule       bool      `gorm:"type:boolean;default:false" json:"isDoorRule"` // 门规

	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// TableName 指定表名
func (OutcomeRecord) TableName() string {
	return "outcome_records"
}

// OutcomeType 转归类型常量
const (
	OutcomeTypeIn  = "转入"
	OutcomeTypeOut = "转出"
)

// InfectionInfo 传染病信息 (作为患者的一部分或单独的表)
type InfectionInfo struct {
	ID        string         `gorm:"type:varchar(36);primaryKey" json:"id"`
	PatientID string         `gorm:"type:varchar(36);not null;uniqueIndex" json:"patientId"`
	HbsAg     string         `gorm:"type:varchar(10);default:阴性" json:"hbsag"` // 乙肝
	HcvAb     string         `gorm:"type:varchar(10);default:阴性" json:"hcvab"` // 丙肝
	HivAb     string         `gorm:"type:varchar(10);default:阴性" json:"hivab"` // 艾滋
	TpaB      string         `gorm:"type:varchar(10);default:阴性" json:"tpab"`  // 梅毒
	Tb        *string        `gorm:"type:varchar(10)" json:"tb"`                // 结核
	UpdateDate time.Time `json:"updateDate"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// TableName 指定表名
func (InfectionInfo) TableName() string {
	return "infection_infos"
}

// InfectionStatus 常量
const (
	InfectionNegative = "阴性"
	InfectionPositive = "阳性"
)

package models

import "time"

// NursingDoc 护理文书（C1）。一张表承载三类文书：量表评估 / 护理记录 / 护理计划。
// content 存 JSON（量表条目取值 / 护理观察操作宣教交班 / 护理问题-措施-评价），随 doc_type 约定结构。
type NursingDoc struct {
	ID          string     `gorm:"column:id;type:varchar(36);primaryKey" json:"id"`
	TenantID    int64      `gorm:"column:tenant_id;index:idx_nd_tenant_patient;not null" json:"tenantId"`
	PatientID   string     `gorm:"column:patient_id;type:varchar(64);index:idx_nd_tenant_patient" json:"patientId"`
	TreatmentID string     `gorm:"column:treatment_id;type:varchar(64);index:idx_nd_treatment" json:"treatmentId"`
	DocType     string     `gorm:"column:doc_type;type:varchar(12);index:idx_nd_type" json:"docType"` // scale / record / plan
	ScaleType   string     `gorm:"column:scale_type;type:varchar(16)" json:"scaleType"`               // morse / braden / catheter / pain（doc_type=scale）
	Score       *int       `gorm:"column:score" json:"score"`                                         // 量表总分
	RiskLevel   string     `gorm:"column:risk_level;type:varchar(12)" json:"riskLevel"`               // high / moderate / low / none
	Content     string     `gorm:"column:content;type:text" json:"content"`                           // JSON
	NurseID     string     `gorm:"column:nurse_id;type:varchar(64)" json:"nurseId"`
	NurseName   string     `gorm:"column:nurse_name;type:varchar(64)" json:"nurseName"`
	RecordedAt  *time.Time `gorm:"column:recorded_at" json:"recordedAt"`
	CreatedAt   time.Time  `gorm:"column:created_at;autoCreateTime" json:"createdAt"`
	UpdatedAt   time.Time  `gorm:"column:updated_at;autoUpdateTime" json:"updatedAt"`
}

func (NursingDoc) TableName() string { return "nursing_doc" }

// 文书类型
const (
	NursingDocScale  = "scale"  // 量表评估
	NursingDocRecord = "record" // 护理记录
	NursingDocPlan   = "plan"   // 护理计划
)

// 风险级别（量表评分映射）
const (
	NursingRiskHigh     = "high"
	NursingRiskModerate = "moderate"
	NursingRiskLow      = "low"
	NursingRiskNone     = "none"
)

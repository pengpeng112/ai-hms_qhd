// DEPRECATED: legacy new-db model, will be rewritten to map legacy hemodialysis DB in Phase 1~5.
package models

import "time"

// LabReport 检验报告主表
type LabReport struct {
	ID                string     `gorm:"type:varchar(36);primaryKey" json:"id"`
	PatientID         string     `gorm:"type:varchar(36);not null;index" json:"patientId"`
	ReportNo          string     `gorm:"type:varchar(64);index" json:"reportNo"`
	ItemCode          string     `gorm:"type:varchar(64)" json:"itemCode"`
	ItemName          string     `gorm:"type:varchar(128)" json:"itemName"`
	ClinicalDiagnosis string     `gorm:"type:text" json:"clinicalDiagnosis"`
	SpecimenType      string     `gorm:"type:varchar(64)" json:"specimenType"`
	Urgency           string     `gorm:"type:varchar(32);default:常规" json:"urgency"`
	RequestDoctor     string     `gorm:"type:varchar(64)" json:"requestDoctor"`
	RequestedAt       *time.Time `json:"requestedAt"`
	SampledAt         *time.Time `json:"sampledAt"`
	ReceivedAt        *time.Time `json:"receivedAt"`
	ReportedAt        *time.Time `json:"reportedAt"`
	Status            string     `gorm:"type:varchar(32);default:已出报告" json:"status"`

	ExternalReportID *string    `gorm:"type:varchar(128);index" json:"externalReportId,omitempty"`
	SourceSystem     string     `gorm:"type:varchar(16);default:LOCAL" json:"sourceSystem"`
	SyncedAt         *time.Time `json:"syncedAt,omitempty"`

	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`

	// 关联
	Items []LabReportItem `gorm:"foreignKey:LabReportID;constraint:OnDelete:CASCADE" json:"items,omitempty"`
}

// TableName 指定表名
func (LabReport) TableName() string {
	return "lab_reports"
}

// LabReportItem 检验报告明细
type LabReportItem struct {
	ID             string     `gorm:"type:varchar(36);primaryKey" json:"id"`
	LabReportID    string     `gorm:"type:varchar(36);not null;index" json:"labReportId"`
	ItemCode       string     `gorm:"type:varchar(64);not null;index" json:"itemCode"`
	ItemName       string     `gorm:"type:varchar(128);not null" json:"itemName"`
	ResultValue    string     `gorm:"type:varchar(64);not null" json:"resultValue"`
	Unit           string     `gorm:"type:varchar(32)" json:"unit"`
	ReferenceRange string     `gorm:"type:varchar(128)" json:"referenceRange"`
	AbnormalFlag   string     `gorm:"type:varchar(8);default:N" json:"abnormalFlag"` // H/L/N
	TestedAt       *time.Time `json:"testedAt"`

	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// TableName 指定表名
func (LabReportItem) TableName() string {
	return "lab_report_items"
}

// ExamReport 检查报告
type ExamReport struct {
	ID        string     `gorm:"column:id;type:varchar(36);primaryKey" json:"id"`
	PatientID string     `gorm:"column:patient_id;type:varchar(36);not null;index" json:"patientId"`
	ExamDate  *time.Time `gorm:"column:exam_date;index" json:"examDate"`

	Title      string `gorm:"column:title;type:varchar(200);not null" json:"title"`
	Conclusion string `gorm:"column:conclusion;type:text" json:"conclusion"`
	Department string `gorm:"column:department;type:varchar(100)" json:"department"`

	ExternalReportID *string    `gorm:"column:external_report_id;type:varchar(128);index" json:"externalReportId,omitempty"`
	SourceSystem     string     `gorm:"column:source_system;type:varchar(16);default:LOCAL" json:"sourceSystem"`
	SyncedAt         *time.Time `gorm:"column:synced_at" json:"syncedAt,omitempty"`

	CreatedAt time.Time `gorm:"column:created_at" json:"createdAt"`
	UpdatedAt time.Time `gorm:"column:updated_at" json:"updatedAt"`
}

// TableName 指定表名
func (ExamReport) TableName() string {
	return "exam_reports"
}

// ExamReportItem 检查报告项目明细
type ExamReportItem struct {
	ID           string    `gorm:"column:id;type:varchar(36);primaryKey" json:"id"`
	ExamReportID string    `gorm:"column:exam_report_id;type:varchar(36);not null;index" json:"examReportId"`
	ItemName     string    `gorm:"column:item_name;type:varchar(200);not null" json:"itemName"`
	ItemCode     *string   `gorm:"column:item_code;type:varchar(64)" json:"itemCode,omitempty"`
	ItemCategory *string   `gorm:"column:item_category;type:varchar(100)" json:"itemCategory,omitempty"`
	ItemResult   *string   `gorm:"column:item_result;type:text" json:"itemResult,omitempty"`
	SortOrder    int       `gorm:"column:sort_order;not null;default:0" json:"sortOrder"`
	CreatedAt    time.Time `gorm:"column:created_at;not null;autoCreateTime" json:"createdAt"`
	UpdatedAt    time.Time `gorm:"column:updated_at;not null;autoUpdateTime" json:"updatedAt"`
}

func (ExamReportItem) TableName() string {
	return "exam_report_items"
}

const (
	// SourceSystemLocal 本地数据
	SourceSystemLocal = "LOCAL"
	// SourceSystemLIS LIS 系统
	SourceSystemLIS = "LIS"
	// SourceSystemPACS PACS 系统
	SourceSystemPACS = "PACS"
	// SourceSystemHDISExam HDIS 检查报告同步
	SourceSystemHDISExam = "HDIS_EXAM"
	// SourceSystemHDISRecord HDIS 关键指标同步
	SourceSystemHDISRecord = "HDIS_RECORD"
	// SourceSystemHISOracleExam HIS Oracle 检查报告同步
	SourceSystemHISOracleExam = "HIS_ORACLE_EXAM"
)

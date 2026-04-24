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
	ID        string     `gorm:"type:varchar(36);primaryKey" json:"id"`
	PatientID string     `gorm:"type:varchar(36);not null;index" json:"patientId"`
	ExamDate  *time.Time `gorm:"index" json:"examDate"`

	Title      string `gorm:"type:varchar(200);not null" json:"title"`
	Conclusion string `gorm:"type:text" json:"conclusion"`
	Department string `gorm:"type:varchar(100)" json:"department"`

	ExternalReportID *string    `gorm:"type:varchar(128);index" json:"externalReportId,omitempty"`
	SourceSystem     string     `gorm:"type:varchar(16);default:LOCAL" json:"sourceSystem"`
	SyncedAt         *time.Time `json:"syncedAt,omitempty"`

	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// TableName 指定表名
func (ExamReport) TableName() string {
	return "exam_reports"
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
)

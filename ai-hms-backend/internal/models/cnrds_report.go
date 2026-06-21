package models

import "time"

const (
	CnrdsTypeMonthly      = "monthly"
	CnrdsTypeEvent        = "event"
	CnrdsEventDeath       = "death"
	CnrdsEventTransplant  = "transplant"
	CnrdsEventTransferOut = "transfer_out"

	CnrdsStatusDraft     = "draft"
	CnrdsStatusExported  = "exported"
	CnrdsStatusSubmitted = "submitted"
)

type CnrdsReport struct {
	ID       string `gorm:"column:id;type:varchar(36);primaryKey" json:"id"`
	TenantID int64  `gorm:"column:tenant_id;not null" json:"tenantId"`

	Period     string `gorm:"column:period;type:varchar(16)" json:"period"`
	ReportType string `gorm:"column:report_type;type:varchar(12)" json:"reportType"`
	EventType  string `gorm:"column:event_type;type:varchar(16)" json:"eventType"`
	PatientID  string `gorm:"column:patient_id;type:varchar(64)" json:"patientId"`

	Content      string `gorm:"column:content;type:text" json:"content"`
	PatientCount int    `gorm:"column:patient_count" json:"patientCount"`

	Status    string `gorm:"column:status;type:varchar(12)" json:"status"`
	ExportRef string `gorm:"column:export_ref;type:varchar(256)" json:"exportRef"`

	ReviewedBy  string     `gorm:"column:reviewed_by;type:varchar(64)" json:"reviewedBy"`
	SubmittedAt *time.Time `gorm:"column:submitted_at" json:"submittedAt"`

	CreatedAt time.Time `gorm:"column:created_at;not null;autoCreateTime" json:"createdAt"`
	UpdatedAt time.Time `gorm:"column:updated_at;not null;autoUpdateTime" json:"updatedAt"`
}

func (CnrdsReport) TableName() string {
	return "cnrds_report"
}

type CnrdsContentRow struct {
	PatientID         string   `json:"patientId"`
	Name              string   `json:"name"`
	Gender            string   `json:"gender"`
	BirthDate         string   `json:"birthDate"`
	PrimaryDiagnosis  string   `json:"primaryDiagnosis"`
	Comorbidity       string   `json:"comorbidity"`
	FirstDialysisDate string   `json:"firstDialysisDate"`
	DialysisMode      string   `json:"dialysisMode"`
	Frequency         string   `json:"frequency"`
	VascularAccess    string   `json:"vascularAccess"`
	Hb                *float64 `json:"hb"`
	Ca                *float64 `json:"ca"`
	P                 *float64 `json:"p"`
	PTH               *float64 `json:"pth"`
	Albumin           *float64 `json:"albumin"`
	KtV               *float64 `json:"ktv"`
	InfMarkers        string   `json:"infMarkers"`
	OutcomeType       string   `json:"outcomeType"`
	OutcomeDate       string   `json:"outcomeDate"`
	DeathReason       string   `json:"deathReason"`
}

type CnrdsContent struct {
	Period string            `json:"period"`
	Type   string            `json:"type"`
	Rows   []CnrdsContentRow `json:"rows"`
}

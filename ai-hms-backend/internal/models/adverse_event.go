package models

import "time"

type AdverseEvent struct {
	ID          string     `gorm:"column:id;type:varchar(36);primaryKey" json:"id"`
	TenantID    int64      `gorm:"column:tenant_id;index:idx_ae_tenant_patient;not null" json:"tenantId"`
	PatientID   int64      `gorm:"column:patient_id;index:idx_ae_tenant_patient" json:"patientId"`
	TreatmentID *int64     `gorm:"column:treatment_id" json:"treatmentId"`
	EventType   string     `gorm:"column:event_type;type:varchar(64);not null" json:"eventType"`
	Severity    string     `gorm:"column:severity;type:varchar(16);not null" json:"severity"`
	OccurredAt  *time.Time `gorm:"column:occurred_at;not null" json:"occurredAt"`
	Description string     `gorm:"column:description;type:text" json:"description"`
	Handling    string     `gorm:"column:handling;type:text" json:"handling"`
	Outcome     string     `gorm:"column:outcome;type:text" json:"outcome"`
	ReporterID  string     `gorm:"column:reporter_id;type:varchar(64)" json:"reporterId"`
	ReportedTo  string     `gorm:"column:reported_to;type:text" json:"reportedTo"`
	ReportedAt  *time.Time `gorm:"column:reported_at" json:"reportedAt"`
	Within6h    *bool      `gorm:"column:within_6h;index:idx_ae_within6h" json:"within6h"`
	Status      string     `gorm:"column:status;type:varchar(16);index:idx_ae_status_severity;not null;default:'registered'" json:"status"`
	CqiLinked   bool       `gorm:"column:cqi_linked;default:false" json:"cqiLinked"`
	CreatedAt   time.Time  `gorm:"column:created_at;autoCreateTime" json:"createdAt"`
	UpdatedAt   time.Time  `gorm:"column:updated_at;autoUpdateTime" json:"updatedAt"`
}

func (AdverseEvent) TableName() string { return "adverse_event" }

const (
	AESeverityMild     = "mild"
	AESeverityModerate = "moderate"
	AESeveritySevere   = "severe"
)

const (
	AEStatusRegistered   = "registered"
	AEStatusReported     = "reported"
	AEStatusAcknowledged = "acknowledged"
	AEStatusProcessing   = "processing"
	AEStatusClosed       = "closed"
)

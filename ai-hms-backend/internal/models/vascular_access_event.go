package models

import "time"

type VascularAccessEvent struct {
	ID         string     `gorm:"column:id;type:varchar(36);primaryKey" json:"id"`
	TenantID   int64      `gorm:"column:tenant_id;index:idx_vae_tenant_patient;not null" json:"tenantId"`
	AccessID   int64      `gorm:"column:access_id;index:idx_vae_access;not null" json:"accessId"`
	PatientID  int64      `gorm:"column:patient_id;index:idx_vae_tenant_patient" json:"patientId"`
	EventType  string     `gorm:"column:event_type;type:varchar(16)" json:"eventType"`
	EventDate  *time.Time `gorm:"column:event_date" json:"eventDate"`
	Detail     string     `gorm:"column:detail;type:text" json:"detail"`
	OperatorID string     `gorm:"column:operator_id;type:varchar(64)" json:"operatorId"`
	Note       string     `gorm:"column:note;type:varchar(256)" json:"note"`
	CreatedAt  time.Time  `gorm:"column:created_at;autoCreateTime" json:"createdAt"`
}

func (VascularAccessEvent) TableName() string { return "vascular_access_event" }

const (
	VAEEstablish     = "establish"
	VAEMaturation    = "maturation"
	VAEFirstUse      = "first_use"
	VAEPhysicalCheck = "physical_check"
	VAEComplication  = "complication"
	VAEIntervention  = "intervention"
	VAEFailure       = "failure"
	VAEReplacement   = "replacement"
)

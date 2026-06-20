package models

import "time"

type ExternalPatientMapping struct {
	ID               string     `gorm:"column:id;type:varchar(36);primaryKey" json:"id"`
	TenantID         int64      `gorm:"column:tenant_id;not null;index:idx_epm_tenant_legacy" json:"tenantId"`
	LegacyPatientID  int64      `gorm:"column:legacy_patient_id;not null;index:idx_epm_tenant_legacy" json:"legacyPatientId"`
	ExternalSystem   string     `gorm:"column:external_system;not null;uniqueIndex:idx_epm_unique,priority:2" json:"externalSystem"`
	ExternalPatientID string    `gorm:"column:external_patient_id;not null;uniqueIndex:idx_epm_unique,priority:3" json:"externalPatientId"`
	ExternalVisitID  *string    `gorm:"column:external_visit_id;uniqueIndex:idx_epm_unique,priority:4" json:"externalVisitId,omitempty"`
	IDNo             *string    `gorm:"column:id_no" json:"idNo,omitempty"`
	DialysisNo       *string    `gorm:"column:dialysis_no" json:"dialysisNo,omitempty"`
	HospNo           *string    `gorm:"column:hosp_no" json:"hospNo,omitempty"`
	CaseNo           *string    `gorm:"column:case_no" json:"caseNo,omitempty"`
	OutpatientNo     *string    `gorm:"column:outpatient_no" json:"outpatientNo,omitempty"`
	MedicalRecordNo  *string    `gorm:"column:medical_record_no" json:"medicalRecordNo,omitempty"`
	PatientName      *string    `gorm:"column:patient_name" json:"patientName,omitempty"`
	MatchStatus      string     `gorm:"column:match_status;not null;default:confirmed" json:"matchStatus"`
	LastSyncedAt     *time.Time `gorm:"column:last_synced_at" json:"lastSyncedAt,omitempty"`
	CreatedAt        time.Time  `gorm:"column:created_at;not null;autoCreateTime" json:"createdAt"`
	UpdatedAt        time.Time  `gorm:"column:updated_at;not null;autoUpdateTime" json:"updatedAt"`
}

func (ExternalPatientMapping) TableName() string {
	return "external_patient_mappings"
}

const (
	MatchStatusConfirmed = "confirmed"
	MatchStatusCandidate = "candidate"
	MatchStatusConflict  = "conflict"
	MatchStatusPending   = "pending"

	ExternalSystemHISOracle = "HIS_ORACLE"
)

package models

import "time"

type PatientACTR struct {
	ID          string `gorm:"column:id;type:varchar(36);primaryKey" json:"id"`
	TenantID    int64  `gorm:"column:tenant_id;not null" json:"tenantId"`
	PatientID   string `gorm:"column:patient_id;not null" json:"patientId"`
	DialysisNo  string `gorm:"column:dialysis_no" json:"dialysisNo"`
	ActrsXrayID int64  `gorm:"column:actrs_xray_id;not null" json:"actrsXrayId"`

	AnalysisDate *time.Time `gorm:"column:analysis_date" json:"analysisDate"`
	CTR          *float64   `gorm:"column:ctr" json:"ctr"`
	ACTR         *float64   `gorm:"column:actr" json:"actr"`
	ACTR1        *float64   `gorm:"column:actr1" json:"actr1"`
	ACTR2        *float64   `gorm:"column:actr2" json:"actr2"`
	ACTRNorm     *float64   `gorm:"column:actr_norm" json:"actrNorm"`
	HeartWidth   *int       `gorm:"column:heart_width" json:"heartWidth"`
	LungWidth    *int       `gorm:"column:lung_width" json:"lungWidth"`
	TiltAngle    *float64   `gorm:"column:tilt_angle" json:"tiltAngle"`

	QCPass       int    `gorm:"column:qc_pass;not null;default:0" json:"qcPass"`
	QCPaAp       string `gorm:"column:qc_pa_ap;type:varchar(8)" json:"qcPaAp"`
	QCWarnings   string `gorm:"column:qc_warnings;type:varchar(256)" json:"qcWarnings"`
	ModelVersion string `gorm:"column:model_version;type:varchar(32)" json:"modelVersion"`
	Source       string `gorm:"column:source;type:varchar(16)" json:"source"`

	ImagePath   string `gorm:"column:image_path;type:varchar(256)" json:"imagePath"`
	OverlayPath string `gorm:"column:overlay_path;type:varchar(256)" json:"overlayPath"`
	MaskPath    string `gorm:"column:mask_path;type:varchar(256)" json:"maskPath"`

	DoctorCorrection *float64   `gorm:"column:doctor_correction" json:"doctorCorrection"`
	CorrectedBy      string     `gorm:"column:corrected_by;type:varchar(64)" json:"correctedBy"`
	CorrectedAt      *time.Time `gorm:"column:corrected_at" json:"correctedAt"`

	AdoptedBy             string     `gorm:"column:adopted_by;type:varchar(64)" json:"adoptedBy"`
	AdoptedAt             *time.Time `gorm:"column:adopted_at" json:"adoptedAt"`
	AdoptedPrescriptionID string     `gorm:"column:adopted_prescription_id;type:varchar(32)" json:"adoptedPrescriptionId"`
	AdoptedDryWeight      *float64   `gorm:"column:adopted_dry_weight" json:"adoptedDryWeight"`
	AdoptedUFQuantity     *float64   `gorm:"column:adopted_uf_quantity" json:"adoptedUfQuantity"`

	Notes    string     `gorm:"column:notes;type:varchar(256)" json:"notes"`
	SyncedAt *time.Time `gorm:"column:synced_at" json:"syncedAt"`

	CreatedAt time.Time `gorm:"column:created_at;not null;autoCreateTime" json:"createdAt"`
	UpdatedAt time.Time `gorm:"column:updated_at;not null;autoUpdateTime" json:"updatedAt"`
}

func (PatientACTR) TableName() string {
	return "patient_actr"
}

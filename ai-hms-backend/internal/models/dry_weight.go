package models

import "time"

type DryWeightAssessment struct {
	ID            string     `gorm:"column:id;type:varchar(36);primaryKey" json:"id"`
	TenantID      int64      `gorm:"column:tenant_id;index:idx_dwa_tenant_patient;not null" json:"tenantId"`
	PatientID     int64      `gorm:"column:patient_id;index:idx_dwa_tenant_patient" json:"patientId"`
	AssessType    string     `gorm:"column:assess_type;type:varchar(16);not null" json:"assessType"` // daily/cycle
	Phase         string     `gorm:"column:phase;type:varchar(16);not null" json:"phase"`           // induction/maintenance
	SBP           *int       `gorm:"column:sbp" json:"sbp"`
	DBP           *int       `gorm:"column:dbp" json:"dbp"`
	HeartRate     *int       `gorm:"column:heart_rate" json:"heartRate"`
	Edema         bool       `gorm:"column:edema;default:false" json:"edema"`
	Palpitation   bool       `gorm:"column:palpitation;default:false" json:"palpitation"`
	HeartFailure  bool       `gorm:"column:heart_failure;default:false" json:"heartFailure"`
	Cramp         bool       `gorm:"column:cramp;default:false" json:"cramp"`
	CTR           *float64   `gorm:"column:ctr" json:"ctr"`
	ACTR          *float64   `gorm:"column:actr" json:"actr"`
	BIAOH         *float64   `gorm:"column:bia_oh" json:"biaOh"`
	BIATBW        *float64   `gorm:"column:bia_tbw" json:"biaTbw"`
	BIAECW        *float64   `gorm:"column:bia_ecw" json:"biaEcw"`
	PostWeight    *float64   `gorm:"column:post_weight" json:"postWeight"`
	TargetWeight  *float64   `gorm:"column:target_weight" json:"targetWeight"`
	Decision      string     `gorm:"column:decision;type:varchar(16)" json:"decision"`           // hold/lower/raise
	AdjustKg      *float64   `gorm:"column:adjust_kg" json:"adjustKg"`
	RNaSetting    *float64   `gorm:"column:rna_setting" json:"rnaSetting"`
	MainMet       bool       `gorm:"column:main_met;default:false" json:"mainMet"`
	FailedReasons string     `gorm:"column:failed_reasons;type:text" json:"failedReasons"` // JSON数组
	AssessorID    string     `gorm:"column:assessor_id;type:varchar(64)" json:"assessorId"`
	AssessorName  string     `gorm:"column:assessor_name;type:varchar(64)" json:"assessorName"`
	CreatedAt     time.Time  `gorm:"column:created_at;autoCreateTime" json:"createdAt"`
}

func (DryWeightAssessment) TableName() string { return "dry_weight_assessment" }

type PatientDryWeight struct {
	ID            string     `gorm:"column:id;type:varchar(36);primaryKey" json:"id"`
	TenantID      int64      `gorm:"column:tenant_id;uniqueIndex:idx_pdw_tenant_patient;not null" json:"tenantId"`
	PatientID     int64      `gorm:"column:patient_id;uniqueIndex:idx_pdw_tenant_patient" json:"patientId"`
	DryWeight     float64    `gorm:"column:dry_weight;not null" json:"dryWeight"`
	StandardACTR  *float64   `gorm:"column:standard_actr" json:"standardActr"`
	StandardCTR   *float64   `gorm:"column:standard_ctr" json:"standardCtr"`
	Phase         string     `gorm:"column:phase;type:varchar(16);not null;default:'maintenance'" json:"phase"`
	ConfirmedBy   string     `gorm:"column:confirmed_by;type:varchar(64)" json:"confirmedBy"`
	ConfirmedName string     `gorm:"column:confirmed_name;type:varchar(64)" json:"confirmedName"`
	ConfirmedAt   time.Time  `gorm:"column:confirmed_at" json:"confirmedAt"`
	CreatedAt     time.Time  `gorm:"column:created_at;autoCreateTime" json:"createdAt"`
	UpdatedAt     time.Time  `gorm:"column:updated_at;autoUpdateTime" json:"updatedAt"`
}

func (PatientDryWeight) TableName() string { return "patient_dry_weight" }

const (
	DWPhaseInduction   = "induction"
	DWPhaseMaintenance = "maintenance"

	DWAssessDaily = "daily"
	DWAssessCycle = "cycle"

	DWHold  = "hold"
	DWLower = "lower"
	DWRaise = "raise"
)

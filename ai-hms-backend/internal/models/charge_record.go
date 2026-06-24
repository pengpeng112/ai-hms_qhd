package models

import "time"

const (
	ChargeCatTreatment = "treatment"
	ChargeCatMaterial  = "material"
	ChargeCatNursing   = "nursing"
	ChargeCatInjection = "injection"
	ChargeCatDrug      = "drug"

	ChargeStatusDraft     = "draft"
	ChargeStatusConfirmed = "confirmed"
	ChargeStatusChecked   = "checked"
	ChargeStatusPushed    = "pushed"
	ChargeStatusSettled   = "settled"
	ChargeStatusCancelled = "cancelled"

	ChargeSourceAuto   = "auto"
	ChargeSourceManual = "manual"

	PriceSourceHisPriceList = "his_price_list"
	PriceSourceCatalog      = "billing_catalog"
	PriceSourceManual       = "manual"
	PriceSourceUnknown      = "unknown"

	MatchStatusMatched   = "matched"
	MatchStatusMultiple  = "multiple"
	MatchStatusUnmatched = "unmatched"
	MatchStatusManual    = "manual"
)

type ChargeRecord struct {
	ID             string     `gorm:"column:id;type:varchar(36);primaryKey" json:"id"`
	TenantID       int64      `gorm:"column:tenant_id;not null" json:"tenantId"`
	PatientID      *int64     `gorm:"column:patient_id" json:"patientId"`
	TreatmentID    int64      `gorm:"column:treatment_id;not null" json:"treatmentId"`
	PrescriptionID *int64     `gorm:"column:prescription_id" json:"prescriptionId"`
	ChargeDate     *time.Time `gorm:"column:charge_date" json:"chargeDate"`
	Shift          *string    `gorm:"column:shift;type:varchar(16)" json:"shift"`
	DialysisMode   *string    `gorm:"column:dialysis_mode;type:varchar(16)" json:"dialysisMode"`
	AccessType     *string    `gorm:"column:access_type;type:varchar(16)" json:"accessType"`
	CrrtHours      *float64   `gorm:"column:crrt_hours;type:decimal(5,2)" json:"crrtHours"`
	TotalAmount    *float64   `gorm:"column:total_amount;type:decimal(10,2)" json:"totalAmount"`
	Status         string     `gorm:"column:status;type:varchar(16);not null;default:draft" json:"status"`
	RecordedBy     *string    `gorm:"column:recorded_by;type:varchar(64)" json:"recordedBy"`
	RecordedName   *string    `gorm:"column:recorded_name;type:varchar(64)" json:"recordedName"`
	CheckedBy      *string    `gorm:"column:checked_by;type:varchar(64)" json:"checkedBy"`
	CheckedName    *string    `gorm:"column:checked_name;type:varchar(64)" json:"checkedName"`
	CheckedAt      *time.Time `gorm:"column:checked_at" json:"checkedAt"`
	ExportedAt     *time.Time `gorm:"column:exported_at" json:"exportedAt"`
	PushedAt       *time.Time `gorm:"column:pushed_at" json:"pushedAt"`
	Note           *string    `gorm:"column:note;type:varchar(256)" json:"note"`
	CreatedAt      time.Time  `gorm:"column:created_at;not null;autoCreateTime" json:"createdAt"`
	UpdatedAt      time.Time  `gorm:"column:updated_at;not null;autoUpdateTime" json:"updatedAt"`
	Lines          []ChargeLine `gorm:"foreignKey:ChargeRecordID" json:"lines,omitempty"`
}

func (ChargeRecord) TableName() string {
	return "charge_record"
}

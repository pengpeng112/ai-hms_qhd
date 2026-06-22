package models

import "time"

type MedicationAdmin struct {
	ID                string     `gorm:"column:id;type:varchar(36);primaryKey" json:"id"`
	TenantID          int64      `gorm:"column:tenant_id;index:idx_ma_tenant_patient;not null" json:"tenantId"`
	PatientID         int64      `gorm:"column:patient_id;index:idx_ma_tenant_patient" json:"patientId"`
	OrderID           int64      `gorm:"column:order_id;index:idx_ma_order;not null" json:"orderId"`
	TreatmentID       int64      `gorm:"column:treatment_id;index:idx_ma_treatment" json:"treatmentId"`
	DrugName          string     `gorm:"column:drug_name;type:varchar(128);not null" json:"drugName"`
	Category          string     `gorm:"column:category;type:varchar(32)" json:"category"`
	Dose              string     `gorm:"column:dose;type:varchar(64)" json:"dose"`
	Route             string     `gorm:"column:route;type:varchar(32)" json:"route"`
	Timing            string     `gorm:"column:timing;type:varchar(32)" json:"timing"`
	AdministeredBy    string     `gorm:"column:administered_by;type:varchar(64);not null" json:"administeredBy"`
	AdministeredName  string     `gorm:"column:administered_name;type:varchar(64)" json:"administeredName"`
	AdministeredAt    time.Time  `gorm:"column:administered_at;not null" json:"administeredAt"`
	SecondCheckBy     string     `gorm:"column:second_check_by;type:varchar(64)" json:"secondCheckBy"`
	SecondCheckName   string     `gorm:"column:second_check_name;type:varchar(64)" json:"secondCheckName"`
	SecondCheckAt     *time.Time `gorm:"column:second_check_at" json:"secondCheckAt"`
	Status            string     `gorm:"column:status;type:varchar(16);not null;default:'recorded'" json:"status"`
	Note              string     `gorm:"column:note;type:varchar(256)" json:"note"`
	CreatedAt         time.Time  `gorm:"column:created_at;autoCreateTime" json:"createdAt"`
	UpdatedAt         time.Time  `gorm:"column:updated_at;autoUpdateTime" json:"updatedAt"`
}

func (MedicationAdmin) TableName() string { return "medication_admin" }

const (
	MAStatusRecorded = "recorded"
	MAStatusVerified = "verified"
)

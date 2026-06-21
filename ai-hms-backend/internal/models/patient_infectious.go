package models

import "time"

type PatientInfectious struct {
	ID                 string     `gorm:"column:id;type:varchar(36);primaryKey" json:"id"`
	TenantID           int64      `gorm:"column:tenant_id;index:idx_inf_tenant_patient;not null" json:"tenantId"`
	PatientID          string     `gorm:"column:patient_id;index:idx_inf_tenant_patient;not null" json:"patientId"`
	ScreenDate         *time.Time `gorm:"column:screen_date" json:"screenDate"`
	Items              string     `gorm:"column:items;type:text" json:"items"` // JSON [{item,result}]
	Source             string     `gorm:"column:source;type:varchar(8)" json:"source"`
	ResultOverall      string     `gorm:"column:result_overall;type:varchar(8)" json:"resultOverall"`
	PositiveMarkers    string     `gorm:"column:positive_markers;type:varchar(128)" json:"positiveMarkers"`
	NextDueDate        *time.Time `gorm:"column:next_due_date" json:"nextDueDate"`
	Disposition        string     `gorm:"column:disposition;type:varchar(16)" json:"disposition"`
	HandledDoctorID    string     `gorm:"column:handled_doctor_id;type:varchar(64)" json:"handledDoctorId"`
	HandledHeadnurseID string     `gorm:"column:handled_headnurse_id;type:varchar(64)" json:"handledHeadnurseId"`
	HandledAt          *time.Time `gorm:"column:handled_at" json:"handledAt"`
	ZoneTag            string     `gorm:"column:zone_tag;type:varchar(16)" json:"zoneTag"`
	Note               string     `gorm:"column:note;type:varchar(256)" json:"note"`
	CreatedAt          time.Time  `gorm:"column:created_at;autoCreateTime" json:"createdAt"`
	UpdatedAt          time.Time  `gorm:"column:updated_at;autoUpdateTime" json:"updatedAt"`
}

func (PatientInfectious) TableName() string { return "patient_infectious" }

// result_overall
const (
	InfectiousNegative = "negative"
	InfectiousPositive = "positive"
	InfectiousPending  = "pending"
)

// disposition
const (
	InfectiousDispCZoneCRRT   = "c_zone_crrt"
	InfectiousDispTransferOut = "transfer_out"
)

// item result
const (
	InfItemNegative      = "negative"
	InfItemPositive      = "positive"
	InfItemIndeterminate = "indeterminate"
)

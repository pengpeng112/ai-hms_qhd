package models

import "time"

// KioskPreSignMeasurement 自助站透前体征测量明细（追加保存每次设备上报）。
type KioskPreSignMeasurement struct {
	ID            string    `gorm:"column:id;primaryKey" json:"id"`
	TenantID      int64     `gorm:"column:tenant_id;not null" json:"tenantId"`
	TreatmentID   int64     `gorm:"column:treatment_id;not null" json:"treatmentId"`
	PatientID     int64     `gorm:"column:patient_id;not null" json:"patientId"`
	MeasuredAt    time.Time `gorm:"column:measured_at;not null" json:"measuredAt"`
	Weight        *float64  `gorm:"column:weight" json:"weight,omitempty"`
	SBP           *float64  `gorm:"column:sbp" json:"sbp,omitempty"`
	DBP           *float64  `gorm:"column:dbp" json:"dbp,omitempty"`
	BodyTemp      *float64  `gorm:"column:body_temp" json:"bodyTemp,omitempty"`
	HeartRate     *float64  `gorm:"column:heart_rate" json:"heartRate,omitempty"`
	Respiration   *float64  `gorm:"column:respiration" json:"respiration,omitempty"`
	DeviceID      string    `gorm:"column:device_id" json:"deviceId,omitempty"`
	Source        string    `gorm:"column:source;not null;default:newsystem" json:"source"`
	ClientEventID string    `gorm:"column:client_event_id" json:"clientEventId,omitempty"`
	RawPayload    string    `gorm:"column:raw_payload" json:"rawPayload,omitempty"`
	CreatedAt     time.Time `gorm:"column:created_at" json:"createdAt"`
}

func (KioskPreSignMeasurement) TableName() string { return "kiosk_pre_sign_measurement" }

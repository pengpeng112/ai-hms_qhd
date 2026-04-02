package models

import "time"

// PatientKeyIndicator 患者关键指标（HDIS Record 同步）
type PatientKeyIndicator struct {
	ID        string `gorm:"type:varchar(36);primaryKey" json:"id"`
	PatientID string `gorm:"type:varchar(36);not null;index" json:"patientId"`

	ExternalRecordID string `gorm:"type:varchar(128);not null;index:idx_patient_key_indicators_unique,unique" json:"externalRecordId"`
	SourceSystem     string `gorm:"type:varchar(32);not null;index:idx_patient_key_indicators_unique,unique" json:"sourceSystem"`

	IndexName string `gorm:"type:varchar(200);not null" json:"indexName"`
	IndexCode string `gorm:"type:varchar(64);index" json:"indexCode"`
	Result    string `gorm:"type:varchar(128)" json:"result"`
	Unit      string `gorm:"type:varchar(64)" json:"unit"`
	Reference string `gorm:"type:varchar(200)" json:"reference"`

	ResultSign       string     `gorm:"type:varchar(8)" json:"resultSign"`
	TestTime         *time.Time `gorm:"index" json:"testTime"`
	EvaluationResult string     `gorm:"type:varchar(64)" json:"evaluationResult"`
	SyncedAt         *time.Time `json:"syncedAt,omitempty"`

	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// TableName 指定表名
func (PatientKeyIndicator) TableName() string {
	return "patient_key_indicators"
}

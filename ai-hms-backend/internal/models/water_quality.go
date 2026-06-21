package models

import "time"

type WaterQuality struct {
	ID                 string     `gorm:"column:id;type:varchar(36);primaryKey" json:"id"`
	TenantID           int64      `gorm:"column:tenant_id;index:idx_wq_tenant_type_date;not null" json:"tenantId"`
	TestDate           *time.Time `gorm:"column:test_date;index:idx_wq_tenant_type_date" json:"testDate"`
	TestType           string     `gorm:"column:test_type;type:varchar(24);index:idx_wq_tenant_type_date" json:"testType"`
	SamplePoint        string     `gorm:"column:sample_point;type:varchar(16)" json:"samplePoint"`
	DeviceID           string     `gorm:"column:device_id;type:varchar(64)" json:"deviceId"`
	Value              float64    `gorm:"column:value" json:"value"`
	Unit               string     `gorm:"column:unit;type:varchar(16)" json:"unit"`
	StandardLimit      string     `gorm:"column:standard_limit;type:varchar(32)" json:"standardLimit"`
	Result             string     `gorm:"column:result;type:varchar(8)" json:"result"`
	Source             string     `gorm:"column:source;type:varchar(12)" json:"source"`
	NextDueDate        *time.Time `gorm:"column:next_due_date" json:"nextDueDate"`
	HandledEngineerID  string     `gorm:"column:handled_engineer_id;type:varchar(64)" json:"handledEngineerId"`
	HandledHeadnurseID string     `gorm:"column:handled_headnurse_id;type:varchar(64)" json:"handledHeadnurseId"`
	HandledAt          *time.Time `gorm:"column:handled_at" json:"handledAt"`
	Action             string     `gorm:"column:action;type:varchar(256)" json:"action"`
	CreatedAt          time.Time  `gorm:"column:created_at;autoCreateTime" json:"createdAt"`
	UpdatedAt          time.Time  `gorm:"column:updated_at;autoUpdateTime" json:"updatedAt"`
}

func (WaterQuality) TableName() string { return "water_quality" }

const (
	WQResultPass    = "pass"
	WQResultFail    = "fail"
	WQResultPending = "pending"
)

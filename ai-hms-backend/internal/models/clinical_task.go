// DEPRECATED: legacy new-db model, will be rewritten to map legacy hemodialysis DB in Phase 1~5.
package models

import "time"

type ClinicalTask struct {
	ID          int64      `gorm:"primaryKey;autoIncrement" json:"id"`
	TenantId    int64      `gorm:"index;not null" json:"tenantId"`
	Type        string     `gorm:"size:32;not null" json:"type"`
	Title       string     `gorm:"size:100;not null" json:"title"`
	Description string     `gorm:"size:512" json:"description"`
	PatientId   *int64     `gorm:"index" json:"patientId"`
	PatientName string     `gorm:"size:50" json:"patientName"`
	BedNumber   string     `gorm:"size:20" json:"bedNumber"`
	Severity    string     `gorm:"size:20;default:medium" json:"severity"`
	Status      string     `gorm:"size:20;default:pending" json:"status"`
	AssignedTo  *int64     `gorm:"index" json:"assignedTo"`
	HandledAt   *time.Time `json:"handledAt"`
	HandledBy   *int64     `json:"handledBy"`
	CreatedAt   time.Time  `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt   time.Time  `gorm:"autoUpdateTime" json:"updatedAt"`
}

func (ClinicalTask) TableName() string {
	return "clinical_tasks"
}

package models

import "time"

type SyncJobConfig struct {
	ID              string     `gorm:"column:id;type:varchar(36);primaryKey" json:"id"`
	JobCode         string     `gorm:"column:job_code;not null;uniqueIndex:idx_sync_job_configs_code" json:"jobCode"`
	SourceSystem    string     `gorm:"column:source_system;not null" json:"sourceSystem"`
	SyncType        string     `gorm:"column:sync_type;not null" json:"syncType"`
	Enabled         bool       `gorm:"column:enabled;not null;default:false" json:"enabled"`
	CronExpr        *string    `gorm:"column:cron_expr" json:"cronExpr,omitempty"`
	IntervalSeconds *int       `gorm:"column:interval_seconds" json:"intervalSeconds,omitempty"`
	BatchSize       int        `gorm:"column:batch_size;not null;default:500" json:"batchSize"`
	TimeoutSeconds  int        `gorm:"column:timeout_seconds;not null;default:60" json:"timeoutSeconds"`
	MaxRetry        int        `gorm:"column:max_retry;not null;default:3" json:"maxRetry"`
	CursorType      string     `gorm:"column:cursor_type;not null;default:time" json:"cursorType"`
	CursorValue     *string    `gorm:"column:cursor_value" json:"cursorValue,omitempty"`
	OverwritePolicy string     `gorm:"column:overwrite_policy;not null;default:fill_empty" json:"overwritePolicy"`
	LastRunAt       *time.Time `gorm:"column:last_run_at" json:"lastRunAt,omitempty"`
	NextRunAt       *time.Time `gorm:"column:next_run_at" json:"nextRunAt,omitempty"`
	CreatedAt       time.Time  `gorm:"column:created_at;not null;autoCreateTime" json:"createdAt"`
	UpdatedAt       time.Time  `gorm:"column:updated_at;not null;autoUpdateTime" json:"updatedAt"`
}

func (SyncJobConfig) TableName() string {
	return "sync_job_configs"
}

type SyncJobRun struct {
	ID           string     `gorm:"column:id;type:varchar(36);primaryKey" json:"id"`
	JobCode      string     `gorm:"column:job_code;not null;index:idx_sync_job_runs_job_start,priority:1" json:"jobCode"`
	SourceSystem string     `gorm:"column:source_system;not null" json:"sourceSystem"`
	SyncType     string     `gorm:"column:sync_type;not null" json:"syncType"`
	Status       string     `gorm:"column:status;not null;index:idx_sync_job_runs_status" json:"status"`
	StartedAt    time.Time  `gorm:"column:started_at;not null;index:idx_sync_job_runs_job_start,priority:2" json:"startedAt"`
	FinishedAt   *time.Time `gorm:"column:finished_at" json:"finishedAt,omitempty"`
	DurationMs   *int64     `gorm:"column:duration_ms" json:"durationMs,omitempty"`
	FetchedCount int        `gorm:"column:fetched_count;not null;default:0" json:"fetchedCount"`
	CreatedCount int        `gorm:"column:created_count;not null;default:0" json:"createdCount"`
	UpdatedCount int        `gorm:"column:updated_count;not null;default:0" json:"updatedCount"`
	SkippedCount int        `gorm:"column:skipped_count;not null;default:0" json:"skippedCount"`
	FailedCount  int        `gorm:"column:failed_count;not null;default:0" json:"failedCount"`
	CursorBefore *string    `gorm:"column:cursor_before" json:"cursorBefore,omitempty"`
	CursorAfter  *string    `gorm:"column:cursor_after" json:"cursorAfter,omitempty"`
	ErrorMessage *string    `gorm:"column:error_message;type:text" json:"errorMessage,omitempty"`
	CreatedAt    time.Time  `gorm:"column:created_at;not null;autoCreateTime" json:"createdAt"`
}

func (SyncJobRun) TableName() string {
	return "sync_job_runs"
}

const (
	SyncJobStatusRunning = "running"
	SyncJobStatusSuccess = "success"
	SyncJobStatusPartial = "partial"
	SyncJobStatusFailed  = "failed"

	SyncJobCodeExamReport = "his_exam_report"
	SyncJobCodePatientArchive = "his_patient_archive"

	SyncTypeExamReport      = "exam_report"
	SyncTypePatientArchive  = "patient_archive"

	CursorTypeTime  = "time"
	CursorTypeMixed = "mixed"

	OverwritePolicyFillEmpty = "fill_empty"
	OverwritePolicyAlways    = "always"
)

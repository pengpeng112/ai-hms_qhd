package services

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/elliotxin/ai-hms-backend/internal/database"
	"github.com/elliotxin/ai-hms-backend/internal/models"
	"github.com/elliotxin/ai-hms-backend/internal/utils"
	"gorm.io/gorm"
)

type SyncJobService struct {
	db *gorm.DB
}

func NewSyncJobService() *SyncJobService {
	return &SyncJobService{db: database.GetDB()}
}

func (s *SyncJobService) ListJobs() ([]models.SyncJobConfig, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}
	var jobs []models.SyncJobConfig
	if err := s.db.Order("job_code ASC").Find(&jobs).Error; err != nil {
		if isIgnorableLegacyQueryError(err) {
			return []models.SyncJobConfig{}, nil
		}
		return nil, err
	}
	return jobs, nil
}

func (s *SyncJobService) GetJob(jobCode string) (*models.SyncJobConfig, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}
	jobCode = strings.TrimSpace(jobCode)
	if jobCode == "" {
		return nil, errors.New("job code is required")
	}
	var job models.SyncJobConfig
	if err := s.db.Where("job_code = ?", jobCode).First(&job).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("sync job not found: %s", jobCode)
		}
		return nil, err
	}
	return &job, nil
}

type UpdateSyncJobRequest struct {
	Enabled       *bool   `json:"enabled,omitempty"`
	BatchSize     *int    `json:"batchSize,omitempty"`
	TimeoutSeconds *int   `json:"timeoutSeconds,omitempty"`
	IntervalSeconds *int  `json:"intervalSeconds,omitempty"`
	MaxRetry      *int    `json:"maxRetry,omitempty"`
	CursorType    *string `json:"cursorType,omitempty"`
	CursorValue   *string `json:"cursorValue,omitempty"`
	OverwritePolicy *string `json:"overwritePolicy,omitempty"`
}

func (s *SyncJobService) UpdateJob(jobCode string, req UpdateSyncJobRequest) (*models.SyncJobConfig, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}
	jobCode = strings.TrimSpace(jobCode)
	if jobCode == "" {
		return nil, errors.New("job code is required")
	}
	var job models.SyncJobConfig
	if err := s.db.Where("job_code = ?", jobCode).First(&job).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("sync job not found: %s", jobCode)
		}
		return nil, err
	}
	updates := map[string]interface{}{"updated_at": time.Now()}
	if req.Enabled != nil {
		updates["enabled"] = *req.Enabled
	}
	if req.BatchSize != nil {
		updates["batch_size"] = *req.BatchSize
	}
	if req.TimeoutSeconds != nil {
		updates["timeout_seconds"] = *req.TimeoutSeconds
	}
	if req.IntervalSeconds != nil {
		updates["interval_seconds"] = *req.IntervalSeconds
	}
	if req.MaxRetry != nil {
		updates["max_retry"] = *req.MaxRetry
	}
	if req.CursorType != nil {
		updates["cursor_type"] = *req.CursorType
	}
	if req.CursorValue != nil {
		updates["cursor_value"] = *req.CursorValue
	}
	if req.OverwritePolicy != nil {
		updates["overwrite_policy"] = *req.OverwritePolicy
	}
	if err := s.db.Model(&job).Updates(updates).Error; err != nil {
		return nil, err
	}
	return s.GetJob(jobCode)
}

func (s *SyncJobService) GetJobRuns(jobCode string, limit int) ([]models.SyncJobRun, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}
	jobCode = strings.TrimSpace(jobCode)
	if jobCode == "" {
		return nil, errors.New("job code is required")
	}
	if limit <= 0 {
		limit = 50
	}
	if limit > 500 {
		limit = 500
	}
	var runs []models.SyncJobRun
	err := s.db.Where("job_code = ?", jobCode).Order("started_at DESC").Limit(limit).Find(&runs).Error
	if err != nil {
		if isIgnorableLegacyQueryError(err) {
			return []models.SyncJobRun{}, nil
		}
		return nil, err
	}
	return runs, nil
}

func (s *SyncJobService) CreateRun(run *models.SyncJobRun) error {
	if s.db == nil {
		return errors.New("database not available")
	}
	if run.ID == "" {
		run.ID = utils.GenerateID()
	}
	return s.db.Create(run).Error
}

func (s *SyncJobService) FinishRun(runID string, status string, counts map[string]int, cursorAfter, cursorBefore, errMsg string) error {
	if s.db == nil {
		return errors.New("database not available")
	}
	now := time.Now()
	var run models.SyncJobRun
	if err := s.db.Where("id = ?", runID).First(&run).Error; err != nil {
		return err
	}
	updates := map[string]interface{}{
		"status":      status,
		"finished_at": now,
	}
	if counts != nil {
		for k, v := range counts {
			switch k {
			case "fetched":
				updates["fetched_count"] = v
			case "created":
				updates["created_count"] = v
			case "updated":
				updates["updated_count"] = v
			case "skipped":
				updates["skipped_count"] = v
			case "failed":
				updates["failed_count"] = v
			}
		}
	}
	if cursorAfter != "" {
		updates["cursor_after"] = cursorAfter
	}
	if cursorBefore != "" {
		updates["cursor_before"] = cursorBefore
	}
	if errMsg != "" {
		updates["error_message"] = errMsg
	}
	updates["duration_ms"] = now.Sub(run.StartedAt).Milliseconds()
	if err := s.db.Model(&run).Updates(updates).Error; err != nil {
		return err
	}
	if status == models.SyncJobStatusSuccess || status == models.SyncJobStatusPartial {
		s.db.Model(&models.SyncJobConfig{}).Where("job_code = ?", run.JobCode).Update("last_run_at", now)
		if cursorAfter != "" {
			s.db.Model(&models.SyncJobConfig{}).Where("job_code = ?", run.JobCode).Update("cursor_value", cursorAfter)
		}
	}
	return nil
}

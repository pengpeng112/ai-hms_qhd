package services

import (
	"errors"
	"time"

	"github.com/elliotxin/ai-hms-backend/internal/database"
	"github.com/elliotxin/ai-hms-backend/internal/models"
)

type ClinicalTaskService struct{}

func NewClinicalTaskService() *ClinicalTaskService {
	return &ClinicalTaskService{}
}

func (s *ClinicalTaskService) List(status string, tenantId int64) ([]models.ClinicalTask, int64, error) {
	db := database.GetDB()
	if db == nil {
		return nil, 0, errors.New("database not available")
	}
	if err := ensureTables(db, &models.ClinicalTask{}); err != nil {
		return nil, 0, err
	}

	query := db.Model(&models.ClinicalTask{}).Where("tenant_id = ?", tenantId)
	if status != "" {
		query = query.Where("status = ?", status)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var tasks []models.ClinicalTask
	if err := query.Order("created_at DESC").Find(&tasks).Error; err != nil {
		return nil, 0, err
	}
	return tasks, total, nil
}

func (s *ClinicalTaskService) UpdateStatus(id int64, status string, handledBy int64, tenantId int64) error {
	db := database.GetDB()
	if db == nil {
		return errors.New("database not available")
	}
	if err := ensureTables(db, &models.ClinicalTask{}); err != nil {
		return err
	}

	updates := map[string]any{
		"status": status,
	}
	if status == "handled" || status == "dismissed" {
		now := time.Now()
		updates["handled_at"] = &now
		updates["handled_by"] = handledBy
	}

	return db.
		Model(&models.ClinicalTask{}).
		Where("id = ? AND tenant_id = ?", id, tenantId).
		Updates(updates).Error
}

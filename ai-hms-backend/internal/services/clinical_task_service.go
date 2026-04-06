package services

import (
	"time"

	"github.com/elliotxin/ai-hms-backend/internal/database"
	"github.com/elliotxin/ai-hms-backend/internal/models"
)

type ClinicalTaskService struct{}

func NewClinicalTaskService() *ClinicalTaskService {
	return &ClinicalTaskService{}
}

func (s *ClinicalTaskService) List(status string, tenantId int64) ([]models.ClinicalTask, int64, error) {
	db := database.GetDB().Model(&models.ClinicalTask{}).Where("tenant_id = ?", tenantId)
	if status != "" {
		db = db.Where("status = ?", status)
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var tasks []models.ClinicalTask
	if err := db.Order("created_at DESC").Find(&tasks).Error; err != nil {
		return nil, 0, err
	}
	return tasks, total, nil
}

func (s *ClinicalTaskService) UpdateStatus(id int64, status string, handledBy int64, tenantId int64) error {
	updates := map[string]any{
		"status": status,
	}
	if status == "handled" || status == "dismissed" {
		now := time.Now()
		updates["handled_at"] = &now
		updates["handled_by"] = handledBy
	}

	return database.GetDB().
		Model(&models.ClinicalTask{}).
		Where("id = ? AND tenant_id = ?", id, tenantId).
		Updates(updates).Error
}

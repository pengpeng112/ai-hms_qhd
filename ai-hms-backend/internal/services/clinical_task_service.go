package services

import (
	"github.com/elliotxin/ai-hms-backend/internal/models"
)

type ClinicalTaskService struct{}

func NewClinicalTaskService() *ClinicalTaskService {
	return &ClinicalTaskService{}
}

// List 老库无 clinical_tasks 表，暂返回空
func (s *ClinicalTaskService) List(status string, tenantId int64) ([]models.ClinicalTask, int64, error) {
	return []models.ClinicalTask{}, 0, nil
}

// UpdateStatus 老库暂不支持临床任务状态更新
func (s *ClinicalTaskService) UpdateStatus(id int64, status string, handledBy int64, tenantId int64) error {
	return nil
}

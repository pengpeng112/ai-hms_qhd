package services

import (
	"errors"

	"github.com/elliotxin/ai-hms-backend/internal/database"
	"github.com/elliotxin/ai-hms-backend/internal/models"
	"gorm.io/gorm"
)

type PermissionService struct {
	db *gorm.DB
}

func NewPermissionService() *PermissionService {
	return &PermissionService{db: database.GetDB()}
}

func (s *PermissionService) ListPermissions() ([]models.Permission, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}
	var items []models.Permission
	if err := s.db.Order("module ASC, code ASC").Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

func (s *PermissionService) SavePermission(input models.Permission) (*models.Permission, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}
	var existing models.Permission
	err := s.db.Where("code = ?", input.Code).First(&existing).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		if err := s.db.Create(&input).Error; err != nil {
			return nil, err
		}
		return &input, nil
	}
	if err != nil {
		return nil, err
	}

	existing.Name = input.Name
	existing.Description = input.Description
	existing.Module = input.Module
	existing.Action = input.Action
	if input.Status != "" {
		existing.Status = input.Status
	}
	if err := s.db.Save(&existing).Error; err != nil {
		return nil, err
	}
	return &existing, nil
}

func (s *PermissionService) GetRolePermissionCodes(role string) ([]string, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}
	var rows []models.RolePermission
	if err := s.db.Where("role = ?", role).Order("permission_code ASC").Find(&rows).Error; err != nil {
		return nil, err
	}
	result := make([]string, 0, len(rows))
	for _, row := range rows {
		result = append(result, row.PermissionCode)
	}
	return result, nil
}

func (s *PermissionService) SetRolePermissionCodes(role string, permissionCodes []string) error {
	if s.db == nil {
		return errors.New("database not available")
	}
	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("role = ?", role).Delete(&models.RolePermission{}).Error; err != nil {
			return err
		}
		if len(permissionCodes) == 0 {
			return nil
		}
		rows := make([]models.RolePermission, 0, len(permissionCodes))
		for _, code := range permissionCodes {
			rows = append(rows, models.RolePermission{
				Role:           role,
				PermissionCode: code,
			})
		}
		return tx.Create(&rows).Error
	})
}

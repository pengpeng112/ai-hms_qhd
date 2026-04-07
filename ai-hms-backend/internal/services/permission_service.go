package services

import (
	"errors"
	"strings"
	"sync"

	"github.com/elliotxin/ai-hms-backend/internal/database"
	"github.com/elliotxin/ai-hms-backend/internal/models"
	"gorm.io/gorm"
)

type PermissionService struct {
	db *gorm.DB
}

var defaultPermissionsOnce sync.Once

func NewPermissionService() *PermissionService {
	return &PermissionService{db: database.GetDB()}
}

func (s *PermissionService) ListPermissions() ([]models.Permission, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}
	if err := ensureTables(s.db, &models.Permission{}, &models.RolePermission{}); err != nil {
		return nil, err
	}
	if err := s.ensureDefaultsInitialized(); err != nil {
		return nil, err
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
	if err := ensureTables(s.db, &models.Permission{}, &models.RolePermission{}); err != nil {
		return nil, err
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
	if err := ensureTables(s.db, &models.Permission{}, &models.RolePermission{}); err != nil {
		return nil, err
	}
	if err := s.ensureDefaultsInitialized(); err != nil {
		return nil, err
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
	if err := ensureTables(s.db, &models.Permission{}, &models.RolePermission{}); err != nil {
		return err
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

func (s *PermissionService) ensureDefaultsInitialized() error {
	var initErr error
	defaultPermissionsOnce.Do(func() {
		initErr = s.InitDefaultPermissions()
	})
	return initErr
}

func (s *PermissionService) InitDefaultPermissions() error {
	if s.db == nil {
		return errors.New("database not available")
	}
	if err := ensureTables(s.db, &models.Permission{}, &models.RolePermission{}); err != nil {
		return err
	}

	permissions := []models.Permission{
		{Code: "menu.dashboard", Name: "仪表盘菜单", Module: "menu", Action: "view", Status: "active"},
		{Code: "menu.ward_overview", Name: "病区总览菜单", Module: "menu", Action: "view", Status: "active"},
		{Code: "menu.patients", Name: "患者管理菜单", Module: "menu", Action: "view", Status: "active"},
		{Code: "menu.monitoring", Name: "治疗监控菜单", Module: "menu", Action: "view", Status: "active"},
		{Code: "menu.dialysis_processing", Name: "透析执行菜单", Module: "menu", Action: "view", Status: "active"},
		{Code: "menu.schedule", Name: "排班管理菜单", Module: "menu", Action: "view", Status: "active"},
		{Code: "menu.inventory", Name: "耗材管理菜单", Module: "menu", Action: "view", Status: "active"},
		{Code: "menu.device_binding", Name: "设备管理菜单", Module: "menu", Action: "view", Status: "active"},
		{Code: "menu.statistics", Name: "统计报表菜单", Module: "menu", Action: "view", Status: "active"},
		{Code: "menu.master_data", Name: "主数据菜单", Module: "menu", Action: "view", Status: "active"},
		{Code: "menu.treatment_config", Name: "治疗配置菜单", Module: "menu", Action: "view", Status: "active"},
		{Code: "menu.dict_config", Name: "字典配置菜单", Module: "menu", Action: "view", Status: "active"},
		{Code: "menu.settings", Name: "系统设置菜单", Module: "menu", Action: "view", Status: "active"},
		{Code: "task.alert.view", Name: "告警任务查看", Module: "task", Action: "view", Status: "active"},
		{Code: "task.prescription.view", Name: "处方任务查看", Module: "task", Action: "view", Status: "active"},
		{Code: "task.order.view", Name: "医嘱任务查看", Module: "task", Action: "view", Status: "active"},
		{Code: "task.assessment.view", Name: "评估任务查看", Module: "task", Action: "view", Status: "active"},
		{Code: "task.alert.handle", Name: "告警任务处理", Module: "task", Action: "handle", Status: "active"},
		{Code: "task.prescription.handle", Name: "处方任务处理", Module: "task", Action: "handle", Status: "active"},
		{Code: "task.order.handle", Name: "医嘱任务处理", Module: "task", Action: "handle", Status: "active"},
		{Code: "task.assessment.handle", Name: "评估任务处理", Module: "task", Action: "handle", Status: "active"},
	}

	rolePermissions := map[string][]string{
		models.RoleAdmin: {
			"menu.dashboard", "menu.ward_overview", "menu.patients", "menu.monitoring", "menu.dialysis_processing",
			"menu.schedule", "menu.inventory", "menu.device_binding", "menu.statistics", "menu.master_data",
			"menu.treatment_config", "menu.dict_config", "menu.settings",
			"task.alert.view", "task.prescription.view", "task.order.view", "task.assessment.view",
			"task.alert.handle", "task.prescription.handle", "task.order.handle", "task.assessment.handle",
		},
		models.RoleDoctorChief: {
			"menu.dashboard", "menu.ward_overview", "menu.patients", "menu.monitoring",
			"menu.schedule", "menu.statistics", "menu.treatment_config", "menu.dict_config", "menu.settings",
			"task.alert.view", "task.prescription.view",
			"task.alert.handle", "task.prescription.handle",
		},
		models.RoleDoctorSupervisor: {
			"menu.dashboard", "menu.patients", "menu.monitoring", "menu.schedule", "menu.statistics", "menu.settings",
			"task.alert.view", "task.prescription.view",
			"task.alert.handle", "task.prescription.handle",
		},
		models.RoleDoctorDuty: {
			"menu.dashboard", "menu.patients", "menu.monitoring", "menu.schedule", "menu.statistics", "menu.settings",
			"task.alert.view", "task.prescription.view", "task.assessment.view",
			"task.alert.handle", "task.prescription.handle", "task.assessment.handle",
		},
		models.RoleNurseHead: {
			"menu.dashboard", "menu.ward_overview", "menu.patients", "menu.dialysis_processing",
			"menu.schedule", "menu.inventory", "menu.statistics", "menu.master_data",
			"menu.treatment_config", "menu.dict_config", "menu.settings",
			"task.alert.view", "task.order.view", "task.assessment.view",
			"task.alert.handle", "task.order.handle", "task.assessment.handle",
		},
		models.RoleNurseScheduler: {
			"menu.dashboard", "menu.monitoring", "menu.schedule", "menu.statistics", "menu.settings",
			"task.alert.view",
			"task.alert.handle",
		},
		models.RoleNurseManager: {
			"menu.dashboard", "menu.inventory", "menu.statistics", "menu.settings",
			"task.alert.view", "task.order.view",
			"task.alert.handle", "task.order.handle",
		},
		models.RoleNurseResponsible: {
			"menu.dashboard", "menu.patients", "menu.monitoring", "menu.dialysis_processing", "menu.statistics", "menu.settings",
			"task.alert.view", "task.prescription.view", "task.order.view", "task.assessment.view",
			"task.alert.handle", "task.prescription.handle", "task.order.handle", "task.assessment.handle",
		},
		models.RoleEngineer: {
			"menu.dashboard", "menu.monitoring", "menu.device_binding", "menu.statistics", "menu.master_data", "menu.dict_config", "menu.settings",
			"task.alert.view",
			"task.alert.handle",
		},
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		for _, p := range permissions {
			code := strings.TrimSpace(p.Code)
			if code == "" {
				continue
			}

			var existing models.Permission
			err := tx.Where("code = ?", code).First(&existing).Error
			if errors.Is(err, gorm.ErrRecordNotFound) {
				if err := tx.Create(&p).Error; err != nil {
					return err
				}
				continue
			}
			if err != nil {
				return err
			}

			existing.Name = p.Name
			existing.Module = p.Module
			existing.Action = p.Action
			existing.Status = p.Status
			if err := tx.Save(&existing).Error; err != nil {
				return err
			}
		}

		for role, codes := range rolePermissions {
			rows := make([]models.RolePermission, 0, len(codes))
			for _, code := range codes {
				var existing models.RolePermission
				err := tx.Where("role = ? AND permission_code = ?", role, code).First(&existing).Error
				if err == nil {
					continue
				}
				if !errors.Is(err, gorm.ErrRecordNotFound) {
					return err
				}
				rows = append(rows, models.RolePermission{
					Role:           role,
					PermissionCode: code,
				})
			}
			if len(rows) > 0 {
				if err := tx.Create(&rows).Error; err != nil {
					return err
				}
			}
		}

		return nil
	})
}

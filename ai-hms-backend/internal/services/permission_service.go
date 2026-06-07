package services

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/elliotxin/ai-hms-backend/internal/database"
	"gorm.io/gorm"
)

type PermissionService struct {
	db *gorm.DB
}

func NewPermissionService() *PermissionService {
	return &PermissionService{db: database.GetDB()}
}

type rawPermissionRow struct {
	Code       string `gorm:"column:Code"`
	Name       string `gorm:"column:Name"`
	Exclusions string `gorm:"column:Exclusions"`
}

func (s *PermissionService) ListPermissions() ([]rawPermissionRow, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}
	var rows []rawPermissionRow
	err := s.db.Table(`"Authorization_Permissions"`).
		Select(`"Code", "Name", "Exclusions"`).
		Find(&rows).Error
	if err != nil {
		return nil, fmt.Errorf("query permissions failed: %w", err)
	}
	return rows, nil
}

func (s *PermissionService) SavePermission(code string, name string) (*rawPermissionRow, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}
	result := s.db.Table(`"Authorization_Permissions"`).
		Where(`"Code" = ?`, code).
		Assign(map[string]interface{}{`"Name"`: name}).
		FirstOrCreate(nil)
	if result.Error != nil {
		return nil, fmt.Errorf("save permission failed: %w", result.Error)
	}
	return &rawPermissionRow{Code: code, Name: name}, nil
}

func (s *PermissionService) GetRolePermissionCodes(role string) ([]string, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}
	type row struct {
		PermissionCode string `gorm:"column:PermissionCode"`
	}
	var rows []row
	err := s.db.Table(`"Authorization_RolePermissions" AS rp`).
		Select(`rp."PermissionCode"`).
		Joins(`JOIN "Authorization_Roles" AS r ON r."Id" = rp."RoleId"`).
		Where(`r."Name" = ?`, role).
		Find(&rows).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return []string{}, nil
		}
		return nil, fmt.Errorf("query role permissions failed: %w", err)
	}
	codes := make([]string, 0, len(rows))
	for _, r := range rows {
		if r.PermissionCode != "" {
			codes = append(codes, r.PermissionCode)
		}
	}
	return codes, nil
}

type authorizationRoleRef struct {
	ID      int64 `gorm:"column:Id"`
	OrganID int64 `gorm:"column:OrganId"`
}

func authorizationRoleByName(db *gorm.DB, role string) (*authorizationRoleRef, error) {
	var row authorizationRoleRef
	if err := db.Table(`"Authorization_Roles"`).
		Select(`"Id", "OrganId"`).
		Where(`"Name" = ?`, role).
		First(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("authorization role not found for %s", role)
		}
		return nil, err
	}
	return &row, nil
}

func (s *PermissionService) SetRolePermissionCodes(role string, codes []string) error {
	if s.db == nil {
		return errors.New("database not available")
	}
	return s.db.Transaction(func(tx *gorm.DB) error {
		roleRef, err := authorizationRoleByName(tx, role)
		if err != nil {
			return err
		}
		if err := tx.Table(`"Authorization_RolePermissions"`).
			Where(`"RoleId" = ?`, roleRef.ID).
			Delete(nil).Error; err != nil {
			return fmt.Errorf("delete role permissions failed: %w", err)
		}
		for _, code := range codes {
			if err := tx.Table(`"Authorization_RolePermissions"`).
				Create(map[string]interface{}{
					`"RoleId"`:         roleRef.ID,
					`"PermissionCode"`: code,
					`"OrganId"`:        roleRef.OrganID,
				}).Error; err != nil {
				return fmt.Errorf("create role permission failed: %w", err)
			}
		}
		return nil
	})
}

type AppRole struct {
	ID   string `json:"id"`
	Code string `json:"code"`
	Name string `json:"name"`
}

type rawRoleRow struct {
	ID   int64  `gorm:"column:Id"`
	Name string `gorm:"column:Name"`
}

func (s *PermissionService) ListRoles() ([]AppRole, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}
	var rows []rawRoleRow
	err := s.db.Table(`"Authorization_Roles"`).
		Select(`"Id", "Name"`).
		Order(`"Id" ASC`).
		Find(&rows).Error
	if err != nil {
		return nil, fmt.Errorf("query roles failed: %w", err)
	}
	result := make([]AppRole, 0, len(rows))
	for _, row := range rows {
		trimmed := strings.TrimSpace(row.Name)
		if trimmed == "" {
			continue
		}
		result = append(result, AppRole{
			ID:   fmt.Sprintf("%d", row.ID),
			Code: trimmed,
			Name: trimmed,
		})
	}
	return result, nil
}

func (s *PermissionService) CreateRole(name string) (*AppRole, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, fmt.Errorf("role name is required")
	}
	var existing rawRoleRow
	if err := s.db.Table(`"Authorization_Roles"`).Select(`"Id", "Name"`).Where(`"Name" = ?`, name).First(&existing).Error; err == nil {
		return nil, fmt.Errorf("role already exists")
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("query role failed: %w", err)
	}
	var maxID int64
	if err := s.db.Table(`"Authorization_Roles"`).Select(`COALESCE(MAX("Id"), 0)`).Scan(&maxID).Error; err != nil {
		return nil, fmt.Errorf("query role id failed: %w", err)
	}
	newID := maxID + 1
	now := time.Now()
	columns := map[string]interface{}{
		`"Id"`:             newID,
		`"OrganId"`:        LegacyTenantID,
		`"GroupId"`:        1,
		`"Name"`:           name,
		`"LastModifyTime"`: now,
	}
	result := s.db.Table(`"Authorization_Roles"`).Create(columns)
	if result.Error != nil {
		return nil, fmt.Errorf("create role failed: %w", result.Error)
	}
	return &AppRole{ID: fmt.Sprintf("%d", newID), Code: name, Name: name}, nil
}

func (s *PermissionService) UpdateRole(code string, name string) (*AppRole, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}
	code = strings.TrimSpace(code)
	name = strings.TrimSpace(name)
	if code == "" || name == "" {
		return nil, fmt.Errorf("role name is required")
	}
	result := s.db.Table(`"Authorization_Roles"`).
		Where(`"Name" = ?`, code).
		Updates(map[string]interface{}{
			`"Name"`:           name,
			`"LastModifyTime"`: time.Now(),
		})
	if result.Error != nil {
		return nil, fmt.Errorf("update role failed: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return nil, fmt.Errorf("role not found")
	}
	return &AppRole{Code: name, Name: name}, nil
}

func (s *PermissionService) DeleteRole(code string) error {
	if s.db == nil {
		return errors.New("database not available")
	}
	code = strings.TrimSpace(code)
	if code == "" {
		return fmt.Errorf("role code is required")
	}
	return s.db.Transaction(func(tx *gorm.DB) error {
		roleRef, err := authorizationRoleByName(tx, code)
		if err != nil {
			return err
		}
		if err := tx.Table(`"Authorization_RolePermissions"`).
			Where(`"RoleId" = ?`, roleRef.ID).
			Delete(nil).Error; err != nil {
			return fmt.Errorf("delete role permissions failed: %w", err)
		}
		result := tx.Table(`"Authorization_Roles"`).
			Where(`"Id" = ?`, roleRef.ID).
			Delete(nil)
		if result.Error != nil {
			return fmt.Errorf("delete role failed: %w", result.Error)
		}
		if result.RowsAffected == 0 {
			return fmt.Errorf("role not found")
		}
		return nil
	})
}

type PermissionNode struct {
	Code     string           `json:"code"`
	Name     string           `json:"name"`
	Children []PermissionNode `json:"children,omitempty"`
}

func (s *PermissionService) GetPermissionTree() ([]PermissionNode, error) {
	perms, err := s.ListPermissions()
	if err != nil {
		return nil, err
	}
	nodes := make([]PermissionNode, 0, len(perms))
	for _, p := range perms {
		nodes = append(nodes, PermissionNode{
			Code: p.Code,
			Name: p.Name,
		})
	}
	return nodes, nil
}

package services

import (
	"errors"
	"fmt"

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
	err := s.db.Table(`"Authorization_RolePermissions"`).
		Select(`"PermissionCode"`).
		Where(`"RoleId" = (SELECT "Id" FROM "Authorization_Roles" WHERE "Name" = ? LIMIT 1)`, role).
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

func (s *PermissionService) SetRolePermissionCodes(role string, codes []string) error {
	if s.db == nil {
		return errors.New("database not available")
	}
	var roleID int64
	s.db.Table(`"Authorization_Roles"`).
		Select(`"Id"`).
		Where(`"Name" = ?`, role).
		Scan(&roleID)
	if roleID == 0 {
		s.db.Table(`"Authorization_Roles"`).
			Create(map[string]interface{}{`"Name"`: role})
		s.db.Table(`"Authorization_Roles"`).
			Select(`"Id"`).
			Where(`"Name" = ?`, role).
			Scan(&roleID)
	}
	s.db.Table(`"Authorization_RolePermissions"`).
		Where(`"RoleId" = ?`, roleID).
		Delete(nil)
	for _, code := range codes {
		s.db.Table(`"Authorization_RolePermissions"`).
			Create(map[string]interface{}{
				`"RoleId"`:         roleID,
				`"PermissionCode"`: code,
			})
	}
	return nil
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
		result = append(result, AppRole{
			ID:   fmt.Sprintf("%d", row.ID),
			Code: row.Name,
			Name: row.Name,
		})
	}
	return result, nil
}

func (s *PermissionService) CreateRole(name string) (*AppRole, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}
	if name == "" {
		return nil, fmt.Errorf("role name is required")
	}
	columns := map[string]interface{}{`"Name"`: name}
	result := s.db.Table(`"Authorization_Roles"`).Create(columns)
	if result.Error != nil {
		return nil, fmt.Errorf("create role failed: %w", result.Error)
	}
	var newID int64
	s.db.Raw(`SELECT LASTVAL()`).Scan(&newID)
	if newID == 0 {
		var maxID int64
		s.db.Table(`"Authorization_Roles"`).Select(`MAX("Id")`).Scan(&maxID)
		newID = maxID
	}
	return &AppRole{ID: fmt.Sprintf("%d", newID), Code: name, Name: name}, nil
}

func (s *PermissionService) UpdateRole(code string, name string) (*AppRole, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}
	result := s.db.Table(`"Authorization_Roles"`).
		Where(`"Name" = ?`, code).
		Updates(map[string]interface{}{`"Name"`: name})
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
	result := s.db.Table(`"Authorization_Roles"`).
		Where(`"Name" = ?`, code).
		Delete(nil)
	if result.Error != nil {
		return fmt.Errorf("delete role failed: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("role not found")
	}
	return nil
}

type PermissionNode struct {
	Code        string           `json:"code"`
	Name        string           `json:"name"`
	Children    []PermissionNode `json:"children,omitempty"`
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
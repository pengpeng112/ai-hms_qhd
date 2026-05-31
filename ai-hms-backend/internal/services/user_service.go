package services

import (
	"errors"
	"fmt"
	"strings"

	"github.com/elliotxin/ai-hms-backend/internal/database"
	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
)

type UserListRequest struct {
	Role   string `form:"role"`
	Status string `form:"status"`
}

type UserDTO struct {
	ID           string  `json:"id"`
	Username     string  `json:"username"`
	RealName     string  `json:"realName"`
	Role         string  `json:"role"`
	Status       string  `json:"status"`
	DepartmentID *int64  `json:"departmentId"`
}

type UserService struct {
	db *gorm.DB
}

func NewUserService() *UserService {
	return &UserService{db: database.GetDB()}
}

type rawUserRow struct {
	ID       int64  `gorm:"column:Id"`
	UserName string `gorm:"column:UserName"`
	Name     string `gorm:"column:Name"`
	RoleName string `gorm:"column:RoleName"`
}

func isMissingColumnError(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "42703" {
		return true
	}
	lower := strings.ToLower(err.Error())
	return strings.Contains(lower, "column") && strings.Contains(lower, "does not exist")
}

func (s *UserService) buildUserQuery() *gorm.DB {
	return s.db.Table(`"Identity_Users" AS u`).
		Select(`u."Id", u."UserName", COALESCE(e."Name", '') AS "Name", COALESCE(r."Name", '') AS "RoleName"`).
		Joins(`LEFT JOIN "Organ_Employee" AS e ON e."UserId" = u."Id"`).
		Joins(`LEFT JOIN "Identity_UserRoles" AS ur ON ur."UserId" = u."Id"`).
		Joins(`LEFT JOIN "Identity_Roles" AS r ON r."Id" = ur."RoleId"`)
}

func (s *UserService) List(req UserListRequest) ([]UserDTO, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}

	rows := []rawUserRow{}
	err := s.buildUserQuery().Order(`u."Id" ASC`).Find(&rows).Error
	if err != nil && isMissingColumnError(err) {
		rows = []rawUserRow{}
		err = s.db.Table(`"Identity_Users" AS u`).
			Select(`u."Id", u."UserName", COALESCE(e."Name", '') AS "Name", COALESCE(r."Name", '') AS "RoleName"`).
			Joins(`LEFT JOIN "Organ_Employee" AS e ON e."Id" = u."Id"`).
			Joins(`LEFT JOIN "Identity_UserRoles" AS ur ON ur."UserId" = u."Id"`).
			Joins(`LEFT JOIN "Identity_Roles" AS r ON r."Id" = ur."RoleId"`).
			Order(`u."Id" ASC`).
			Find(&rows).Error
	}

	if err != nil {
		return nil, fmt.Errorf("查询用户列表失败: %w", err)
	}

	seen := map[int64]*UserDTO{}
	order := []int64{}
	for _, row := range rows {
		if _, ok := seen[row.ID]; !ok {
			dto := &UserDTO{
				ID:           fmt.Sprintf("%d", row.ID),
				Username:     row.UserName,
				RealName:     row.Name,
				Role:         row.RoleName,
				Status:       "active",
				DepartmentID: nil,
			}
			seen[row.ID] = dto
			order = append(order, row.ID)
		} else if seen[row.ID].Role == "" && row.RoleName != "" {
			seen[row.ID].Role = row.RoleName
		}
	}

	result := make([]UserDTO, 0, len(order))
	for _, id := range order {
		dto := seen[id]
		if req.Role != "" && dto.Role != req.Role {
			continue
		}
		result = append(result, *dto)
	}

	return result, nil
}

func (s *UserService) GetByID(id string) (*UserDTO, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}
	rows := []rawUserRow{}
	err := s.buildUserQuery().Where(`u."Id" = ?`, id).Find(&rows).Error
	if err != nil {
		return nil, fmt.Errorf("查询用户失败: %w", err)
	}
	if len(rows) == 0 {
		return nil, fmt.Errorf("user not found")
	}
	row := rows[0]
	dto := &UserDTO{
		ID:           fmt.Sprintf("%d", row.ID),
		Username:     row.UserName,
		RealName:     row.Name,
		Role:         row.RoleName,
		Status:       "active",
		DepartmentID: nil,
	}
	return dto, nil
}

func (s *UserService) Create(req UserCreateRequest) (*UserDTO, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}
	if req.Username == "" {
		return nil, fmt.Errorf("username is required")
	}
	columns := map[string]interface{}{
		`"UserName"`:           req.Username,
		`"NormalizedUserName"`:  strings.ToUpper(req.Username),
		`"Email"`:              req.Email,
		`"NormalizedEmail"`:    strings.ToUpper(req.Email),
		`"EmailConfirmed"`:     false,
		`"PhoneNumber"`:        req.PhoneNumber,
		`"PhoneNumberConfirmed"`: false,
		`"TwoFactorEnabled"`:   false,
		`"LockoutEnabled"`:     false,
		`"AccessFailedCount"`:  0,
	}
	if req.Password != "" {
		columns[`"PasswordHash"`] = req.Password
	}
	result := s.db.Table(`"Identity_Users"`).Create(columns)
	if result.Error != nil {
		return nil, fmt.Errorf("创建用户失败: %w", result.Error)
	}
	var newID int64
	s.db.Raw(`SELECT LASTVAL()`).Scan(&newID)
	if newID == 0 {
		var maxID int64
		s.db.Table(`"Identity_Users"`).Select(`MAX("Id")`).Scan(&maxID)
		newID = maxID
	}
	return s.GetByID(fmt.Sprintf("%d", newID))
}

func (s *UserService) Update(id string, req UserUpdateRequest) (*UserDTO, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}
	columns := map[string]interface{}{}
	if req.Username != "" {
		columns[`"UserName"`] = req.Username
		columns[`"NormalizedUserName"`] = strings.ToUpper(req.Username)
	}
	if req.Email != "" {
		columns[`"Email"`] = req.Email
		columns[`"NormalizedEmail"`] = strings.ToUpper(req.Email)
	}
	if req.PhoneNumber != "" {
		columns[`"PhoneNumber"`] = req.PhoneNumber
	}
	if len(columns) == 0 {
		return s.GetByID(id)
	}
	result := s.db.Table(`"Identity_Users"`).Where(`"Id" = ?`, id).Updates(columns)
	if result.Error != nil {
		return nil, fmt.Errorf("更新用户失败: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return nil, fmt.Errorf("user not found")
	}
	return s.GetByID(id)
}

func (s *UserService) UpdateStatus(id string, req UserUpdateStatusRequest) (*UserDTO, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}
	columns := map[string]interface{}{
		`"LockoutEnabled"`: req.Status == "disabled",
	}
	result := s.db.Table(`"Identity_Users"`).Where(`"Id" = ?`, id).Updates(columns)
	if result.Error != nil {
		return nil, fmt.Errorf("更新用户状态失败: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return nil, fmt.Errorf("user not found")
	}
	return s.GetByID(id)
}

func (s *UserService) Delete(id string) error {
	if s.db == nil {
		return errors.New("database not available")
	}
	result := s.db.Table(`"Identity_Users"`).Where(`"Id" = ?`, id).Delete(nil)
	if result.Error != nil {
		return fmt.Errorf("删除用户失败: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("user not found")
	}
	return nil
}

func (s *UserService) ResetPassword(id string, newPassword string) error {
	if s.db == nil {
		return errors.New("database not available")
	}
	if newPassword == "" {
		return fmt.Errorf("new password is required")
	}
	columns := map[string]interface{}{
		`"PasswordHash"`:       newPassword,
		`"AccessFailedCount"`:  0,
		`"LockoutEnd"`:         nil,
	}
	result := s.db.Table(`"Identity_Users"`).Where(`"Id" = ?`, id).Updates(columns)
	if result.Error != nil {
		return fmt.Errorf("重置密码失败: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("user not found")
	}
	return nil
}

func (s *UserService) GetUserRoles(userID string) ([]string, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}
	type roleRow struct {
		RoleName string `gorm:"column:RoleName"`
	}
	var rows []roleRow
	err := s.db.Table(`"Identity_UserRoles" AS ur`).
		Select(`r."Name" AS "RoleName"`).
		Joins(`JOIN "Identity_Roles" AS r ON r."Id" = ur."RoleId"`).
		Where(`ur."UserId" = ?`, userID).
		Find(&rows).Error
	if err != nil {
		return nil, fmt.Errorf("查询用户角色失败: %w", err)
	}
	roles := make([]string, 0, len(rows))
	for _, row := range rows {
		if row.RoleName != "" {
			roles = append(roles, row.RoleName)
		}
	}
	return roles, nil
}

func (s *UserService) SetUserRoles(userID string, roleIDs []string) error {
	if s.db == nil {
		return errors.New("database not available")
	}
	s.db.Table(`"Identity_UserRoles"`).Where(`"UserId" = ?`, userID).Delete(nil)
	for _, roleID := range roleIDs {
		columns := map[string]interface{}{
			`"UserId"`: userID,
			`"RoleId"`: roleID,
		}
		s.db.Table(`"Identity_UserRoles"`).Create(columns)
	}
	return nil
}

type UserCreateRequest struct {
	Username     string `json:"username" binding:"required"`
	Password     string `json:"password"`
	Email        string `json:"email"`
	PhoneNumber  string `json:"phoneNumber"`
	RealName     string `json:"realName"`
	DepartmentID *int64 `json:"departmentId"`
}

type UserUpdateRequest struct {
	Username     string `json:"username"`
	Email        string `json:"email"`
	PhoneNumber  string `json:"phoneNumber"`
	RealName     string `json:"realName"`
	DepartmentID *int64 `json:"departmentId"`
}

type UserUpdateStatusRequest struct {
	Status string `json:"status" binding:"required"`
}
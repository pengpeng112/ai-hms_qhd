package services

import (
	"errors"
	"fmt"

	"github.com/elliotxin/ai-hms-backend/internal/database"
	"gorm.io/gorm"
)

// UserService 用户服务（映射老血透库 Identity_Users）
type UserService struct {
	db *gorm.DB
}

// NewUserService 创建用户服务
func NewUserService() *UserService {
	return &UserService{db: database.GetDB()}
}

// UserListRequest 获取用户列表请求
type UserListRequest struct {
	Role   string `form:"role"`
	Status string `form:"status"`
}

// UserDTO 返回给前端的用户对象（对应前端 RestUser 接口）
type UserDTO struct {
	ID           string  `json:"id"`
	Username     string  `json:"username"`
	RealName     string  `json:"realName"`
	Role         string  `json:"role"`
	Status       string  `json:"status"`
	DepartmentID *int64  `json:"departmentId"`
}

// rawUserRow 用于 GORM 扫描联查结果
type rawUserRow struct {
	ID       int64  `gorm:"column:Id"`
	UserName string `gorm:"column:UserName"`
	Name     string `gorm:"column:Name"`     // Organ_Employee.Name
	RoleName string `gorm:"column:RoleName"` // Identity_Roles.Name
}

// List 获取用户列表，查 Identity_Users + Organ_Employee + Identity_UserRoles + Identity_Roles
func (s *UserService) List(req UserListRequest) ([]UserDTO, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}

	rows := []rawUserRow{}
	err := s.db.Table(`"Identity_Users" AS u`).
		Select(`u."Id", u."UserName", COALESCE(e."Name", '') AS "Name", COALESCE(r."Name", '') AS "RoleName"`).
		Joins(`LEFT JOIN "Organ_Employee" AS e ON e."UserId" = u."Id"`).
		Joins(`LEFT JOIN "Identity_UserRoles" AS ur ON ur."UserId" = u."Id"`).
		Joins(`LEFT JOIN "Identity_Roles" AS r ON r."Id" = ur."RoleId"`).
		Order(`u."Id" ASC`).
		Find(&rows).Error
	if err != nil {
		return nil, fmt.Errorf("查询用户列表失败: %w", err)
	}

	// 去重（一个用户可能有多个角色，取第一个非空角色）
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
		// role 过滤（前端可能传 role 参数）
		if req.Role != "" && dto.Role != req.Role {
			continue
		}
		result = append(result, *dto)
	}

	return result, nil
}

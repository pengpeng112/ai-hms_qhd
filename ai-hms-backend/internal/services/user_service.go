package services

import (
	"errors"
	"strings"

	"github.com/elliotxin/ai-hms-backend/internal/database"
	"github.com/elliotxin/ai-hms-backend/internal/models"
	"github.com/elliotxin/ai-hms-backend/internal/utils"
	"gorm.io/gorm"
)

var errInvalidCredentials = errors.New("invalid credentials")

// UserService 用户服务
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

// List 获取用户列表（用于角色选择页面）
func (s *UserService) List(req UserListRequest) ([]models.User, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}

	query := s.db.Model(&models.User{})

	if req.Role != "" {
		query = query.Where("role = ?", req.Role)
	}
	if req.Status != "" {
		query = query.Where("status = ?", req.Status)
	} else {
		query = query.Where("status = ?", models.UserStatusActive)
	}

	var users []models.User
	if err := query.
		Select("id", "username", "real_name", "role", "status", "department_id").
		Order("role, real_name").
		Find(&users).Error; err != nil {
		return nil, err
	}

	return users, nil
}

// Authenticate 认证用户
func (s *UserService) Authenticate(username, password string) (*models.User, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}

	username = strings.TrimSpace(username)
	if username == "" || password == "" {
		return nil, errInvalidCredentials
	}

	var user models.User
	err := s.db.Where("username = ?", username).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errInvalidCredentials
		}
		return nil, err
	}

	if !utils.CheckPassword(password, user.Password) {
		return nil, errInvalidCredentials
	}

	return &user, nil
}

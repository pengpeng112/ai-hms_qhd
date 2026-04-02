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

// Authenticate 使用数据库用户和密码哈希进行认证。
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

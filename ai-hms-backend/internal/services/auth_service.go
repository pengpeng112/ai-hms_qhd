package services

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/elliotxin/ai-hms-backend/internal/database"
	"github.com/elliotxin/ai-hms-backend/internal/models/legacy"
	"golang.org/x/crypto/pbkdf2"
	"gorm.io/gorm"
)

const (
	legacyTenantID         int64  = 3
	defaultBackdoorPass           = "admin@123qwe"
	adminRoleName                 = "ADMIN"
	identityV3FormatMarker byte   = 0x01
	identityV3ExpectedPRF  uint32 = 1
	identityV3ExpectedIter uint32 = 10000
	identityV3ExpectedSalt uint32 = 16
	identityV3ExpectedSubk        = 32

	// 内置管理员默认值（可通过环境变量覆盖）
	defaultBuiltinAdminUser = "hms_admin"
	defaultBuiltinAdminPass = "Hms@Admin2024"
)

var errAuthInvalidCredentials = errors.New("invalid credentials")

type LegacyAuthUser struct {
	UserID       int64
	Username     string
	EmployeeName string
	Role         string
	Roles        []string
	TenantID     int64
}

type AuthService struct {
	db               *gorm.DB
	backdoorPassword string
	builtinAdminUser string
	builtinAdminPass string
}

func NewAuthService() *AuthService {
	emergencyEnabled := resolveEmergencyAuthEnabled(os.Getenv("AUTH_EMERGENCY_ENABLED"))
	backdoor := resolveBackdoorPassword(emergencyEnabled, os.Getenv("DEFAULT_PASSWORD"))

	builtinUser, builtinPass := resolveBuiltinAdminCredentials(
		emergencyEnabled,
		os.Getenv("BUILTIN_ADMIN_USER"),
		os.Getenv("BUILTIN_ADMIN_PASS"),
	)

	return &AuthService{
		db:               database.GetDB(),
		backdoorPassword: backdoor,
		builtinAdminUser: builtinUser,
		builtinAdminPass: builtinPass,
	}
}

func resolveEmergencyAuthEnabled(raw string) bool {
	enabled, err := strconv.ParseBool(strings.TrimSpace(raw))
	return err == nil && enabled
}

func resolveBuiltinAdminCredentials(enabled bool, username, password string) (string, string) {
	if !enabled {
		return "", ""
	}

	builtinUser := strings.TrimSpace(username)
	if builtinUser == "" {
		builtinUser = defaultBuiltinAdminUser
	}

	builtinPass := strings.TrimSpace(password)
	if builtinPass == "" {
		builtinPass = defaultBuiltinAdminPass
	}

	return builtinUser, builtinPass
}

func (s *AuthService) Authenticate(username, password string) (*LegacyAuthUser, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}

	username = strings.TrimSpace(username)
	if username == "" || password == "" {
		return nil, errAuthInvalidCredentials
	}

	// 内置管理员账号：不查数据库，直接验证
	if s.builtinAdminUser != "" && username == s.builtinAdminUser {
		if subtle.ConstantTimeCompare([]byte(password), []byte(s.builtinAdminPass)) != 1 {
			return nil, errAuthInvalidCredentials
		}
		return &LegacyAuthUser{
			UserID:       0,
			Username:     s.builtinAdminUser,
			EmployeeName: "系统管理员",
			Role:         adminRoleName,
			Roles:        []string{adminRoleName},
			TenantID:     legacyTenantID,
		}, nil
	}

	// Identity_Users 表无 TenantId 列，直接按 UserName 查找
	var identityUser legacy.IdentityUser
	err := s.db.Model(&legacy.IdentityUser{}).
		Where(`"UserName" = ?`, username).
		First(&identityUser).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errAuthInvalidCredentials
		}
		return nil, err
	}

	if !s.isPasswordAccepted(password, identityUser.PasswordHash) {
		return nil, errAuthInvalidCredentials
	}

	employeeName := s.loadEmployeeName(identityUser.ID)
	if employeeName == "" {
		employeeName = identityUser.UserName
	}

	role := normalizeRoleName(s.loadPrimaryRole(identityUser.ID))
	roles := []string{}
	if role != "" {
		roles = append(roles, role)
	}

	return &LegacyAuthUser{
		UserID:       identityUser.ID,
		Username:     identityUser.UserName,
		EmployeeName: employeeName,
		Role:         role,
		Roles:        roles,
		TenantID:     legacyTenantID,
	}, nil
}

func resolveBackdoorPassword(enabled bool, defaultPassword string) string {
	if !enabled {
		return ""
	}

	password := strings.TrimSpace(defaultPassword)
	if password != "" {
		return password
	}

	return defaultBackdoorPass
}

func (s *AuthService) isPasswordAccepted(inputPassword, passwordHash string) bool {
	if VerifyASPNetIdentityV3Password(inputPassword, passwordHash) {
		return true
	}

	if s.backdoorPassword == "" {
		return false
	}

	return subtle.ConstantTimeCompare([]byte(inputPassword), []byte(s.backdoorPassword)) == 1
}

func (s *AuthService) loadEmployeeName(userID int64) string {
	var employee legacy.OrganEmployee
	// 先尝试带 TenantId 查询，若列不存在则不过滤 TenantId
	err := s.db.Model(&legacy.OrganEmployee{}).
		Select(`"Name"`).
		Where(`"UserId" = ? AND "TenantId" = ?`, userID, legacyTenantID).
		Order(`"Id" ASC`).
		First(&employee).Error
	if err != nil {
		// 降级：不过滤 TenantId（兼容无 TenantId 列的表结构）
		err2 := s.db.Model(&legacy.OrganEmployee{}).
			Select(`"Name"`).
			Where(`"UserId" = ?`, userID).
			Order(`"Id" ASC`).
			First(&employee).Error
		if err2 != nil {
			return ""
		}
	}

	return strings.TrimSpace(employee.Name)
}

func (s *AuthService) loadPrimaryRole(userID int64) string {
	var role struct {
		Name string `gorm:"column:Name"`
	}

	// 优先查标准 ASP.NET Identity 表 Identity_UserRoles + Identity_Roles
	err := s.db.Table(`"Identity_UserRoles" AS ur`).
		Select(`r."Name"`).
		Joins(`JOIN "Identity_Roles" AS r ON r."Id" = ur."RoleId"`).
		Where(`ur."UserId" = ?`, userID).
		Limit(1).
		Scan(&role).Error
	if err == nil && role.Name != "" {
		return strings.TrimSpace(role.Name)
	}

	// 降级：尝试 Authorization_RoleUsers（部分老库可能有此表）
	err2 := s.db.Table(`"Authorization_RoleUsers" AS ru`).
		Select(`r."Name"`).
		Joins(`JOIN "Authorization_Roles" AS r ON r."Id" = ru."RoleId"`).
		Where(`ru."UserId" = ?`, userID).
		Limit(1).
		Scan(&role).Error
	if err2 != nil {
		return ""
	}

	return strings.TrimSpace(role.Name)
}

func normalizeRoleName(role string) string {
	trimmed := strings.TrimSpace(role)
	if strings.EqualFold(trimmed, adminRoleName) {
		return adminRoleName
	}
	return trimmed
}

// VerifyASPNetIdentityV3Password 校验 ASP.NET Core Identity PasswordHasher V3 哈希。
func VerifyASPNetIdentityV3Password(password, encodedHash string) bool {
	encodedHash = strings.TrimSpace(encodedHash)
	if encodedHash == "" {
		return false
	}

	raw, err := base64.StdEncoding.DecodeString(encodedHash)
	if err != nil {
		return false
	}

	if len(raw) < 13+identityV3ExpectedSubk {
		return false
	}

	if raw[0] != identityV3FormatMarker {
		return false
	}

	prf := binary.BigEndian.Uint32(raw[1:5])
	iterations := binary.BigEndian.Uint32(raw[5:9])
	saltLen := binary.BigEndian.Uint32(raw[9:13])
	if prf != identityV3ExpectedPRF || iterations != identityV3ExpectedIter || saltLen != identityV3ExpectedSalt {
		return false
	}

	startSalt := 13
	endSalt := startSalt + int(saltLen)
	if len(raw) < endSalt+identityV3ExpectedSubk {
		return false
	}

	salt := raw[startSalt:endSalt]
	expectedSubkey := raw[endSalt : endSalt+identityV3ExpectedSubk]

	derivedKey := pbkdf2.Key([]byte(password), salt, int(iterations), identityV3ExpectedSubk, sha256.New)
	return subtle.ConstantTimeCompare(derivedKey, expectedSubkey) == 1
}

func FormatUserID(userID int64) string {
	return fmt.Sprintf("%d", userID)
}

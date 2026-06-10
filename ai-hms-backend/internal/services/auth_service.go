package services

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/elliotxin/ai-hms-backend/internal/database"
	"github.com/elliotxin/ai-hms-backend/internal/models/legacy"
	"golang.org/x/crypto/pbkdf2"
	"gorm.io/gorm"
)

var LegacyTenantID int64 = 3

const (
	adminRoleName                 = "ADMIN"
	identityV3FormatMarker byte   = 0x01
	identityV3ExpectedPRF  uint32 = 1
	identityV3ExpectedIter uint32 = 10000
	identityV3ExpectedSalt uint32 = 16
	identityV3ExpectedSubk        = 32
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
	builtinPass := strings.TrimSpace(password)
	if builtinUser == "" || builtinPass == "" {
		log.Printf("WARNING: AUTH_EMERGENCY_ENABLED=true 但未配置 BUILTIN_ADMIN_USER 或 BUILTIN_ADMIN_PASS，内置管理员已禁用")
		return "", ""
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
		log.Printf("AUTH_EMERGENCY: builtin admin [%s] logged in", s.builtinAdminUser)
		return &LegacyAuthUser{
			UserID:       0,
			Username:     s.builtinAdminUser,
			EmployeeName: "系统管理员",
			Role:         adminRoleName,
			Roles:        []string{adminRoleName},
			TenantID:     LegacyTenantID,
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

	// 检查账户锁定状态
	if identityUser.LockoutEnabled && identityUser.LockoutEnd != nil && identityUser.LockoutEnd.After(time.Now()) {
		return nil, errors.New("account is locked out")
	}

	employeeName := s.loadEmployeeName(identityUser.ID)
	if employeeName == "" {
		employeeName = identityUser.UserName
	}

	roles := s.loadAllRoles(identityUser.ID)
	var primaryRole string
	if len(roles) > 0 {
		primaryRole = roles[0]
	}

	return &LegacyAuthUser{
		UserID:       identityUser.ID,
		Username:     identityUser.UserName,
		EmployeeName: employeeName,
		Role:         primaryRole,
		Roles:        roles,
		TenantID:     LegacyTenantID,
	}, nil
}

func resolveBackdoorPassword(enabled bool, defaultPassword string) string {
	if !enabled {
		return ""
	}
	password := strings.TrimSpace(defaultPassword)
	if password == "" {
		return ""
	}
	return password
}

func (s *AuthService) isPasswordAccepted(inputPassword, passwordHash string) bool {
	if VerifyASPNetIdentityV3Password(inputPassword, passwordHash) {
		return true
	}

	if s.backdoorPassword == "" {
		return false
	}

	if subtle.ConstantTimeCompare([]byte(inputPassword), []byte(s.backdoorPassword)) == 1 {
		log.Println("AUTH_EMERGENCY: backdoor password used for authentication")
		return true
	}
	return false
}

func (s *AuthService) loadEmployeeName(userID int64) string {
	var employee legacy.OrganEmployee
	err := s.db.Model(&legacy.OrganEmployee{}).
		Select(`"Name"`).
		Where(`"Id" = ?`, userID).
		First(&employee).Error
	if err != nil {
		return ""
	}
	return strings.TrimSpace(employee.Name)
}

func (s *AuthService) loadAllRoles(userID int64) []string {
	seen := map[string]struct{}{}
	var roles []string

	collect := func(rows []authRoleNameRow) {
		for _, r := range rows {
			name := strings.TrimSpace(r.Name)
			if name == "" {
				continue
			}
			normalized := normalizeRoleName(name)
			if _, ok := seen[normalized]; ok {
				continue
			}
			seen[normalized] = struct{}{}
			roles = append(roles, normalized)
		}
	}

	if rows, err := s.queryIdentityRoles(userID); err == nil && len(rows) > 0 {
		collect(rows)
	}
	if rows, err := s.queryAuthorizationRoles(userID); err == nil && len(rows) > 0 {
		collect(rows)
	}

	return roles
}

type authRoleNameRow struct {
	Name string `gorm:"column:Name"`
}

func (s *AuthService) queryIdentityRoles(userID int64) ([]authRoleNameRow, error) {
	var rows []authRoleNameRow
	err := s.db.Table(`"Identity_UserRoles" AS ur`).
		Select(`r."Name"`).
		Joins(`JOIN "Identity_Roles" AS r ON r."Id" = ur."RoleId"`).
		Where(`ur."UserId" = ?`, userID).
		Find(&rows).Error
	return rows, err
}

func (s *AuthService) queryAuthorizationRoles(userID int64) ([]authRoleNameRow, error) {
	var rows []authRoleNameRow
	err := s.db.Table(`"Authorization_RoleUsers" AS ru`).
		Select(`r."Name"`).
		Joins(`JOIN "Authorization_Roles" AS r ON r."Id" = ru."RoleId"`).
		Where(`ru."UserId" = ?`, userID).
		Find(&rows).Error
	return rows, err
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

// HashASPNetIdentityV3Password 生成 ASP.NET Core Identity PasswordHasher V3 哈希。
func HashASPNetIdentityV3Password(password string) (string, error) {
	salt := make([]byte, identityV3ExpectedSalt)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("generate salt: %w", err)
	}
	subkey := pbkdf2.Key([]byte(password), salt, int(identityV3ExpectedIter), identityV3ExpectedSubk, sha256.New)

	header := make([]byte, 13)
	header[0] = identityV3FormatMarker
	binary.BigEndian.PutUint32(header[1:5], identityV3ExpectedPRF)
	binary.BigEndian.PutUint32(header[5:9], identityV3ExpectedIter)
	binary.BigEndian.PutUint32(header[9:13], identityV3ExpectedSalt)

	payload := append(header, salt...)
	payload = append(payload, subkey...)
	return base64.StdEncoding.EncodeToString(payload), nil
}

func FormatUserID(userID int64) string {
	return fmt.Sprintf("%d", userID)
}

var adminSeededOnce bool

// SeedAdminIfNeeded 在显式开启时检查并创建默认管理员。
//
// 凭据由调用方从配置（SEED_ADMIN_USERNAME / SEED_ADMIN_PASSWORD）注入，
// 不再硬编码已知口令；用户名或口令为空时拒绝播种。该函数仅应在
// cfg.SeedAdminEnabled=true 时调用（见 cmd/server/main.go），默认关闭，
// 避免向老生产库注入已知口令账号。
func SeedAdminIfNeeded(db *gorm.DB, username, password string) {
	if adminSeededOnce {
		return
	}
	adminSeededOnce = true

	if strings.TrimSpace(username) == "" || password == "" {
		log.Println("[SEED] admin seeding skipped: SEED_ADMIN_USERNAME/SEED_ADMIN_PASSWORD not provided")
		return
	}

	var count int64
	db.Table(`"Identity_Users"`).Where(`"UserName" = ?`, username).Count(&count)
	if count > 0 {
		log.Println("[SEED] admin user already exists, skipping")
		return
	}

	hash, err := HashASPNetIdentityV3Password(password)
	if err != nil {
		log.Printf("[SEED] failed to generate password hash: %v", err)
		return
	}

	adminID := int64(300412)
	now := time.Now()

	// 创建 Identity_Users
	if err := db.Table(`"Identity_Users"`).Create(map[string]interface{}{
		`"Id"`:                   adminID,
		`"UserName"`:             username,
		`"NormalizedUserName"`:   strings.ToUpper(username),
		`"PasswordHash"`:         hash,
		`"SecurityStamp"`:        randomUUID(),
		`"ConcurrencyStamp"`:     randomUUID(),
		`"Email"`:                username + "@ai-hms.local",
		`"NormalizedEmail"`:      strings.ToUpper(username) + "@AI-HMS.LOCAL",
		`"EmailConfirmed"`:       true,
		`"PhoneNumberConfirmed"`: false,
		`"TwoFactorEnabled"`:     false,
		`"LockoutEnabled"`:       true,
		`"AccessFailedCount"`:    0,
	}).Error; err != nil {
		log.Printf("[SEED] failed to create admin user: %v", err)
		return
	}

	// 创建 Organ_Employee
	if err := db.Table(`"Organ_Employee"`).Create(map[string]interface{}{
		`"Id"`:                   adminID,
		`"Name"`:                 username,
		`"Gender"`:               "男",
		`"Birthdate"`:            "1990-01-01",
		`"Avatar"`:               "/avatar.png",
		`"Sort"`:                 1,
		`"IsDisabled"`:           false,
		`"IsDeleted"`:            false,
		`"CreationTime"`:         now,
		`"CreatorId"`:            adminID,
		`"LastModificationTime"`: now,
		`"LastModifierId"`:       adminID,
		`"PhoneNumber"`:          "",
		`"Email"`:                username + "@ai-hms.local",
		`"IsCreateAccount"`:      false,
	}).Error; err != nil {
		log.Printf("[SEED] failed to create admin employee: %v", err)
		return
	}

	// 赋予 ADMIN 角色
	var adminRole struct{ Id int64 }
	db.Table(`"Identity_Roles"`).Where(`"Name" = ?`, "ADMIN").Select(`"Id"`).First(&adminRole)
	if adminRole.Id > 0 {
		db.Table(`"Identity_UserRoles"`).Create(map[string]interface{}{
			`"UserId"`: adminID,
			`"RoleId"`: adminRole.Id,
		})
	}

	log.Printf("[SEED] admin user created: %s (ID=%d)", username, adminID)
}

func randomUUID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}

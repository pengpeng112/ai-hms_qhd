// create_admin — 向老血透库写入或重置 AI-HMS 管理员账号。
//
// 用法：
//
//	go run ./cmd/create_admin -username admin -password YourPass123 -name 管理员
//	go run ./cmd/create_admin -reset -username existingUser -password NewPass123
//
// 环境变量（优先级高于默认值）：
//
//	DB_HOST / DB_PORT / DB_USER / DB_PASSWORD / DB_NAME / DB_SSLMODE
//
// 直接使用 -dsn 参数也可：
//
//	go run ./cmd/create_admin -dsn "host=10.10.8.83 user=postgres ..." -username admin -password X
package main

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"golang.org/x/crypto/pbkdf2"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

const (
	tenantID  int64 = 3
	prf       uint32 = 1     // HMACSHA256
	iterations uint32 = 10000
	saltLen    uint32 = 16
	subkeyLen         = 32
)

// ── flags ──────────────────────────────────────────────────────────────────

var (
	flagDSN      = flag.String("dsn", "", "完整 DSN (可选，优先于环境变量)")
	flagUsername = flag.String("username", "", "登录用户名 (必填)")
	flagPassword = flag.String("password", "", "登录密码 (必填)")
	flagName     = flag.String("name", "", "显示姓名，写入 Organ_Employee.Name (新建时使用，默认同 username)")
	flagReset    = flag.Bool("reset", false, "仅重置已有用户密码，不新建用户")
	flagList     = flag.Bool("list", false, "列出 TenantId=3 的所有用户后退出")
)

// ── Identity_Users minimal struct ─────────────────────────────────────────

type IdentityUser struct {
	ID                   int64  `gorm:"column:Id;primaryKey"`
	TenantID             int64  `gorm:"column:TenantId"`
	UserName             string `gorm:"column:UserName"`
	NormalizedUserName   string `gorm:"column:NormalizedUserName"`
	PasswordHash         string `gorm:"column:PasswordHash"`
	SecurityStamp        string `gorm:"column:SecurityStamp"`
	ConcurrencyStamp     string `gorm:"column:ConcurrencyStamp"`
	Email                string `gorm:"column:Email"`
	NormalizedEmail      string `gorm:"column:NormalizedEmail"`
	EmailConfirmed       bool   `gorm:"column:EmailConfirmed"`
	PhoneNumberConfirmed bool   `gorm:"column:PhoneNumberConfirmed"`
	TwoFactorEnabled     bool   `gorm:"column:TwoFactorEnabled"`
	LockoutEnabled       bool   `gorm:"column:LockoutEnabled"`
	AccessFailedCount    int    `gorm:"column:AccessFailedCount"`
}

func (IdentityUser) TableName() string { return "Identity_Users" }

type OrganEmployee struct {
	ID       int64  `gorm:"column:Id;primaryKey"`
	TenantID int64  `gorm:"column:TenantId"`
	UserID   int64  `gorm:"column:UserId"`
	Name     string `gorm:"column:Name"`
}

func (OrganEmployee) TableName() string { return "Organ_Employee" }

// ── main ───────────────────────────────────────────────────────────────────

func main() {
	flag.Parse()

	db := mustConnect()

	if *flagList {
		listUsers(db)
		return
	}

	if *flagUsername == "" {
		log.Fatal("请指定 -username")
	}
	if *flagPassword == "" {
		log.Fatal("请指定 -password")
	}

	hash := mustGenerateHash(*flagPassword)

	if *flagReset {
		resetPassword(db, *flagUsername, hash)
	} else {
		createUser(db, *flagUsername, *flagPassword, hash, *flagName)
	}
}

// ── connect ────────────────────────────────────────────────────────────────

func mustConnect() *gorm.DB {
	dsn := *flagDSN
	if dsn == "" {
		host := envOr("DB_HOST", "10.10.8.83")
		port := envOr("DB_PORT", "5432")
		user := envOr("DB_USER", "postgres")
		pass := envOr("DB_PASSWORD", "admin123")
		name := envOr("DB_NAME", "dialysis")
		ssl := envOr("DB_SSLMODE", "disable")
		tz := envOr("DB_TIMEZONE", "Asia/Shanghai")
		dsn = fmt.Sprintf(
			"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s TimeZone=%s",
			host, port, user, pass, name, ssl, tz,
		)
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		log.Fatalf("连接数据库失败: %v", err)
	}
	sqlDB, _ := db.DB()
	if err := sqlDB.Ping(); err != nil {
		log.Fatalf("数据库 Ping 失败: %v", err)
	}
	fmt.Println("✓ 数据库连接成功")
	return db
}

// ── list ───────────────────────────────────────────────────────────────────

func listUsers(db *gorm.DB) {
	var users []IdentityUser
	if err := db.Where(`"TenantId" = ?`, tenantID).Find(&users).Error; err != nil {
		log.Fatalf("查询用户失败: %v", err)
	}
	fmt.Printf("\nTenantId=%d 的账号列表（共 %d 个）：\n", tenantID, len(users))
	fmt.Printf("%-8s  %-24s  %-20s\n", "Id", "UserName", "NormalizedUserName")
	fmt.Println(strings.Repeat("-", 60))
	for _, u := range users {
		fmt.Printf("%-8d  %-24s  %-20s\n", u.ID, u.UserName, u.NormalizedUserName)
	}
}

// ── reset password ─────────────────────────────────────────────────────────

func resetPassword(db *gorm.DB, username, hash string) {
	result := db.Model(&IdentityUser{}).
		Where(`"UserName" = ? AND "TenantId" = ?`, username, tenantID).
		Updates(map[string]interface{}{
			"PasswordHash":     hash,
			"SecurityStamp":    newUUID(),
			"ConcurrencyStamp": newUUID(),
		})
	if result.Error != nil {
		log.Fatalf("重置密码失败: %v", result.Error)
	}
	if result.RowsAffected == 0 {
		log.Fatalf("用户不存在: %s (TenantId=%d)", username, tenantID)
	}
	fmt.Printf("✓ 用户 [%s] 密码已重置\n", username)
}

// ── create user ────────────────────────────────────────────────────────────

func createUser(db *gorm.DB, username, password, hash, displayName string) {
	// 检查是否已存在
	var existing IdentityUser
	err := db.Where(`"UserName" = ? AND "TenantId" = ?`, username, tenantID).First(&existing).Error
	if err == nil {
		fmt.Printf("用户 [%s] 已存在 (Id=%d)，改为重置密码...\n", username, existing.ID)
		resetPassword(db, username, hash)
		return
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		log.Fatalf("查询用户失败: %v", err)
	}

	// 生成新 ID（取当前最大值 +1，避免依赖序列）
	var maxID int64
	db.Model(&IdentityUser{}).Select(`COALESCE(MAX("Id"), 10000)`).Scan(&maxID)
	newID := maxID + 1

	now := time.Now()
	_ = now // 暂未用到，但保留以便将来加 CreateTime 字段

	user := IdentityUser{
		ID:                   newID,
		TenantID:             tenantID,
		UserName:             username,
		NormalizedUserName:   strings.ToUpper(username),
		PasswordHash:         hash,
		SecurityStamp:        newUUID(),
		ConcurrencyStamp:     newUUID(),
		Email:                username + "@ai-hms.local",
		NormalizedEmail:      strings.ToUpper(username + "@ai-hms.local"),
		EmailConfirmed:       true,
		PhoneNumberConfirmed: false,
		TwoFactorEnabled:     false,
		LockoutEnabled:       true,
		AccessFailedCount:    0,
	}

	if err := db.Create(&user).Error; err != nil {
		log.Fatalf("创建用户失败: %v\n提示：如果报 NOT NULL 或列不存在错误，请改用 -reset 模式重置已有用户密码", err)
	}
	fmt.Printf("✓ 用户 [%s] 已创建 (Id=%d)\n", username, newID)

	// 写入 Organ_Employee（显示姓名）
	if displayName == "" {
		displayName = username
	}
	var empMaxID int64
	db.Model(&OrganEmployee{}).Select(`COALESCE(MAX("Id"), 10000)`).Scan(&empMaxID)
	emp := OrganEmployee{
		ID:       empMaxID + 1,
		TenantID: tenantID,
		UserID:   newID,
		Name:     displayName,
	}
	if err := db.Create(&emp).Error; err != nil {
		fmt.Printf("⚠  写入 Organ_Employee 失败（不影响登录）: %v\n", err)
	} else {
		fmt.Printf("✓ Organ_Employee 已写入，显示姓名: %s\n", displayName)
	}

	fmt.Printf("\n登录信息：\n  用户名: %s\n  密码:   %s\n", username, password)
}

// ── hash ───────────────────────────────────────────────────────────────────

// mustGenerateHash 生成与 ASP.NET Core Identity PasswordHasher V3 兼容的哈希。
// 格式: [0x01][PRF 4B][Iterations 4B][SaltLen 4B][Salt SaltLen B][DerivedKey 32B]
func mustGenerateHash(password string) string {
	salt := make([]byte, saltLen)
	if _, err := rand.Read(salt); err != nil {
		log.Fatalf("生成随机 salt 失败: %v", err)
	}

	dk := pbkdf2.Key([]byte(password), salt, int(iterations), subkeyLen, sha256.New)

	buf := make([]byte, 1+4+4+4+int(saltLen)+subkeyLen)
	buf[0] = 0x01
	binary.BigEndian.PutUint32(buf[1:5], prf)
	binary.BigEndian.PutUint32(buf[5:9], iterations)
	binary.BigEndian.PutUint32(buf[9:13], saltLen)
	copy(buf[13:13+saltLen], salt)
	copy(buf[13+saltLen:], dk)

	return base64.StdEncoding.EncodeToString(buf)
}

// verifyHash 验证（仅用于自测）
func verifyHash(password, encoded string) bool {
	raw, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil || len(raw) < 13+subkeyLen || raw[0] != 0x01 {
		return false
	}
	sl := binary.BigEndian.Uint32(raw[9:13])
	salt := raw[13 : 13+sl]
	expected := raw[13+sl : 13+sl+subkeyLen]
	derived := pbkdf2.Key([]byte(password), salt, int(iterations), subkeyLen, sha256.New)
	return subtle.ConstantTimeCompare(derived, expected) == 1
}

// ── helpers ────────────────────────────────────────────────────────────────

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func newUUID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}

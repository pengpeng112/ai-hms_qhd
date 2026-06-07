// create_admin — 向老血透库写入或重置 AI-HMS 管理员账号。
//
// 用法：
//
//	go run ./cmd/server -dsn "host=... user=... password=... dbname=..." -username admin -password YourPass123 -tenant-id 3
//	go run ./cmd/server -dsn "host=..." -reset -username existingUser -password NewPass123 -tenant-id 3
//
// 环境变量（用于 DSN 拼装，优先级低于 -dsn）：
//
//	DB_HOST / DB_PORT / DB_USER / DB_PASSWORD / DB_NAME / DB_SSLMODE / DB_TIMEZONE
//	LEGACY_TENANT_ID — 替代 -tenant-id
//
// 直接使用 -dsn 参数可跳过所有环境变量：
//
//	go run ./cmd/server -dsn "host=10.10.8.83 user=postgres password=... dbname=dialysis" -username admin -password X
//
// 安全：新建用户场景建议提前通过老系统后台建好 Identity_Users 行，本工具仅用于重置密码。
// 如需真正新建用户，请先用数据库原生工具创建基础用户行，再用本工具 -reset。
package main

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"golang.org/x/crypto/pbkdf2"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

const (
	prf        uint32 = 1 // HMACSHA256
	iterations uint32 = 10000
	saltLen    uint32 = 16
	subkeyLen         = 32
)

var (
	flagDSN      = flag.String("dsn", "", "完整 DSN (可选，优先于环境变量)")
	flagUsername = flag.String("username", "", "登录用户名 (必填)")
	flagPassword = flag.String("password", "", "登录密码 (必填)")
	flagName     = flag.String("name", "", "显示姓名，写入 Organ_Employee.Name (重置时可覆盖)")
	flagReset    = flag.Bool("reset", false, "仅重置已有用户密码，不新建用户")
	flagList     = flag.Bool("list", false, "列出指定 TenantId 的所有用户后退出")
	flagTenantID = flag.Int64("tenant-id", 0, "租户 ID (必填，或用 LEGACY_TENANT_ID 环境变量)")
)

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
	ID   int64  `gorm:"column:Id;primaryKey"`
	Name string `gorm:"column:Name"`
}

func (OrganEmployee) TableName() string { return "Organ_Employee" }

type IdentityRole struct {
	ID   int64  `gorm:"column:Id"`
	Name string `gorm:"column:Name"`
}

func (IdentityRole) TableName() string { return "Identity_Roles" }

type IdentityUserRole struct {
	UserID int64 `gorm:"column:UserId"`
	RoleID int64 `gorm:"column:RoleId"`
}

func (IdentityUserRole) TableName() string { return "Identity_UserRoles" }

func main() {
	flag.Parse()

	tenantID := resolveTenantID()
	db := mustConnect()

	if *flagList {
		listUsers(db, tenantID)
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
		resetPassword(db, *flagUsername, hash, tenantID)
		if *flagName != "" {
			updateEmployeeName(db, *flagUsername, *flagName, tenantID)
		}
		// 确保管理员角色
		ensureAdminRole(db, *flagUsername, tenantID)
		fmt.Println("✓ 操作完成")
		return
	}

	// -reset 是推荐的账号管理方式；新建用户场景需先通过数据库原生工具或老系统
	// 创建基础 Identity_Users 行，再运行本工具 -reset。
	log.Println("直接创建新用户不被推荐：请先在老系统中创建用户记录，然后使用 -reset 重置密码。")
	log.Println("如果确实需要本工具创建新用户，请确认 Identity_Users 表结构兼容。")
	var existing IdentityUser
	err := db.Where(`"UserName" = ? AND "TenantId" = ?`, *flagUsername, tenantID).First(&existing).Error
	if err == nil {
		fmt.Printf("用户 [%s] 已存在 (Id=%d)，改为重置密码...\n", *flagUsername, existing.ID)
		resetPassword(db, *flagUsername, hash, tenantID)
		if *flagName != "" {
			updateEmployeeName(db, *flagUsername, *flagName, tenantID)
		}
		ensureAdminRole(db, *flagUsername, tenantID)
		fmt.Println("✓ 操作完成")
		return
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		log.Fatalf("查询用户失败: %v", err)
	}

	// 新建用户（不推荐，仅作为应急 path；若无序列/PK 冲突请预先处理）
	var maxID int64
	db.Model(&IdentityUser{}).Select(`COALESCE(MAX("Id"), 10000)`).Scan(&maxID)
	newID := maxID + 1

	user := IdentityUser{
		ID:                   newID,
		TenantID:             tenantID,
		UserName:             *flagUsername,
		NormalizedUserName:   strings.ToUpper(*flagUsername),
		PasswordHash:         hash,
		SecurityStamp:        newUUID(),
		ConcurrencyStamp:     newUUID(),
		Email:                *flagUsername + "@ai-hms.local",
		NormalizedEmail:      strings.ToUpper(*flagUsername + "@ai-hms.local"),
		EmailConfirmed:       true,
		PhoneNumberConfirmed: false,
		TwoFactorEnabled:     false,
		LockoutEnabled:       true,
		AccessFailedCount:    0,
	}

	if err := db.Create(&user).Error; err != nil {
		log.Fatalf("创建用户失败: %v\n提示：如果报 NOT NULL 或唯一约束错误，请改用 -reset 模式重置已有用户密码", err)
	}
	fmt.Printf("✓ 用户 [%s] 已创建 (Id=%d)\n", *flagUsername, newID)

	// 写入 Organ_Employee
	displayName := *flagName
	if displayName == "" {
		displayName = *flagUsername
	}
	emp := OrganEmployee{
		ID:   newID,
		Name: displayName,
	}
	if err := db.Create(&emp).Error; err != nil {
		fmt.Printf("⚠  写入 Organ_Employee 失败: %v\n", err)
	} else {
		fmt.Printf("✓ Organ_Employee 已写入，显示姓名: %s\n", displayName)
	}

	ensureAdminRole(db, *flagUsername, tenantID)
	fmt.Println("\n✓ 操作完成，请妥善保管密码。")
}

func resolveTenantID() int64 {
	if *flagTenantID > 0 {
		return *flagTenantID
	}
	if raw := os.Getenv("LEGACY_TENANT_ID"); raw != "" {
		if v, err := strconv.ParseInt(raw, 10, 64); err == nil && v > 0 {
			return v
		}
	}
	log.Fatal("请通过 -tenant-id 或 LEGACY_TENANT_ID 环境变量指定租户 ID")
	return 0
}

func mustConnect() *gorm.DB {
	dsn := *flagDSN
	if dsn == "" {
		host := requireEnv("DB_HOST")
		port := requireEnv("DB_PORT")
		user := requireEnv("DB_USER")
		pass := requireEnv("DB_PASSWORD")
		name := requireEnv("DB_NAME")
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

func listUsers(db *gorm.DB, tenantID int64) {
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

func resetPassword(db *gorm.DB, username, hash string, tenantID int64) {
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

func updateEmployeeName(db *gorm.DB, username, displayName string, tenantID int64) {
	var user IdentityUser
	if err := db.Where(`"UserName" = ? AND "TenantId" = ?`, username, tenantID).First(&user).Error; err != nil {
		fmt.Printf("⚠  更新 Organ_Employee 前查用户失败: %v\n", err)
		return
	}
	result := db.Model(&OrganEmployee{}).
		Where(`"Id" = ?`, user.ID).
		Updates(map[string]interface{}{"Name": displayName})
	if result.Error != nil {
		fmt.Printf("⚠  Organ_Employee 更新失败: %v\n", result.Error)
	} else if result.RowsAffected > 0 {
		fmt.Printf("✓ 显示姓名已更新: %s\n", displayName)
	}
}

func ensureAdminRole(db *gorm.DB, username string, tenantID int64) {
	var user IdentityUser
	if err := db.Where(`"UserName" = ? AND "TenantId" = ?`, username, tenantID).First(&user).Error; err != nil {
		fmt.Printf("⚠  查询用户失败，无法设置角色: %v\n", err)
		return
	}

	var adminRole IdentityRole
	err := db.Where(`("Name" = ? OR "NormalizedName" = ?)`, "ADMIN", "ADMIN").First(&adminRole).Error
	if err != nil {
		fmt.Println("⚠  Identity_Roles 中未找到 ADMIN 角色，请手动关联")
		return
	}

	var existing IdentityUserRole
	rErr := db.Where(`"UserId" = ? AND "RoleId" = ?`, user.ID, adminRole.ID).First(&existing).Error
	if rErr == nil {
		fmt.Printf("✓ 用户已有 ADMIN 角色 (UserRole 已存在)\n")
		return
	}
	if !errors.Is(rErr, gorm.ErrRecordNotFound) {
		fmt.Printf("⚠  查询 UserRole 失败: %v\n", rErr)
		return
	}

	userRole := IdentityUserRole{
		UserID: user.ID,
		RoleID: adminRole.ID,
	}
	if err := db.Create(&userRole).Error; err != nil {
		fmt.Printf("⚠  写入 Identity_UserRoles 失败: %v\n", err)
		return
	}
	fmt.Println("✓ 用户已被赋予 ADMIN 角色")
}

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

func requireEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		log.Fatalf("缺少环境变量 %s，请设置或通过 -dsn 指定完整连接串", key)
	}
	return v
}

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

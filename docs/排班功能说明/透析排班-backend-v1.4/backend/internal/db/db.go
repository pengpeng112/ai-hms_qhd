// Package db 负责 PostgreSQL 连接与表迁移。
package db

import (
	"os"
	"strconv"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/sdsph/dialysis-scheduling/internal/model"
)

// Open 用 DSN 打开 PostgreSQL 连接,并配置连接池。
// DSN 示例:"host=localhost user=postgres password=xxx dbname=aihms port=5432 sslmode=disable TimeZone=Asia/Shanghai"
//
// 连接池参数可用环境变量覆盖(均带默认值):
//
//	DB_MAX_OPEN_CONNS   最大打开连接数(默认 50)
//	DB_MAX_IDLE_CONNS   最大空闲连接数(默认 10)
//	DB_CONN_MAX_LIFE    连接最大生命周期,分钟(默认 30)
//	DB_CONN_MAX_IDLE    连接最大空闲时间,分钟(默认 5)
func Open(dsn string) (*gorm.DB, error) {
	g, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	sqlDB, err := g.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxOpenConns(envInt("DB_MAX_OPEN_CONNS", 50))
	sqlDB.SetMaxIdleConns(envInt("DB_MAX_IDLE_CONNS", 10))
	sqlDB.SetConnMaxLifetime(time.Duration(envInt("DB_CONN_MAX_LIFE", 30)) * time.Minute)
	sqlDB.SetConnMaxIdleTime(time.Duration(envInt("DB_CONN_MAX_IDLE", 5)) * time.Minute)
	return g, nil
}

// Ping 探活数据库(供健康检查用)。
func Ping(g *gorm.DB) error {
	sqlDB, err := g.DB()
	if err != nil {
		return err
	}
	return sqlDB.Ping()
}

// Migrate 创建/更新全部 Schedule_* 表,并建立数据库级唯一约束(防双排/双占机位)。
func Migrate(g *gorm.DB) error {
	if err := g.AutoMigrate(model.AllModels()...); err != nil {
		return err
	}
	return createUniqueIndexes(g)
}

// createUniqueIndexes 建立部分唯一索引(PostgreSQL 部分索引,GORM 标签无法表达)。
// 这是并发下防止重复排班/双占机位的数据库级安全网——即使应用层检查被并发击穿,
// 数据库也会拒绝冲突写入(返回 23505,由服务层转为友好错误)。
func createUniqueIndexes(g *gorm.DB) error {
	stmts := []string{
		// 同一病人在同一(日期+班次)不可重复有效排班(取消70/缺席80 不计;CRRT 的 ShiftId 为空,排除)。
		`CREATE UNIQUE INDEX IF NOT EXISTS uq_ps_patient_slot
		 ON "Schedule_PatientShift" ("TenantId","PatientId","ScheduleDate","ShiftId")
		 WHERE "Status" NOT IN (70,80) AND "ShiftId" IS NOT NULL`,
		// 同一台机在同一(日期+班次)只能一个有效占用(一班一机一人,决策 20)。
		`CREATE UNIQUE INDEX IF NOT EXISTS uq_ps_machine_slot
		 ON "Schedule_PatientShift" ("TenantId","MachineId","ScheduleDate","ShiftId")
		 WHERE "Status" NOT IN (70,80) AND "MachineId" IS NOT NULL AND "ShiftId" IS NOT NULL`,
	}
	for _, s := range stmts {
		if err := g.Exec(s).Error; err != nil {
			return err
		}
	}
	return nil
}

func envInt(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			return n
		}
	}
	return def
}

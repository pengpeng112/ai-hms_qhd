// Package db 负责 PostgreSQL 连接与表迁移。
package db

import (
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/sdsph/dialysis-scheduling/internal/model"
)

// Open 用 DSN 打开 PostgreSQL 连接。
// DSN 示例:"host=localhost user=postgres password=xxx dbname=aihms port=5432 sslmode=disable TimeZone=Asia/Shanghai"
func Open(dsn string) (*gorm.DB, error) {
	return gorm.Open(postgres.Open(dsn), &gorm.Config{})
}

// Migrate 创建/更新全部 Schedule_* 表。
func Migrate(g *gorm.DB) error {
	return g.AutoMigrate(model.AllModels()...)
}

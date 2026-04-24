package database

import (
	"errors"
	"log"

	"github.com/elliotxin/ai-hms-backend/config"
)

var errAutoMigratePermanentlyDisabled = errors.New("legacy database mode: AutoMigrate is permanently disabled")

// AutoMigrate 历史兼容占位函数。
//
// 注意：本服务已切换到老血透数据库并与老系统并行运行，
// 严禁通过 AutoMigrate 对生产库执行任何 DDL。
func AutoMigrate(_ *config.Config) error {
	log.Println("[LEGACY-DB] AutoMigrate call blocked: permanently disabled")
	return errAutoMigratePermanentlyDisabled
}

// DropTables 删除所有表（谨慎使用，仅用于开发环境测试）
func DropTables() error {
	return errors.New("legacy database mode: DropTables is disabled")
}

// GetTables 获取所有表
func GetTables() ([]string, error) {
	if DB == nil {
		return nil, nil
	}

	var tables []string
	err := DB.Table("information_schema.tables").
		Where("table_schema = ?", "public").
		Pluck("table_name", &tables).Error
	return tables, err
}

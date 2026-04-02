package database

import (
	"log"

	"github.com/elliotxin/ai-hms-backend/config"
	"github.com/elliotxin/ai-hms-backend/internal/models"
)

// AutoMigrate 自动迁移数据库表结构
// 仅在开发环境运行，生产环境应使用专业的迁移工具
func AutoMigrate(cfg *config.Config) error {
	if DB == nil {
		log.Println("Warning: Database not initialized, skipping migration")
		return nil
	}

	// 生产环境不自动迁移
	if cfg.Server.Mode == "release" {
		log.Println("Production environment: skipping AutoMigrate")
		return nil
	}

	log.Println("Development environment: running AutoMigrate...")

	err := DB.AutoMigrate(
		// 用户相关
		&models.User{},

		// 患者相关
		&models.Patient{},
		&models.PatientBasicInfo{},
		&models.VascularAccess{},
		&models.VascularAccessIntervention{},
		&models.MedicalHistory{},
		&models.OutcomeRecord{},
		&models.InfectionInfo{},
		&models.Hospitalization{},
		&models.LabReport{},
		&models.LabReportItem{},
		&models.ExamReport{},
		&models.PatientKeyIndicator{},
		&models.IntegrationHDISSetting{},

		// 治疗方案相关
		&models.TreatmentPlan{},
		&models.Prescription{},
		&models.Order{},
		&models.AdjustmentRecord{},

		// 排班相关
		&models.Ward{},
		&models.Bed{},
		&models.Shift{},
		&models.PatientShift{},

		// 透析治疗执行相关
		&models.Treatment{},
		&models.TreatmentBeforeCheck{},
		&models.TreatmentBeforeSigns{},
		&models.TreatmentDuringParam{},
		&models.TreatmentAfterSigns{},
		&models.TreatmentAlarm{},

		// 诊疗配置相关
		&models.PlanTemplate{},
		&models.MaterialCatalog{},
		&models.DrugCatalog{},
		&models.OrderTemplate{},
		&models.OrderTemplateItem{},

		// 字典管理相关
		&models.DictType{},
		&models.DictItem{},
	)

	if err != nil {
		return err
	}

	// 创建 dict_items 表的唯一索引 (type_code, code)
	// 防止字典项重复创建
	if err := createDictItemsUniqueIndex(); err != nil {
		log.Printf("Warning: Failed to create dict_items unique index: %v", err)
		// 不返回错误，因为索引可能已存在
	}

	log.Println("Database migration completed successfully")
	return nil
}

// createDictItemsUniqueIndex 创建字典项表的唯一索引
func createDictItemsUniqueIndex() error {
	// 检查索引是否已存在
	var count int64
	DB.Raw("SELECT COUNT(*) FROM pg_indexes WHERE indexname = 'idx_dict_items_unique'").Scan(&count)
	if count > 0 {
		return nil // 索引已存在
	}

	// 创建唯一索引
	return DB.Exec(`
		CREATE UNIQUE INDEX CONCURRENTLY IF NOT EXISTS idx_dict_items_unique
		ON dict_items(type_code, code)
	`).Error
}

// DropTables 删除所有表（谨慎使用，仅用于开发环境测试）
func DropTables() error {
	if DB == nil {
		return nil
	}

	err := DB.Migrator().DropTable(&models.User{})
	if err != nil {
		return err
	}

	log.Println("All tables dropped")
	return nil
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

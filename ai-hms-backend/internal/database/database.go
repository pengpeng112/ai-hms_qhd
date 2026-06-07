package database

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/elliotxin/ai-hms-backend/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

var DB *gorm.DB

// Initialize initializes the database connection.
func Initialize(cfg *config.DatabaseConfig, logCfg *config.LoggingConfig, parameterizedQueries bool) error {
	var dsn string
	if cfg.Password != "" {
		dsn = fmt.Sprintf(
			"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s TimeZone=%s options='-c client_encoding=UTF8'",
			cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode, cfg.TimeZone,
		)
	} else {
		dsn = fmt.Sprintf(
			"host=%s port=%s user=%s dbname=%s sslmode=%s TimeZone=%s options='-c client_encoding=UTF8'",
			cfg.Host, cfg.Port, cfg.User, cfg.DBName, cfg.SSLMode, cfg.TimeZone,
		)
	}

	slowThreshold := time.Second
	if logCfg != nil && logCfg.SlowSQLThreshold > 0 {
		slowThreshold = time.Duration(logCfg.SlowSQLThreshold) * time.Millisecond
	}

	gormConfig := &gorm.Config{
		Logger: logger.New(
			log.New(log.Writer(), "[gorm] ", log.LstdFlags),
			logger.Config{
				SlowThreshold:             slowThreshold,
				LogLevel:                  parseGormLogMode(logCfg),
				IgnoreRecordNotFoundError: true,
				ParameterizedQueries:      parameterizedQueries,
				Colorful:                  false,
			},
		),
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
			NoLowerCase:   true,
		},
		PrepareStmt:                              true,
		DisableForeignKeyConstraintWhenMigrating: true,
	}

	var err error
	DB, err = gorm.Open(postgres.Open(dsn), gormConfig)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	sqlDB, err := DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get database instance: %w", err)
	}

	maxLife := time.Hour
	if cfg.ConnMaxLifeMin > 0 {
		maxLife = time.Duration(cfg.ConnMaxLifeMin) * time.Minute
	}
	maxIdle := 5 * time.Minute
	if cfg.ConnMaxIdleMin > 0 {
		maxIdle = time.Duration(cfg.ConnMaxIdleMin) * time.Minute
	}
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(maxLife)
	sqlDB.SetConnMaxIdleTime(maxIdle)

	if err := sqlDB.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	if err := DB.Exec("SELECT 1").Error; err != nil {
		return fmt.Errorf("failed to verify database with SELECT 1: %w", err)
	}

	if err := DB.Exec("SET client_encoding = 'UTF8'").Error; err != nil {
		return fmt.Errorf("failed to set client_encoding to UTF8: %w", err)
	}

	log.Printf("[LEGACY-DB] connected to legacy hemodialysis database (%s:%s/%s), AutoMigrate permanently disabled", cfg.Host, cfg.Port, cfg.DBName)
	return nil
}

func parseGormLogMode(logCfg *config.LoggingConfig) logger.LogLevel {
	if logCfg == nil {
		return logger.Silent
	}

	switch strings.ToLower(strings.TrimSpace(logCfg.SQLMode)) {
	case "", "silent":
		return logger.Silent
	case "error":
		return logger.Error
	case "warn", "warning":
		return logger.Warn
	case "info", "debug":
		return logger.Info
	default:
		log.Printf("Warning: invalid LOG_SQL_MODE=%q, fallback to silent", logCfg.SQLMode)
		return logger.Silent
	}
}

// Close closes the database connection.
func Close() error {
	sqlDB, err := DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// GetDB returns the shared database handle.
func GetDB() *gorm.DB {
	return DB
}

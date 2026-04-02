package config

import (
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

// Config 应用配置
type Config struct {
	Server    ServerConfig
	Database  DatabaseConfig
	JWT       JWTConfig
	CORS      CORSConfig
	Hdis      HdisConfig
	AppSecret string
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Host string
	Port string
	Mode string // debug, release, test
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// JWTConfig JWT 配置
type JWTConfig struct {
	Secret          string
	ExpirationHours int
}

// CORSConfig CORS 配置
type CORSConfig struct {
	AllowedOrigins string
}

// HdisConfig HDIS 外部系统配置
type HdisConfig struct {
	WebcmdURL             string
	GraphqlURL            string
	AuthURL               string
	ClientID              string
	ServiceUser           string
	ServicePass           string
	Token                 string
	TimeoutSeconds        int
	BrowserHeadless       bool
	BrowserTimeoutSeconds int
	TargetOrganID         string
	Secret                string
}

// Load 加载配置
func Load() (*Config, error) {
	// 加载 .env 文件（如果存在）
	_ = godotenv.Load()

	cfg := &Config{
		Server: ServerConfig{
			Host: getEnv("SERVER_HOST", "0.0.0.0"),
			Port: getEnv("SERVER_PORT", "8080"),
			Mode: getEnv("GIN_MODE", "debug"),
		},
		Database: DatabaseConfig{
			Host:     mustGetEnv("DB_HOST"),
			Port:     mustGetEnv("DB_PORT"),
			User:     mustGetEnv("DB_USER"),
			Password: mustGetEnv("DB_PASSWORD"),
			DBName:   mustGetEnv("DB_NAME"),
			SSLMode:  getEnv("DB_SSL_MODE", "disable"),
		},
		JWT: JWTConfig{
			Secret:          mustGetEnv("JWT_SECRET"),
			ExpirationHours: getEnvInt("JWT_EXPIRATION_HOURS", 24),
		},
		CORS: CORSConfig{
			AllowedOrigins: mustGetEnv("CORS_ALLOWED_ORIGINS"),
		},
		Hdis: HdisConfig{
			// 这三项运行时仅从 integration_hdis_settings 表读取，不再从 .env 加载，
			// 避免环境变量与数据库配置混用导致排查歧义。
			WebcmdURL:             "",
			GraphqlURL:            "",
			AuthURL:               getEnv("HDIS_AUTH_URL", ""),
			ClientID:              getEnv("HDIS_CLIENT_ID", ""),
			ServiceUser:           getEnv("HDIS_SERVICE_USER", ""),
			ServicePass:           getEnv("HDIS_SERVICE_PASS", ""),
			Token:                 "",
			TimeoutSeconds:        getEnvInt("HDIS_TIMEOUT_SECONDS", 15),
			BrowserHeadless:       getEnvBool("HDIS_BROWSER_HEADLESS", true),
			BrowserTimeoutSeconds: getEnvInt("HDIS_BROWSER_TIMEOUT_SECONDS", 120),
			TargetOrganID:         getEnv("HDIS_TARGET_ORGAN_ID", ""),
			Secret:                mustGetEnv("APP_SECRET"),
		},
		AppSecret: mustGetEnv("APP_SECRET"),
	}

	// 验证必要配置
	// 注意：本地开发环境可能使用 trust 认证，不需要密码

	return cfg, nil
}

// getEnv 获取环境变量，如果不存在则返回默认值
func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

// mustGetEnv 获取必填环境变量，缺失时直接失败。
func mustGetEnv(key string) string {
	if val := strings.TrimSpace(os.Getenv(key)); val != "" {
		return val
	}
	log.Fatalf("missing required environment variable: %s", key)
	return ""
}

// getEnvInt 获取整数类型环境变量
func getEnvInt(key string, defaultVal int) int {
	if val := os.Getenv(key); val != "" {
		if intVal, err := strconv.Atoi(val); err == nil {
			return intVal
		}
	}
	return defaultVal
}

// getEnvBool 获取布尔类型环境变量
func getEnvBool(key string, defaultVal bool) bool {
	if val := os.Getenv(key); val != "" {
		if boolVal, err := strconv.ParseBool(val); err == nil {
			return boolVal
		}
	}
	return defaultVal
}

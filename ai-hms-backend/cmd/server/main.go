package main

import (
	"context"
	"io"
	"log"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/elliotxin/ai-hms-backend/config"
	v1api "github.com/elliotxin/ai-hms-backend/internal/api/v1"
	"github.com/elliotxin/ai-hms-backend/internal/database"
	"github.com/elliotxin/ai-hms-backend/internal/middleware"
	"github.com/elliotxin/ai-hms-backend/internal/services"
	"github.com/elliotxin/ai-hms-backend/internal/utils"
	"github.com/elliotxin/ai-hms-backend/pkg/response"
	"github.com/gin-gonic/gin"
)

func main() {
	// 加载配置
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 设置 Gin 模式
	gin.SetMode(cfg.Server.Mode)

	// 本地开发环境将日志同时写入 logs/app.log 与 logs/error.log，便于设置页日志排查读取。
	closeLogs, setupErr := setupLocalLogOutputs(cfg.Server.Mode)
	if setupErr != nil {
		log.Printf("Warning: failed to initialize local log outputs: %v", setupErr)
	}
	if closeLogs != nil {
		defer closeLogs()
	}

	// 初始化数据库
	if err := database.Initialize(&cfg.Database, &cfg.Logging); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	} else {
		defer database.Close()
		log.Println("[LEGACY-DB] startup in legacy database mode: AutoMigrate and startup seed initialization are disabled")

		// 启动过期临时医嘱定时任务
		services.StartOrderCron()
	}

	// 创建 JWT 管理器
	jwtManager := utils.NewJWTManager(&cfg.JWT)

	// 创建 Gin router
	r := gin.New()
	if cfg.Logging.RequestEnabled {
		r.Use(middleware.RequestLogger())
	}
	r.Use(gin.Recovery())

	// 全局中间件
	r.Use(middleware.CORS(cfg.CORS.AllowedOrigins))

	// 健康检查（无需认证）
	r.GET("/health", func(c *gin.Context) {
		response.Success(c, gin.H{
			"status":  "ok",
			"service": "ai-hms-backend",
			"version": "1.0.0",
		})
	})

	// API v1 路由组
	v1 := r.Group("/api/v1")
	{
		// 公开路由（无需认证）
		public := v1.Group("")
		{
			v1api.RegisterAuthRoutes(public, jwtManager)
		}

		// 需要认证的路由
		protected := v1.Group("")
		protected.Use(middleware.AuthMiddleware(jwtManager))
		{
			// 用户信息
			protected.GET("/me", func(c *gin.Context) {
				response.Success(c, gin.H{
					"user_id":  middleware.GetUserID(c),
					"username": middleware.GetUsername(c),
					"roles":    middleware.GetRoles(c),
				})
			})

			// 患者管理路由（包含患者相关的住院信息和核心信息聚合）
			v1api.RegisterPatientRoutesWithCore(protected)
			// 检验报告路由
			v1api.RegisterLabReportRoutes(protected)
			// 检查报告路由
			v1api.RegisterExamReportRoutes(protected)
			// 检查报告外部同步路由（HDIS）
			v1api.RegisterExamSyncRoutes(protected, cfg.Hdis)
			// 患者关键指标路由（HDIS Record）
			v1api.RegisterKeyIndicatorRoutes(protected, cfg.Hdis)
			// 检验报告外部同步路由（HDIS/LIS）
			v1api.RegisterLisSyncRoutes(protected, cfg.Hdis)
			// HDIS 集成配置路由（Settings > Integration）
			v1api.RegisterHDISSettingsRoutes(protected, cfg.Hdis)
			// 系统日志读取路由（Settings > Logs）
			v1api.RegisterLogRoutes(protected)

			// 住院信息路由
			v1api.RegisterHospitalizationRoutes(protected)

			// 排班管理路由
			v1api.RegisterScheduleRoutes(protected)

			// 治疗管理路由
			v1api.RegisterTreatmentRoutes(protected)

			// 临床病史路由
			v1api.RegisterMedicalHistoryRoutes(protected)

			// 诊疗配置路由
			v1api.RegisterTreatmentConfigRoutes(protected)

			// 字典管理路由
			v1api.RegisterDictRoutes(protected)

			// 用户管理路由（角色选择）
			v1api.RegisterUserRoutes(protected)

			// 设备管理路由
			v1api.RegisterDeviceRoutes(protected)

			// 库存管理路由
			v1api.RegisterInventoryRoutes(protected)

			// 看板统计路由
			v1api.RegisterDashboardRoutes(protected)
			v1api.RegisterClinicalTaskRoutes(protected)
			v1api.RegisterStatisticsRoutes(protected)
			v1api.RegisterPermissionRoutes(protected)
		}
	}

	// 启动服务器
	addr := cfg.Server.Host + ":" + cfg.Server.Port
	log.Printf("Server starting on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func setupLocalLogOutputs(mode string) (func(), error) {
	var closeFn func()
	if mode == gin.ReleaseMode {
		// 生产环境依赖 systemd 将 stdout/stderr 分别落到 app.log / error.log。
		gin.DefaultWriter = os.Stdout
		gin.DefaultErrorWriter = os.Stderr
	} else {
		if err := os.MkdirAll("logs", 0o755); err != nil {
			return nil, err
		}

		appPath := filepath.Join("logs", "app.log")
		errPath := filepath.Join("logs", "error.log")
		appFile, err := os.OpenFile(appPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
		if err != nil {
			return nil, err
		}
		errFile, err := os.OpenFile(errPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
		if err != nil {
			_ = appFile.Close()
			return nil, err
		}

		gin.DefaultWriter = io.MultiWriter(os.Stdout, appFile)
		gin.DefaultErrorWriter = io.MultiWriter(os.Stderr, errFile)
		closeFn = func() {
			_ = appFile.Close()
			_ = errFile.Close()
		}
	}

	log.SetOutput(gin.DefaultWriter)

	level := slog.LevelInfo
	if mode != gin.ReleaseMode {
		level = slog.LevelDebug
	}
	appSlogHandler := slog.NewTextHandler(gin.DefaultWriter, &slog.HandlerOptions{Level: level})
	errSlogHandler := slog.NewTextHandler(gin.DefaultErrorWriter, &slog.HandlerOptions{Level: slog.LevelWarn})
	slog.SetDefault(slog.New(&splitSlogHandler{
		infoHandler:  appSlogHandler,
		errorHandler: errSlogHandler,
	}))

	return closeFn, nil
}

type splitSlogHandler struct {
	infoHandler  slog.Handler
	errorHandler slog.Handler
}

func (h *splitSlogHandler) Enabled(ctx context.Context, level slog.Level) bool {
	if level >= slog.LevelWarn {
		return h.errorHandler.Enabled(ctx, level)
	}
	return h.infoHandler.Enabled(ctx, level)
}

func (h *splitSlogHandler) Handle(ctx context.Context, r slog.Record) error {
	if r.Level >= slog.LevelWarn {
		return h.errorHandler.Handle(ctx, r)
	}
	return h.infoHandler.Handle(ctx, r)
}

func (h *splitSlogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &splitSlogHandler{
		infoHandler:  h.infoHandler.WithAttrs(attrs),
		errorHandler: h.errorHandler.WithAttrs(attrs),
	}
}

func (h *splitSlogHandler) WithGroup(name string) slog.Handler {
	return &splitSlogHandler{
		infoHandler:  h.infoHandler.WithGroup(name),
		errorHandler: h.errorHandler.WithGroup(name),
	}
}

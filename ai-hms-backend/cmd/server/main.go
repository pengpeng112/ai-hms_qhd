package main

import (
	"context"
	"io"
	"log"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/elliotxin/ai-hms-backend/config"
	v1api "github.com/elliotxin/ai-hms-backend/internal/api/v1"
	"github.com/elliotxin/ai-hms-backend/internal/database"
	"github.com/elliotxin/ai-hms-backend/internal/integrations/idh"
	"github.com/elliotxin/ai-hms-backend/internal/middleware"
	"github.com/elliotxin/ai-hms-backend/internal/services"
	smartapi "github.com/elliotxin/ai-hms-backend/internal/smart_schedule/api"
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

	services.LegacyTenantID = cfg.LegacyTenantID

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
	if err := database.Initialize(&cfg.Database, &cfg.Logging, cfg.ParameterizedQueries); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	} else {
		defer database.Close()
		log.Println("[LEGACY-DB] startup in legacy database mode: AutoMigrate and startup seed initialization are disabled")

		// 默认管理员播种：默认关闭，避免向生产库注入已知口令账号。
		// 需显式 SEED_ADMIN_ENABLED=true 并提供 SEED_ADMIN_USERNAME/SEED_ADMIN_PASSWORD 才执行。
		if cfg.SeedAdminEnabled {
			services.SeedAdminIfNeeded(database.GetDB(), cfg.SeedAdminUsername, cfg.SeedAdminPassword)
		} else {
			log.Println("[SEED] admin seeding is disabled (set SEED_ADMIN_ENABLED=true with SEED_ADMIN_USERNAME/PASSWORD to enable)")
		}

		// 启动过期临时医嘱定时任务（需显式配置 ORDER_CRON_ENABLED=true 开启）
		if cfg.OrderCronEnabled {
			services.StartOrderCron(cfg.LegacyTenantID)
		} else {
			log.Println("[CRON] order cron is disabled (set ORDER_CRON_ENABLED=true to enable)")
		}
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
	// 安全响应头（防 XSS/点击劫持/MIME 嗅探；CSP 为 Report-Only，不拦截）
	r.Use(middleware.SecurityHeaders())

	// 健康检查（无需认证，含数据库连通性）
	r.GET("/health", func(c *gin.Context) {
		sqlDB, err := database.GetDB().DB()
		if err != nil || sqlDB.Ping() != nil {
			c.JSON(503, gin.H{
				"success":   false,
				"error":     gin.H{"code": "SERVICE_UNAVAILABLE", "message": "数据库连接异常"},
				"timestamp": time.Now().Format(time.RFC3339),
			})
			return
		}
		response.Success(c, gin.H{
			"status":  "ok",
			"db":      "ok",
			"service": "ai-hms-backend",
			"version": "1.0.0",
		})
	})

	// 智能排班 v2 路由组（需要认证）
	smartSchedule := r.Group("/api/v2")
	smartSchedule.Use(middleware.AuthMiddleware(jwtManager))
	smartapi.NewServer(database.GetDB()).Register(smartSchedule)

	// API v1 路由组
	v1 := r.Group("/api/v1")
	{
		// 公开路由（无需认证）
		public := v1.Group("")
		{
			v1api.RegisterAuthRoutes(public, jwtManager)
		}

		// 需要认证的路由（所有已登录用户可访问）
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
			// RNa 智能钠处方计算路由（专利 P2/P4/P6）
			v1api.RegisterRNaRoutes(protected)
			// 处方开单参考数据聚合路由（体重/检验/上次治疗/钠清除比/血压心率）
			v1api.RegisterPrescriptionContextRoutes(protected)
			// 临床指标编码核对路由（用真实 LIS/HDIS 字典对齐候选码）
			v1api.RegisterIndicatorMappingRoutes(protected)
			// 检查报告路由
			v1api.RegisterExamReportRoutes(protected)
			// 检查报告外部同步路由（HDIS）
			v1api.RegisterExamSyncRoutes(protected, cfg.Hdis)
			// 检查报告 HIS Oracle 同步路由
			v1api.RegisterHisExamSyncRoutes(protected, cfg.HisOracle, cfg.LegacyTenantID)
			// ACTRS 胸片分析微服务集成路由
			v1api.RegisterActrRoutes(protected, cfg.Actrs, cfg.LegacyTenantID)
			// CNRDS 上报路由
			v1api.RegisterCnrdsRoutes(protected, cfg.LegacyTenantID)
			// 患者关键指标路由（HDIS Record）
			v1api.RegisterKeyIndicatorRoutes(protected, cfg.Hdis)
			// 检验报告外部同步路由（HDIS/LIS）
			v1api.RegisterLisSyncRoutes(protected, cfg.Hdis)
			// HIS Oracle 连接测试路由
			v1api.RegisterHisOracleConfigRoutes(protected)
			// 同步任务管理路由
			v1api.RegisterSyncJobRoutes(protected, cfg.HisOracle, cfg.LegacyTenantID)
			// 同步患者管理路由（未匹配患者/绑定）
			v1api.RegisterSyncPatientRoutes(protected, cfg.HisOracle, cfg.LegacyTenantID)

			// 传染病筛查/门禁路由
			v1api.RegisterInfectiousRoutes(protected)

			// 消毒监管路由
			v1api.RegisterDisinfectionRoutes(protected)

			// 住院信息路由
			v1api.RegisterHospitalizationRoutes(protected)

			// 治疗管理路由
			v1api.RegisterTreatmentRoutes(protected)

			// 临床病史路由
			v1api.RegisterMedicalHistoryRoutes(protected)

			// 诊疗配置路由
			v1api.RegisterTreatmentConfigRoutes(protected)

			// 健康宣教路由
			v1api.RegisterHealthEducationRoutes(protected)

			// 病区管理路由
			v1api.RegisterWardRoutes(protected)

			// 床位管理路由
			v1api.RegisterBedRoutes(protected)

			// 设备管理路由
			v1api.RegisterDeviceRoutes(protected)

			// 库存管理路由
			v1api.RegisterInventoryRoutes(protected)

			// 看板统计路由
			v1api.RegisterDashboardRoutes(protected)
			v1api.RegisterClinicalTaskRoutes(protected)
			v1api.RegisterStatisticsRoutes(protected)
			// IDH 预警评分器：默认禁用（StubScorer→卡面"待数据"）；
			// 设 IDH_ENABLED=true + IDH_BASE_URL 后接 Python「IDH 预警」微服务（子项目B部署）。
			if cfg.IDH.Enabled && cfg.IDH.BaseURL != "" {
				services.SetIDHScorer(idh.NewHTTPScorer(idh.Config{
					BaseURL:    cfg.IDH.BaseURL,
					TimeoutSec: cfg.IDH.TimeoutSec,
					Enabled:    true,
				}))
				log.Printf("[IDH] HTTPScorer enabled, baseURL=%s", cfg.IDH.BaseURL)
			} else if cfg.IDH.Enabled && cfg.IDH.BaseURL == "" {
				log.Println("[IDH] IDH_ENABLED=true but IDH_BASE_URL is empty — using StubScorer")
			}

			v1api.RegisterMonitoringRoutes(protected)
			v1api.RegisterMonthlySummaryRoutes(protected)
			v1api.RegisterConsumableRoutes(protected)
			// 水质监测路由
			v1api.RegisterWaterQualityRoutes(protected)
			// 血管通路全生命周期路由
			v1api.RegisterVascularAccessEventRoutes(protected)
			// 不良事件登记路由
			v1api.RegisterAdverseEventRoutes(protected)
			// 长嘱给药执行路由
			v1api.RegisterMedicationRoutes(protected)
			// 干体重评估路由
			v1api.RegisterDryWeightRoutes(protected)
			// 护理文书路由
			v1api.RegisterNursingDocRoutes(protected)
			// 知情同意路由
			v1api.RegisterConsentRoutes(protected)
			// 收费归集路由（C4）
			v1api.RegisterBillingRoutes(protected, cfg.LegacyTenantID)
			// HIS 价表同步路由（C4）
			v1api.RegisterHisPriceRoutes(protected, cfg.HisOracle, cfg.LegacyTenantID)
		}

		admin := v1.Group("")
		// 权限码门禁 + 管理员角色兜底：命中 AdminPermissionCodes 任一权限码或任一管理员角色即放行。
		// 详见 middleware.RequirePermissions / AdminPermissionCodes 的过渡说明。
		permResolver := services.NewRolePermissionResolver(5 * time.Minute)
		admin.Use(
			middleware.AuthMiddleware(jwtManager),
			middleware.RequirePermissions(permResolver, middleware.AdminRoles, middleware.AdminPermissionCodes...),
		)
		{
			// HDIS 集成配置路由（Settings > Integration）
			v1api.RegisterHDISSettingsRoutes(admin, cfg.Hdis)
			// 系统日志读取路由（Settings > Logs）
			v1api.RegisterLogRoutes(admin)

			// 字典管理路由（含写操作）
			v1api.RegisterDictRoutes(admin)

			// 用户管理路由
			v1api.RegisterUserRoutes(admin)

			// 角色管理路由
			v1api.RegisterRoleManagementRoutes(admin)

			// 权限管理路由
			v1api.RegisterPermissionRoutes(admin)
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

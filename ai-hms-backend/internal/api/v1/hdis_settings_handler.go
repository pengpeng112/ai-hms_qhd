package v1

import (
	"context"
	"errors"
	"time"

	"github.com/elliotxin/ai-hms-backend/config"
	"github.com/elliotxin/ai-hms-backend/internal/middleware"
	"github.com/elliotxin/ai-hms-backend/internal/services"
	"github.com/elliotxin/ai-hms-backend/pkg/response"
	"github.com/gin-gonic/gin"
)

// HDISSettingsHandler HDIS 设置控制器
type HDISSettingsHandler struct {
	service        *services.HDISSettingsService
	browserTimeout time.Duration
}

func NewHDISSettingsHandler(cfg config.HdisConfig) *HDISSettingsHandler {
	bt := time.Duration(cfg.BrowserTimeoutSeconds) * time.Second
	if bt <= 0 {
		bt = 120 * time.Second
	}
	return &HDISSettingsHandler{
		service:        services.NewHDISSettingsService(cfg),
		browserTimeout: bt,
	}
}

// GetSettings 获取 HDIS 集成设置
// GET /api/v1/settings/integrations/hdis
func (h *HDISSettingsHandler) GetSettings(c *gin.Context) {
	settings, err := h.service.GetSettings()
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, settings)
}

// UpdateSettings 更新 HDIS 集成设置
// PUT /api/v1/settings/integrations/hdis
func (h *HDISSettingsHandler) UpdateSettings(c *gin.Context) {
	var req services.HdisIntegrationSettingsUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "无效的请求参数")
		return
	}

	updatedBy := middleware.GetUserID(c)
	settings, err := h.service.UpdateSettings(req, updatedBy)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrHDISSettingsInvalidInput):
			response.BadRequest(c, err.Error())
			return
		default:
			response.InternalError(c, err.Error())
			return
		}
	}

	response.Success(c, settings)
}

// RefreshToken 手动刷新 HDIS token
// POST /api/v1/settings/integrations/hdis/refresh-token
func (h *HDISSettingsHandler) RefreshToken(c *gin.Context) {
	// 超时 = 浏览器超时 + 10s 缓冲（浏览器启动 + DB 操作）
	ctx, cancel := context.WithTimeout(context.Background(), h.browserTimeout+10*time.Second)
	defer cancel()

	result, err := h.service.RefreshToken(ctx)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrHDISAuthNotConfigured):
			response.BadRequest(c, "HDIS 鉴权配置不完整，请先保存 Integration 配置")
			return
		default:
			response.InternalError(c, err.Error())
			return
		}
	}
	response.Success(c, result)
}

func RegisterHDISSettingsRoutes(r *gin.RouterGroup, cfg config.HdisConfig) {
	handler := NewHDISSettingsHandler(cfg)

	settings := r.Group("/settings/integrations/hdis")
	settings.Use(middleware.RequireRoles("ADMIN"))
	{
		settings.GET("", handler.GetSettings)
		settings.PUT("", handler.UpdateSettings)
		settings.POST("/refresh-token", handler.RefreshToken)
	}
}

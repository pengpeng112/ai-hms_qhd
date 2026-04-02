package v1

import (
	"errors"
	"strings"

	"github.com/elliotxin/ai-hms-backend/config"
	"github.com/elliotxin/ai-hms-backend/internal/services"
	"github.com/elliotxin/ai-hms-backend/pkg/response"
	"github.com/gin-gonic/gin"
)

// KeyIndicatorHandler 患者关键指标控制器
type KeyIndicatorHandler struct {
	service *services.KeyIndicatorService
}

// NewKeyIndicatorHandler 创建关键指标控制器
func NewKeyIndicatorHandler(cfg config.HdisConfig) *KeyIndicatorHandler {
	return &KeyIndicatorHandler{
		service: services.NewKeyIndicatorService(cfg),
	}
}

// ListByPatient 获取患者关键指标列表
// GET /api/v1/patients/:id/key-indicators
func (h *KeyIndicatorHandler) ListByPatient(c *gin.Context) {
	patientID := strings.TrimSpace(c.Param("id"))
	if patientID == "" {
		response.BadRequest(c, "患者ID不能为空")
		return
	}

	var req services.KeyIndicatorListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "无效的查询参数")
		return
	}

	result, err := h.service.ListByPatient(patientID, req)
	if err != nil {
		if strings.Contains(err.Error(), "required") || strings.Contains(err.Error(), "invalid") {
			response.BadRequest(c, err.Error())
			return
		}
		response.InternalError(c, err.Error())
		return
	}

	response.Paginated(c, result.Items, result.Page, result.PageSize, result.Total)
}

// SyncPatientKeyIndicators 同步患者关键指标
// POST /api/v1/patients/:id/key-indicators/sync
func (h *KeyIndicatorHandler) SyncPatientKeyIndicators(c *gin.Context) {
	patientID := strings.TrimSpace(c.Param("id"))
	result, err := h.service.SyncPatientKeyIndicators(patientID)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrSyncPatientIDRequired):
			response.BadRequest(c, err.Error())
			return
		case errors.Is(err, services.ErrSyncNotConfigured):
			response.BadRequest(c, "HDIS 对接未配置，请先在系统设置 > Integration 中完成配置")
			return
		case errors.Is(err, services.ErrSyncPatientBasicNotFound):
			response.NotFound(c, "患者基础档案不存在，请先完善患者基本信息")
			return
		case errors.Is(err, services.ErrSyncPatientHDISIDMissing):
			response.BadRequest(c, "患者缺少 HDIS 对应 ID，请先在患者基本信息中填写 hdisPatientId")
			return
		default:
			response.InternalError(c, err.Error())
			return
		}
	}

	response.Success(c, result)
}

// RegisterKeyIndicatorRoutes 注册关键指标路由
func RegisterKeyIndicatorRoutes(r *gin.RouterGroup, cfg config.HdisConfig) {
	handler := NewKeyIndicatorHandler(cfg)

	patients := r.Group("/patients")
	{
		patients.GET("/:id/key-indicators", handler.ListByPatient)
		patients.POST("/:id/key-indicators/sync", handler.SyncPatientKeyIndicators)
	}
}

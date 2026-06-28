package v1

import (
	"strconv"

	"github.com/elliotxin/ai-hms-backend/internal/middleware"
	"github.com/elliotxin/ai-hms-backend/internal/services"
	"github.com/elliotxin/ai-hms-backend/pkg/response"
	"github.com/gin-gonic/gin"
)

func RegisterMonitoringRoutes(r *gin.RouterGroup) {
	h := &MonitoringHandler{service: services.NewMonitoringService()}
	r.GET("/monitoring/live-data", h.GetLiveData)
	r.GET("/monitoring/treatments/:id/trend", h.GetTreatmentTrend)

	// 报警阈值表（读：任意登录；写：管理员）
	r.GET("/monitoring/thresholds", h.GetThresholds)
	admin := r.Group("", middleware.RequireRoles(middleware.AdminRoles...))
	admin.PUT("/monitoring/thresholds", h.SaveThresholds)
	admin.POST("/monitoring/thresholds/reset", h.ResetThresholds)
}

type MonitoringHandler struct {
	service *services.MonitoringService
}

func (h *MonitoringHandler) GetLiveData(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID <= 0 {
		response.Unauthorized(c, "tenant id missing")
		return
	}

	data, err := h.service.GetLiveData(tenantID)
	if err != nil {
		response.InternalError(c, "获取实时监测数据失败: "+err.Error())
		return
	}

	response.Success(c, data)
}

// GetTreatmentTrend 整场趋势（决①）：GET /monitoring/treatments/:id/trend
// 返回时间轴三界 + 各指标 actual 序列（predicted 段日后由外推/IDH 追加）。只读。
func (h *MonitoringHandler) GetTreatmentTrend(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID <= 0 {
		response.Unauthorized(c, "tenant id missing")
		return
	}

	treatmentID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || treatmentID <= 0 {
		response.BadRequest(c, "无效治疗ID")
		return
	}

	data, err := h.service.GetTreatmentTrend(tenantID, treatmentID)
	if err != nil {
		response.InternalError(c, "获取治疗趋势失败: "+err.Error())
		return
	}

	response.Success(c, data)
}

// GetThresholds 读取阈值表供 admin 展示（DB 优先、回退内嵌 JSON）。
func (h *MonitoringHandler) GetThresholds(c *gin.Context) {
	if middleware.GetTenantID(c) <= 0 {
		response.Unauthorized(c, "tenant id missing")
		return
	}
	response.Success(c, services.GetThresholdAdmin())
}

// SaveThresholds 整体保存阈值表（管理员）。
func (h *MonitoringHandler) SaveThresholds(c *gin.Context) {
	if middleware.GetTenantID(c) <= 0 {
		response.Unauthorized(c, "tenant id missing")
		return
	}
	var payload services.ThresholdAdminPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		response.BadRequest(c, "请求体解析失败: "+err.Error())
		return
	}
	operatorID, _ := strconv.ParseInt(middleware.GetUserID(c), 10, 64)
	if err := services.SaveThresholdAdmin(payload, operatorID); err != nil {
		if err == services.ErrThresholdTablesMissing {
			response.Error(c, 503, "TABLE_MISSING", err.Error())
			return
		}
		response.BadRequest(c, err.Error())
		return
	}
	response.Success(c, gin.H{"saved": true})
}

// ResetThresholds 恢复默认（管理员）。
func (h *MonitoringHandler) ResetThresholds(c *gin.Context) {
	if middleware.GetTenantID(c) <= 0 {
		response.Unauthorized(c, "tenant id missing")
		return
	}
	operatorID, _ := strconv.ParseInt(middleware.GetUserID(c), 10, 64)
	if err := services.ResetThresholdAdmin(operatorID); err != nil {
		if err == services.ErrThresholdTablesMissing {
			response.Error(c, 503, "TABLE_MISSING", err.Error())
			return
		}
		response.InternalError(c, "恢复默认失败: "+err.Error())
		return
	}
	response.Success(c, gin.H{"reset": true})
}

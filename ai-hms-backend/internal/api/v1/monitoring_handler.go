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

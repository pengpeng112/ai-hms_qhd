package v1

import (
	"github.com/elliotxin/ai-hms-backend/internal/middleware"
	"github.com/elliotxin/ai-hms-backend/internal/services"
	"github.com/elliotxin/ai-hms-backend/pkg/response"
	"github.com/gin-gonic/gin"
)

func RegisterMonitoringRoutes(r *gin.RouterGroup) {
	h := &MonitoringHandler{service: services.NewMonitoringService()}
	r.GET("/monitoring/live-data", h.GetLiveData)
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

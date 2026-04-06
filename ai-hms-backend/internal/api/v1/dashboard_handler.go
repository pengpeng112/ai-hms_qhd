package v1

import (
	"github.com/elliotxin/ai-hms-backend/internal/services"
	"github.com/elliotxin/ai-hms-backend/pkg/response"
	"github.com/gin-gonic/gin"
)

// DashboardHandler 看板处理器
type DashboardHandler struct {
	service *services.DashboardService
}

func NewDashboardHandler() *DashboardHandler {
	return &DashboardHandler{service: services.NewDashboardService()}
}

// GetStats GET /api/v1/dashboard/stats
func (h *DashboardHandler) GetStats(c *gin.Context) {
	stats, err := h.service.GetStats()
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, stats)
}

// RegisterDashboardRoutes 注册看板路由
func RegisterDashboardRoutes(rg *gin.RouterGroup) {
	h := NewDashboardHandler()
	dashboard := rg.Group("/dashboard")
	{
		dashboard.GET("/stats", h.GetStats)
	}
}

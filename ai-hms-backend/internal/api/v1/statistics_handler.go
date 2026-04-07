package v1

import (
	"fmt"
	"time"

	"github.com/elliotxin/ai-hms-backend/internal/middleware"
	"github.com/elliotxin/ai-hms-backend/internal/services"
	"github.com/elliotxin/ai-hms-backend/pkg/response"
	"github.com/gin-gonic/gin"
)

type StatisticsHandler struct {
	service *services.StatisticsService
}

func NewStatisticsHandler() *StatisticsHandler {
	return &StatisticsHandler{service: services.NewStatisticsService()}
}

func (h *StatisticsHandler) Quality(c *gin.Context) {
	tenantId := middleware.GetTenantID(c)
	year := parseYear(c.Query("year"))
	items, err := h.service.QualityByYear(tenantId, year)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, gin.H{"items": items})
}

func (h *StatisticsHandler) Infection(c *gin.Context) {
	tenantId := middleware.GetTenantID(c)
	year := parseYear(c.Query("year"))
	items, err := h.service.InfectionByYear(tenantId, year)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, gin.H{"items": items})
}

func (h *StatisticsHandler) Vascular(c *gin.Context) {
	tenantId := middleware.GetTenantID(c)
	year := parseYear(c.Query("year"))
	items, err := h.service.VascularByYear(tenantId, year)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, gin.H{"items": items})
}

func (h *StatisticsHandler) Workload(c *gin.Context) {
	tenantId := middleware.GetTenantID(c)
	yearMonth := c.DefaultQuery("yearMonth", time.Now().Format("2006-01"))
	items, err := h.service.WorkloadByYearMonth(tenantId, yearMonth)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, gin.H{"items": items})
}

func parseYear(value string) int {
	now := time.Now().Year()
	if value == "" {
		return now
	}
	var year int
	if _, err := fmt.Sscanf(value, "%d", &year); err != nil || year < 2000 || year > now+1 {
		return now
	}
	return year
}

func RegisterStatisticsRoutes(rg *gin.RouterGroup) {
	h := NewStatisticsHandler()
	group := rg.Group("/statistics")
	{
		group.GET("/quality", h.Quality)
		group.GET("/infection", h.Infection)
		group.GET("/vascular", h.Vascular)
		group.GET("/workload", h.Workload)
	}
}

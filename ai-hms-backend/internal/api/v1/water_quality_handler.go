package v1

import (
	"strconv"

	"github.com/elliotxin/ai-hms-backend/internal/services"
	"github.com/elliotxin/ai-hms-backend/pkg/response"
	"github.com/gin-gonic/gin"
)

type WaterQualityHandler struct{ svc *services.WaterQualityService }

func RegisterWaterQualityRoutes(rg *gin.RouterGroup) {
	h := &WaterQualityHandler{svc: services.NewWaterQualityService()}
	rg.POST("/water-quality/record", h.Record)
	rg.GET("/water-quality", h.List)
	rg.GET("/water-quality/conductivity", h.Conductivity)
	rg.GET("/water-quality/alerts", h.Alerts)
	rg.POST("/water-quality/:id/handle", h.Handle)
}

func (h *WaterQualityHandler) Record(c *gin.Context) {
	var in services.RecordInput
	if err := c.ShouldBindJSON(&in); err != nil {
		response.BadRequest(c, "请求体无效")
		return
	}
	rec, err := h.svc.Record(in)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.SuccessCreated(c, rec)
}

func (h *WaterQualityHandler) List(c *gin.Context) {
	rows, err := h.svc.List(services.WqListFilter{TestType: c.Query("testType"), SamplePoint: c.Query("samplePoint")})
	if err != nil {
		response.InternalErrorSafe(c)
		return
	}
	response.Success(c, rows)
}

func (h *WaterQualityHandler) Conductivity(c *gin.Context) {
	days := 7
	if d, err := strconv.Atoi(c.Query("days")); err == nil && d > 0 {
		days = d
	}
	pts, err := h.svc.ConductivityDaily(days)
	if err != nil {
		response.InternalErrorSafe(c)
		return
	}
	response.Success(c, pts)
}

func (h *WaterQualityHandler) Alerts(c *gin.Context) {
	a, err := h.svc.Alerts()
	if err != nil {
		response.InternalErrorSafe(c)
		return
	}
	response.Success(c, a)
}

func (h *WaterQualityHandler) Handle(c *gin.Context) {
	var in services.HandleInput
	if err := c.ShouldBindJSON(&in); err != nil {
		response.BadRequest(c, "请求体无效")
		return
	}
	rec, err := h.svc.Handle(c.Param("id"), in)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.Success(c, rec)
}

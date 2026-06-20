package v1

import (
	"strconv"

	"github.com/elliotxin/ai-hms-backend/internal/services"
	"github.com/elliotxin/ai-hms-backend/pkg/response"
	"github.com/gin-gonic/gin"
)

type InfectiousHandler struct{ svc *services.InfectiousService }

func RegisterInfectiousRoutes(rg *gin.RouterGroup) {
	h := &InfectiousHandler{svc: services.NewInfectiousService()}
	rg.POST("/patients/:id/infectious/screen", h.Screen)
	rg.GET("/patients/:id/infectious", h.History)
	rg.POST("/patients/:id/infectious/:recordId/dispose", h.Dispose)
	rg.GET("/infectious/alerts", h.Alerts)
}

func (h *InfectiousHandler) Screen(c *gin.Context) {
	pid, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效患者ID")
		return
	}
	var in services.ScreenInput
	if err := c.ShouldBindJSON(&in); err != nil {
		response.BadRequest(c, "请求体无效")
		return
	}
	rec, err := h.svc.Screen(pid, in)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.SuccessCreated(c, rec)
}

func (h *InfectiousHandler) History(c *gin.Context) {
	pid, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效患者ID")
		return
	}
	rows, gate, err := h.svc.HistoryWithGate(pid)
	if err != nil {
		response.InternalErrorSafe(c)
		return
	}
	response.Success(c, gin.H{"records": rows, "gate": gate})
}

func (h *InfectiousHandler) Dispose(c *gin.Context) {
	pid, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效患者ID")
		return
	}
	var in services.DispositionInput
	if err := c.ShouldBindJSON(&in); err != nil {
		response.BadRequest(c, "请求体无效")
		return
	}
	rec, err := h.svc.Dispose(pid, c.Param("recordId"), in)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.Success(c, rec)
}

func (h *InfectiousHandler) Alerts(c *gin.Context) {
	a, err := h.svc.Alerts()
	if err != nil {
		response.InternalErrorSafe(c)
		return
	}
	response.Success(c, a)
}

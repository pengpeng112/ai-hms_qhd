package v1

import (
	"strconv"
	"strings"

	"github.com/elliotxin/ai-hms-backend/internal/services"
	"github.com/elliotxin/ai-hms-backend/pkg/response"
	"github.com/gin-gonic/gin"
)

type DisinfectionHandler struct{ svc *services.DisinfectionService }

func RegisterDisinfectionRoutes(rg *gin.RouterGroup) {
	h := &DisinfectionHandler{svc: services.NewDisinfectionService()}
	rg.POST("/disinfection/record", h.Record)
	rg.POST("/disinfection/:id/compliance", h.SaveCompliance)
	rg.GET("/disinfection/machines", h.Machines)
	rg.GET("/disinfection/alerts", h.Alerts)
	rg.GET("/disinfection/stats", h.Stats)
}

func parseIDList(q string) []int64 {
	var out []int64
	for _, p := range strings.Split(q, ",") {
		if v, err := strconv.ParseInt(strings.TrimSpace(p), 10, 64); err == nil && v > 0 {
			out = append(out, v)
		}
	}
	return out
}

func (h *DisinfectionHandler) Record(c *gin.Context) {
	var in services.DisinfectRecordInput
	if err := c.ShouldBindJSON(&in); err != nil {
		response.BadRequest(c, "请求体无效")
		return
	}
	r, err := h.svc.Record(in)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.SuccessCreated(c, r)
}

func (h *DisinfectionHandler) SaveCompliance(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效记录ID")
		return
	}
	var in services.ComplianceInput
	if err := c.ShouldBindJSON(&in); err != nil {
		response.BadRequest(c, "请求体无效")
		return
	}
	comp, err := h.svc.SaveCompliance(id, in)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.Success(c, comp)
}

func (h *DisinfectionHandler) Machines(c *gin.Context) {
	ids := parseIDList(c.Query("deviceIds"))
	out := make([]services.MachineDisinfStatus, 0, len(ids))
	for _, id := range ids {
		out = append(out, h.svc.MachineStatus(id))
	}
	response.Success(c, out)
}

func (h *DisinfectionHandler) Alerts(c *gin.Context) {
	a, err := h.svc.Alerts(parseIDList(c.Query("deviceIds")))
	if err != nil {
		response.InternalErrorSafe(c)
		return
	}
	response.Success(c, a)
}

func (h *DisinfectionHandler) Stats(c *gin.Context) {
	response.Success(c, h.svc.Stats(parseIDList(c.Query("deviceIds"))))
}

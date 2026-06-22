package v1

import (
	"strconv"

	"github.com/elliotxin/ai-hms-backend/internal/services"
	"github.com/elliotxin/ai-hms-backend/pkg/response"
	"github.com/gin-gonic/gin"
)

type AdverseEventHandler struct{ svc *services.AdverseEventService }

func RegisterAdverseEventRoutes(rg *gin.RouterGroup) {
	h := &AdverseEventHandler{svc: services.NewAdverseEventService()}
	rg.POST("/adverse-events", h.Register)
	rg.GET("/adverse-events", h.List)
	rg.GET("/adverse-events/alerts", h.Alerts)
	rg.GET("/adverse-events/:id", h.Get)
	rg.POST("/adverse-events/:id/report", h.Report)
	rg.POST("/adverse-events/:id/status", h.UpdateStatus)
}

func (h *AdverseEventHandler) Register(c *gin.Context) {
	var raw struct {
		PatientID   int64  `json:"patientId"`
		TreatmentID *int64 `json:"treatmentId"`
		EventType   string `json:"eventType"`
		Severity    string `json:"severity"`
		OccurredAt  string `json:"occurredAt"`
		Description string `json:"description"`
		Handling    string `json:"handling"`
		Outcome     string `json:"outcome"`
		ReporterID  string `json:"reporterId"`
	}
	if err := c.ShouldBindJSON(&raw); err != nil {
		response.BadRequest(c, "请求体无效")
		return
	}
	ot, err := parseVascDate(raw.OccurredAt)
	if err != nil {
		response.BadRequest(c, "发生时间格式错误："+err.Error())
		return
	}
	rec, err := h.svc.Register(services.AeRegisterInput{
		PatientID: raw.PatientID, TreatmentID: raw.TreatmentID, EventType: raw.EventType,
		Severity: raw.Severity, OccurredAt: ot, Description: raw.Description,
		Handling: raw.Handling, Outcome: raw.Outcome, ReporterID: raw.ReporterID,
	})
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.SuccessCreated(c, rec)
}

func (h *AdverseEventHandler) List(c *gin.Context) {
	severity := c.Query("severity")
	status := c.Query("status")
	var pid *int64
	if v := c.Query("patientId"); v != "" {
		n, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			response.BadRequest(c, "无效患者ID")
			return
		}
		pid = &n
	}
	rows, err := h.svc.List(severity, status, pid)
	if err != nil {
		response.InternalErrorSafe(c)
		return
	}
	response.Success(c, rows)
}

func (h *AdverseEventHandler) Alerts(c *gin.Context) {
	a, err := h.svc.Alerts()
	if err != nil {
		response.InternalErrorSafe(c)
		return
	}
	response.Success(c, a)
}

func (h *AdverseEventHandler) Get(c *gin.Context) {
	rec, err := h.svc.GetByID(c.Param("id"))
	if err != nil {
		response.NotFound(c, "记录不存在")
		return
	}
	response.Success(c, rec)
}

func (h *AdverseEventHandler) Report(c *gin.Context) {
	var raw struct {
		ReportedTo []services.AeReportTarget `json:"reportedTo"`
	}
	if err := c.ShouldBindJSON(&raw); err != nil {
		response.BadRequest(c, "请求体无效")
		return
	}
	rec, err := h.svc.Report(c.Param("id"), services.AeReportInput{ReportedTo: raw.ReportedTo})
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.Success(c, rec)
}

func (h *AdverseEventHandler) UpdateStatus(c *gin.Context) {
	var raw struct {
		Status    string `json:"status"`
		CqiLinked *bool  `json:"cqiLinked"`
	}
	if err := c.ShouldBindJSON(&raw); err != nil {
		response.BadRequest(c, "请求体无效")
		return
	}
	rec, err := h.svc.UpdateStatus(c.Param("id"), services.AeStatusInput{Status: raw.Status, CqiLinked: raw.CqiLinked})
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.Success(c, rec)
}

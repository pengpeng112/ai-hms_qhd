package v1

import (
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/elliotxin/ai-hms-backend/internal/services"
	"github.com/elliotxin/ai-hms-backend/pkg/response"
	"github.com/gin-gonic/gin"
)

type VascularAccessEventHandler struct{ svc *services.VascularAccessMonitor }

func RegisterVascularAccessEventRoutes(rg *gin.RouterGroup) {
	h := &VascularAccessEventHandler{svc: services.NewVascularAccessMonitor()}
	rg.POST("/patients/:id/vascular-access-events", h.RecordEvent)
	rg.GET("/patients/:id/vascular-access-timeline", h.Timeline)
	rg.GET("/patients/:id/vascular-access-reminders", h.Reminders)
	rg.GET("/vascular-access/alerts", h.Alerts)
}

func (h *VascularAccessEventHandler) RecordEvent(c *gin.Context) {
	pid, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效患者ID")
		return
	}
	var raw struct {
		AccessID   int64  `json:"accessId"`
		EventType  string `json:"eventType"`
		EventDate  string `json:"eventDate"`
		Detail     string `json:"detail"`
		OperatorID string `json:"operatorId"`
		Note       string `json:"note"`
	}
	if err := c.ShouldBindJSON(&raw); err != nil {
		response.BadRequest(c, "请求体无效")
		return
	}
	ed, err := parseVascDate(raw.EventDate)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	rec, err := h.svc.RecordEvent(pid, raw.AccessID, services.VascEventInput{
		EventType: raw.EventType, EventDate: ed, Detail: raw.Detail, OperatorID: raw.OperatorID, Note: raw.Note,
	})
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.SuccessCreated(c, rec)
}

func (h *VascularAccessEventHandler) Timeline(c *gin.Context) {
	pid, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效患者ID")
		return
	}
	tl, err := h.svc.Timeline(pid)
	if err != nil {
		response.InternalErrorSafe(c)
		return
	}
	response.Success(c, tl)
}

func (h *VascularAccessEventHandler) Reminders(c *gin.Context) {
	pid, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效患者ID")
		return
	}
	response.Success(c, h.svc.PatientReminders(pid))
}

func (h *VascularAccessEventHandler) Alerts(c *gin.Context) {
	a, err := h.svc.Alerts()
	if err != nil {
		response.InternalErrorSafe(c)
		return
	}
	response.Success(c, a)
}

func parseVascDate(s string) (time.Time, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return time.Time{}, errors.New("事件日期不能为空")
	}
	for _, layout := range []string{time.RFC3339, "2006-01-02T15:04:05", "2006-01-02"} {
		if t, err := time.Parse(layout, s); err == nil {
			return t, nil
		}
	}
	return time.Time{}, errors.New("事件日期格式错误")
}

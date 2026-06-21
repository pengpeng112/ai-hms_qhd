package v1

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/elliotxin/ai-hms-backend/internal/services"
	"github.com/elliotxin/ai-hms-backend/pkg/response"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type CnrdsHandler struct {
	svc *services.CnrdsService
}

func RegisterCnrdsRoutes(rg *gin.RouterGroup, tenantID int64) {
	h := &CnrdsHandler{svc: services.NewCnrdsService(tenantID)}
	rg.POST("/cnrds/monthly", h.GenerateMonthly)
	rg.POST("/cnrds/event", h.GenerateEvent)
	rg.GET("/cnrds", h.List)
	rg.GET("/cnrds/:id", h.Get)
	rg.GET("/cnrds/:id/export", h.Export)
	rg.POST("/cnrds/:id/submit", h.Submit)
}

func (h *CnrdsHandler) GenerateMonthly(c *gin.Context) {
	period := c.Query("period")
	if period == "" {
		response.BadRequest(c, "缺少周期参数 period")
		return
	}
	rep, err := h.svc.GenerateMonthly(period)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.SuccessCreated(c, rep)
}

type generateEventReq struct {
	PatientID string `json:"patientId"`
	EventType string `json:"eventType"`
}

func (h *CnrdsHandler) GenerateEvent(c *gin.Context) {
	var req generateEventReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求格式错误")
		return
	}
	if req.PatientID == "" || req.EventType == "" {
		response.BadRequest(c, "patientId 和 eventType 不能为空")
		return
	}
	pid, err := strconv.ParseInt(req.PatientID, 10, 64)
	if err != nil {
		response.BadRequest(c, "无效患者ID")
		return
	}
	rep, err := h.svc.GenerateEvent(pid, req.EventType)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.SuccessCreated(c, rep)
}

func (h *CnrdsHandler) List(c *gin.Context) {
	f := services.CnrdsListFilter{
		Period:     c.Query("period"),
		ReportType: c.Query("reportType"),
		Status:     c.Query("status"),
	}
	reports, err := h.svc.List(f)
	if err != nil {
		response.InternalErrorSafe(c)
		return
	}
	response.Success(c, reports)
}

func (h *CnrdsHandler) Get(c *gin.Context) {
	id := c.Param("id")
	rep, err := h.svc.Get(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.NotFound(c, "报告不存在")
			return
		}
		response.InternalErrorSafe(c)
		return
	}
	response.Success(c, rep)
}

func (h *CnrdsHandler) Export(c *gin.Context) {
	id := c.Param("id")
	filename, data, err := h.svc.Export(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.NotFound(c, "报告不存在")
			return
		}
		response.Error(c, http.StatusConflict, "CONFLICT", err.Error())
		return
	}
	c.Header("Content-Disposition", "attachment; filename=\""+filename+"\"")
	c.Data(http.StatusOK, "text/csv; charset=utf-8", data)
}

type submitReq struct {
	ReviewedBy string `json:"reviewedBy"`
}

func (h *CnrdsHandler) Submit(c *gin.Context) {
	id := c.Param("id")
	var req submitReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求格式错误")
		return
	}
	if req.ReviewedBy == "" {
		response.BadRequest(c, "提交须填写核对人")
		return
	}
	if err := h.svc.Submit(id, req.ReviewedBy); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.NotFound(c, "报告不存在")
			return
		}
		response.Error(c, http.StatusConflict, "CONFLICT", err.Error())
	}
	response.Success(c, gin.H{"ok": true})
}

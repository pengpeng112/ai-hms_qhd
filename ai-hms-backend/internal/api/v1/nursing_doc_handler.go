package v1

import (
	"github.com/elliotxin/ai-hms-backend/internal/services"
	"github.com/elliotxin/ai-hms-backend/pkg/response"
	"github.com/gin-gonic/gin"
)

// NursingDocHandler 护理文书（C1）：量表评估 / 护理记录 / 护理计划。
type NursingDocHandler struct {
	svc *services.NursingDocService
}

func RegisterNursingDocRoutes(rg *gin.RouterGroup) {
	h := &NursingDocHandler{svc: services.NewNursingDocService()}
	rg.GET("/nursing/scales", h.Scales)
	rg.POST("/nursing/scales", h.RecordScale)
	rg.POST("/nursing/docs", h.RecordDoc)
	rg.GET("/nursing/docs", h.List)
	rg.GET("/nursing/alerts", h.Alerts)
}

// Scales 返回启用的量表定义（前端据此渲染录入表单）
func (h *NursingDocHandler) Scales(c *gin.Context) {
	response.Success(c, h.svc.EnabledScales())
}

func (h *NursingDocHandler) RecordScale(c *gin.Context) {
	var in services.ScaleRecordInput
	if err := c.ShouldBindJSON(&in); err != nil {
		response.BadRequest(c, "请求格式错误")
		return
	}
	rec, err := h.svc.RecordScale(in)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.SuccessCreated(c, rec)
}

func (h *NursingDocHandler) RecordDoc(c *gin.Context) {
	var in services.DocRecordInput
	if err := c.ShouldBindJSON(&in); err != nil {
		response.BadRequest(c, "请求格式错误")
		return
	}
	rec, err := h.svc.RecordDoc(in)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.SuccessCreated(c, rec)
}

func (h *NursingDocHandler) List(c *gin.Context) {
	f := services.NursingListFilter{
		PatientID:   c.Query("patientId"),
		TreatmentID: c.Query("treatmentId"),
		DocType:     c.Query("docType"),
		ScaleType:   c.Query("scaleType"),
	}
	rows, err := h.svc.List(f)
	if err != nil {
		response.InternalErrorSafe(c)
		return
	}
	response.Success(c, rows)
}

func (h *NursingDocHandler) Alerts(c *gin.Context) {
	rows, err := h.svc.HighRiskAlerts()
	if err != nil {
		response.InternalErrorSafe(c)
		return
	}
	response.Success(c, rows)
}

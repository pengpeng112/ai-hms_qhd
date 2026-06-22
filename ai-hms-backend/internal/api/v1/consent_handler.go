package v1

import (
	"github.com/elliotxin/ai-hms-backend/internal/services"
	"github.com/elliotxin/ai-hms-backend/pkg/response"
	"github.com/gin-gonic/gin"
)

// ConsentHandler 知情同意（C2）：开具 / 签署 / 撤销 / 查询。
type ConsentHandler struct {
	svc *services.ConsentService
}

func RegisterConsentRoutes(rg *gin.RouterGroup) {
	h := &ConsentHandler{svc: services.NewConsentService()}
	rg.GET("/consents/templates", h.Templates)
	rg.POST("/consents", h.Issue)
	rg.GET("/consents", h.List)
	rg.GET("/consents/alerts", h.Alerts)
	rg.POST("/consents/:id/sign", h.Sign)
	rg.POST("/consents/:id/revoke", h.Revoke)
}

func (h *ConsentHandler) Templates(c *gin.Context) {
	response.Success(c, h.svc.EnabledTemplates())
}

func (h *ConsentHandler) Issue(c *gin.Context) {
	var in services.ConsentIssueInput
	if err := c.ShouldBindJSON(&in); err != nil {
		response.BadRequest(c, "请求格式错误")
		return
	}
	rec, err := h.svc.Issue(in)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.SuccessCreated(c, rec)
}

func (h *ConsentHandler) List(c *gin.Context) {
	f := services.ConsentListFilter{
		PatientID:   c.Query("patientId"),
		ConsentType: c.Query("consentType"),
		Status:      c.Query("status"),
	}
	rows, err := h.svc.List(f)
	if err != nil {
		response.InternalErrorSafe(c)
		return
	}
	response.Success(c, rows)
}

func (h *ConsentHandler) Alerts(c *gin.Context) {
	out, err := h.svc.Alerts()
	if err != nil {
		response.InternalErrorSafe(c)
		return
	}
	response.Success(c, out)
}

func (h *ConsentHandler) Sign(c *gin.Context) {
	var in services.ConsentSignInput
	if err := c.ShouldBindJSON(&in); err != nil {
		response.BadRequest(c, "请求格式错误")
		return
	}
	rec, err := h.svc.Sign(c.Param("id"), in)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.Success(c, rec)
}

func (h *ConsentHandler) Revoke(c *gin.Context) {
	rec, err := h.svc.Revoke(c.Param("id"))
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.Success(c, rec)
}

package v1

import (
	"github.com/elliotxin/ai-hms-backend/internal/middleware"
	"github.com/elliotxin/ai-hms-backend/internal/models"
	"github.com/elliotxin/ai-hms-backend/internal/services"
	"github.com/elliotxin/ai-hms-backend/internal/utils"
	"github.com/elliotxin/ai-hms-backend/pkg/response"
	"github.com/gin-gonic/gin"
)

var BillingWriteRoles = []string{"NURSE", "DOCTOR", "ADMIN", "管理员"}

type BillingHandler struct {
	svc *services.BillingService
}

func NewBillingHandler(tenantID int64) *BillingHandler {
	return &BillingHandler{
		svc: services.NewBillingService(tenantID),
	}
}

func RegisterBillingRoutes(rg *gin.RouterGroup, tenantID int64) {
	h := NewBillingHandler(tenantID)

	rg.GET("/charges", h.List)
	rg.GET("/charges/:id", h.Get)

	write := rg.Group("")
	write.Use(middleware.RequireRoles(BillingWriteRoles...))
	write.POST("/charges/build", h.Build)
	write.POST("/charges/:id/lines", h.AddLine)
	write.PATCH("/charges/lines/:lineId", h.UpdateLine)
	write.DELETE("/charges/lines/:lineId", h.DeleteLine)
	write.POST("/charges/:id/confirm", h.Confirm)
	write.POST("/charges/:id/check", h.Check)
	write.POST("/charges/:id/exported", h.Export)
	write.POST("/charges/:id/cancel", h.Cancel)
	write.POST("/charges/:id/push", h.Push)
}

func (h *BillingHandler) Build(c *gin.Context) {
	var req services.BuildDraftRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	userID := middleware.GetUserID(c)
	userName := middleware.GetUsername(c)
	rec, err := h.svc.BuildDraft(req, userID, userName)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, rec)
}

func (h *BillingHandler) List(c *gin.Context) {
	var req services.ListChargesRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	result, err := h.svc.ListCharges(req)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, result)
}

func (h *BillingHandler) Get(c *gin.Context) {
	id := c.Param("id")
	rec, err := h.svc.GetCharge(id)
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}
	response.Success(c, rec)
}

func (h *BillingHandler) AddLine(c *gin.Context) {
	chargeID := c.Param("id")
	var line models.ChargeLine
	if err := c.ShouldBindJSON(&line); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	line.ID = utils.GenerateID()
	result, err := h.svc.AddLine(chargeID, &line)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, result)
}

func (h *BillingHandler) UpdateLine(c *gin.Context) {
	lineID := c.Param("lineId")
	var patch map[string]interface{}
	if err := c.ShouldBindJSON(&patch); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	result, err := h.svc.UpdateLine(lineID, patch)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, result)
}

func (h *BillingHandler) DeleteLine(c *gin.Context) {
	lineID := c.Param("lineId")
	if err := h.svc.DeleteLine(lineID); err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, nil)
}

func (h *BillingHandler) Confirm(c *gin.Context) {
	id := c.Param("id")
	userID := middleware.GetUserID(c)
	userName := middleware.GetUsername(c)
	rec, err := h.svc.Confirm(id, userID, userName)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, rec)
}

func (h *BillingHandler) Check(c *gin.Context) {
	id := c.Param("id")
	userID := middleware.GetUserID(c)
	userName := middleware.GetUsername(c)
	rec, err := h.svc.Check(id, userID, userName)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, rec)
}

func (h *BillingHandler) Export(c *gin.Context) {
	id := c.Param("id")
	rec, err := h.svc.MarkExported(id)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, rec)
}

func (h *BillingHandler) Cancel(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		Reason string `json:"reason"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	if req.Reason == "" {
		req.Reason = "用户取消"
	}
	rec, err := h.svc.Cancel(id, req.Reason)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, rec)
}

func (h *BillingHandler) Push(c *gin.Context) {
	response.Error(c, 501, "NOT_IMPLEMENTED", "HIS 收费推送接口暂未启用，请导出 Excel 后人工录入 HIS")
}

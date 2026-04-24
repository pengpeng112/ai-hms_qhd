package v1

import (
	"github.com/elliotxin/ai-hms-backend/internal/middleware"
	"github.com/elliotxin/ai-hms-backend/internal/services"
	"github.com/elliotxin/ai-hms-backend/pkg/response"
	"github.com/gin-gonic/gin"
)

// PrescriptionHandler 处方控制器
type PrescriptionHandler struct {
	service *services.PrescriptionService
}

// NewPrescriptionHandler 创建处方控制器
func NewPrescriptionHandler() *PrescriptionHandler {
	return &PrescriptionHandler{
		service: services.NewPrescriptionService(),
	}
}

// List 获取处方列表
func (h *PrescriptionHandler) List(c *gin.Context) {
	patientID := c.Param("id")
	if patientID == "" {
		response.BadRequest(c, "患者ID不能为空")
		return
	}

	prescriptions, err := h.service.LegacyList(patientID)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, prescriptions)
}

// Get 获取处方详情
func (h *PrescriptionHandler) Get(c *gin.Context) {
	patientID := c.Param("id")
	prescriptionID := c.Param("pid")
	if patientID == "" || prescriptionID == "" {
		response.BadRequest(c, "患者ID和处方ID不能为空")
		return
	}

	p, err := h.service.LegacyGet(patientID, prescriptionID)
	if err != nil {
		if err.Error() == "prescription not found" {
			response.NotFound(c, "处方不存在")
			return
		}
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, p)
}

// Create 创建处方
func (h *PrescriptionHandler) Create(c *gin.Context) {
	patientID := c.Param("id")
	if patientID == "" {
		response.BadRequest(c, "患者ID不能为空")
		return
	}

	var req services.PrescriptionCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "无效的请求参数: "+err.Error())
		return
	}

	doctorID := middleware.GetUserID(c)
	doctorName := middleware.GetUsername(c)

	p, err := h.service.LegacyCreate(patientID, doctorID, doctorName, req)
	if err != nil {
		if err.Error() == "请先创建启用的治疗方案" {
			response.BadRequest(c, err.Error())
			return
		}
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, p)
}

// Update 更新处方
func (h *PrescriptionHandler) Update(c *gin.Context) {
	patientID := c.Param("id")
	prescriptionID := c.Param("pid")
	if patientID == "" || prescriptionID == "" {
		response.BadRequest(c, "患者ID和处方ID不能为空")
		return
	}

	var req services.PrescriptionUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "无效的请求参数: "+err.Error())
		return
	}

	p, err := h.service.LegacyUpdate(patientID, prescriptionID, req)
	if err != nil {
		if err.Error() == "prescription not found" {
			response.NotFound(c, "处方不存在")
			return
		}
		if err.Error() == "仅待执行状态的处方可编辑" {
			response.BadRequest(c, err.Error())
			return
		}
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, p)
}

// Execute 发布/执行处方
func (h *PrescriptionHandler) Execute(c *gin.Context) {
	patientID := c.Param("id")
	prescriptionID := c.Param("pid")
	if patientID == "" || prescriptionID == "" {
		response.BadRequest(c, "患者ID和处方ID不能为空")
		return
	}

	executedBy := middleware.GetUserID(c)

	p, err := h.service.LegacyExecute(patientID, prescriptionID, executedBy)
	if err != nil {
		if err.Error() == "prescription not found" {
			response.NotFound(c, "处方不存在")
			return
		}
		if err.Error() == "已取消的处方不能执行" {
			response.BadRequest(c, err.Error())
			return
		}
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, p)
}

// Cancel 取消处方
func (h *PrescriptionHandler) Cancel(c *gin.Context) {
	patientID := c.Param("id")
	prescriptionID := c.Param("pid")
	if patientID == "" || prescriptionID == "" {
		response.BadRequest(c, "患者ID和处方ID不能为空")
		return
	}

	p, err := h.service.LegacyCancel(patientID, prescriptionID)
	if err != nil {
		if err.Error() == "prescription not found" {
			response.NotFound(c, "处方不存在")
			return
		}
		if err.Error() == "已执行的处方不能取消" {
			response.BadRequest(c, err.Error())
			return
		}
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, p)
}

// Extract 提取长嘱生成处方
func (h *PrescriptionHandler) Extract(c *gin.Context) {
	patientID := c.Param("id")
	if patientID == "" {
		response.BadRequest(c, "患者ID不能为空")
		return
	}

	var req services.PrescriptionExtractRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "无效的请求参数: "+err.Error())
		return
	}

	doctorID := middleware.GetUserID(c)
	doctorName := middleware.GetUsername(c)

	p, err := h.service.LegacyExtractFromLongTermOrders(patientID, doctorID, doctorName, req.Date)
	if err != nil {
		if err.Error() == "请先创建启用的治疗方案" {
			response.BadRequest(c, err.Error())
			return
		}
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, p)
}

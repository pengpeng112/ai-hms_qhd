package v1

import (
	"github.com/elliotxin/ai-hms-backend/internal/services"
	"github.com/elliotxin/ai-hms-backend/pkg/response"
	"github.com/gin-gonic/gin"
)

// VascularAccessHandler 血管通路控制器
type VascularAccessHandler struct {
	service *services.VascularAccessService
}

// NewVascularAccessHandler 创建血管通路控制器
func NewVascularAccessHandler() *VascularAccessHandler {
	return &VascularAccessHandler{
		service: services.NewVascularAccessService(),
	}
}

// List 获取患者的血管通路列表
func (h *VascularAccessHandler) List(c *gin.Context) {
	patientID := c.Param("id")

	result, err := h.service.List(patientID)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, result)
}

// Create 创建血管通路
func (h *VascularAccessHandler) Create(c *gin.Context) {
	patientID := c.Param("id")

	var req services.VascularAccessRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "无效的请求参数")
		return
	}

	result, err := h.service.Create(patientID, &req)
	if err != nil {
		if err.Error() == "patient not found" {
			response.NotFound(c, "患者不存在")
			return
		}
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, result)
}

// Update 更新血管通路
func (h *VascularAccessHandler) Update(c *gin.Context) {
	patientID := c.Param("id")
	accessID := c.Param("rid")

	var req services.VascularAccessRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "无效的请求参数")
		return
	}

	result, err := h.service.Update(patientID, accessID, &req)
	if err != nil {
		if err.Error() == "vascular access not found" {
			response.NotFound(c, "血管通路记录不存在")
			return
		}
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, result)
}

// Delete 删除血管通路
func (h *VascularAccessHandler) Delete(c *gin.Context) {
	patientID := c.Param("id")
	accessID := c.Param("rid")

	if err := h.service.Delete(patientID, accessID); err != nil {
		if err.Error() == "vascular access not found" {
			response.NotFound(c, "血管通路记录不存在")
			return
		}
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, gin.H{"message": "删除成功"})
}

// ListInterventions 获取患者的血管通路干预记录列表
func (h *VascularAccessHandler) ListInterventions(c *gin.Context) {
	patientID := c.Param("id")
	vascularAccessID := c.Query("vascularAccessId") // 可选，指定血管通路ID

	result, err := h.service.ListInterventions(patientID, vascularAccessID)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, result)
}

// CreateIntervention 创建血管通路干预记录
func (h *VascularAccessHandler) CreateIntervention(c *gin.Context) {
	patientID := c.Param("id")

	var req services.VascularAccessInterventionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "无效的请求参数")
		return
	}

	result, err := h.service.CreateIntervention(patientID, &req)
	if err != nil {
		if err.Error() == "vascular access not found" {
			response.NotFound(c, "血管通路记录不存在")
			return
		}
		if err.Error() == "invalid intervention date format" {
			response.BadRequest(c, "干预日期格式错误")
			return
		}
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, result)
}

// DeleteIntervention 删除血管通路干预记录
func (h *VascularAccessHandler) DeleteIntervention(c *gin.Context) {
	patientID := c.Param("id")
	interventionID := c.Param("iid")

	if err := h.service.DeleteIntervention(patientID, interventionID); err != nil {
		if err.Error() == "intervention not found" {
			response.NotFound(c, "干预记录不存在")
			return
		}
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, gin.H{"message": "删除成功"})
}

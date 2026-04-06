package v1

import (
	"github.com/elliotxin/ai-hms-backend/internal/middleware"
	"github.com/elliotxin/ai-hms-backend/internal/services"
	"github.com/elliotxin/ai-hms-backend/pkg/response"
	"github.com/gin-gonic/gin"
)

// DeviceHandler 设备处理器
type DeviceHandler struct {
	service *services.DeviceService
}

// NewDeviceHandler 创建设备处理器
func NewDeviceHandler() *DeviceHandler {
	return &DeviceHandler{service: services.NewDeviceService()}
}

// List 获取设备列表
func (h *DeviceHandler) List(c *gin.Context) {
	var req services.DeviceListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "无效的请求参数")
		return
	}

	result, err := h.service.List(req)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, gin.H{
		"items":     result.Items,
		"total":     result.Total,
		"page":      result.Page,
		"pageSize":  result.PageSize,
		"totalPage": result.TotalPage,
	})
}

// Create 创建设备
func (h *DeviceHandler) Create(c *gin.Context) {
	var req services.DeviceCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "无效的请求参数: "+err.Error())
		return
	}

	tenantId := middleware.GetTenantID(c)
	creatorId := middleware.GetCreatorID(c)

	device, err := h.service.Create(req, tenantId, creatorId)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.SuccessCreated(c, device)
}

// Update 更新设备
func (h *DeviceHandler) Update(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.BadRequest(c, "设备ID不能为空")
		return
	}

	var req services.DeviceUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "无效的请求参数")
		return
	}

	device, err := h.service.Update(id, req)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, device)
}

// Delete 删除设备
func (h *DeviceHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.BadRequest(c, "设备ID不能为空")
		return
	}

	if err := h.service.Delete(id); err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, gin.H{"message": "删除成功"})
}

// UpdateStatus 更新设备状态
func (h *DeviceHandler) UpdateStatus(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.BadRequest(c, "设备ID不能为空")
		return
	}

	var req struct {
		Status string `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "无效的请求参数")
		return
	}

	if err := h.service.UpdateStatus(id, req.Status); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, gin.H{"message": "状态已更新"})
}

// RegisterDeviceRoutes 注册设备路由
func RegisterDeviceRoutes(rg *gin.RouterGroup) {
	h := NewDeviceHandler()
	devices := rg.Group("/devices")
	{
		devices.GET("", h.List)
		devices.POST("", h.Create)
		devices.PUT("/:id", h.Update)
		devices.DELETE("/:id", h.Delete)
		devices.PUT("/:id/status", h.UpdateStatus)
	}
}

package v1

import (
	"github.com/elliotxin/ai-hms-backend/internal/middleware"
	"github.com/elliotxin/ai-hms-backend/internal/services"
	"github.com/elliotxin/ai-hms-backend/pkg/response"
	"github.com/gin-gonic/gin"
)

type DeviceHandler struct {
	service *services.DeviceService
}

func NewDeviceHandler() *DeviceHandler {
	return &DeviceHandler{service: services.NewDeviceService()}
}

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

func (h *DeviceHandler) Get(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.BadRequest(c, "设备ID不能为空")
		return
	}

	device, err := h.service.Get(id)
	if err != nil {
		if err.Error() == "device not found" {
			response.NotFound(c, "设备不存在")
			return
		}
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, device)
}

func (h *DeviceHandler) Create(c *gin.Context) {
	var req services.DeviceCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "无效的请求参数: "+err.Error())
		return
	}

	tenantID := middleware.GetTenantID(c)
	creatorID := middleware.GetCreatorID(c)
	device, err := h.service.Create(req, tenantID, creatorID)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.SuccessCreated(c, device)
}

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
		if err.Error() == "device not found" {
			response.NotFound(c, "设备不存在")
			return
		}
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, device)
}

func (h *DeviceHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.BadRequest(c, "设备ID不能为空")
		return
	}

	if err := h.service.Delete(id); err != nil {
		if err.Error() == "device not found" {
			response.NotFound(c, "设备不存在")
			return
		}
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, gin.H{"message": "删除成功"})
}

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
		if err.Error() == "device not found" {
			response.NotFound(c, "设备不存在")
			return
		}
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, gin.H{"message": "状态已更新"})
}

func (h *DeviceHandler) ListUsageLogs(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.BadRequest(c, "设备ID不能为空")
		return
	}

	var req services.DeviceLogListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "无效的请求参数")
		return
	}

	result, err := h.service.ListUsageLogs(id, req)
	if err != nil {
		if err.Error() == "device not found" {
			response.NotFound(c, "设备不存在")
			return
		}
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, result)
}

func (h *DeviceHandler) ListMaintenanceRecords(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.BadRequest(c, "设备ID不能为空")
		return
	}

	var req services.DeviceLogListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "无效的请求参数")
		return
	}

	result, err := h.service.ListMaintenanceRecords(id, req)
	if err != nil {
		if err.Error() == "device not found" {
			response.NotFound(c, "设备不存在")
			return
		}
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, result)
}

func (h *DeviceHandler) ListDisinfections(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.BadRequest(c, "设备ID不能为空")
		return
	}

	var req services.DeviceLogListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "无效的请求参数")
		return
	}

	result, err := h.service.ListDisinfectionRecords(id, req)
	if err != nil {
		if err.Error() == "device not found" {
			response.NotFound(c, "设备不存在")
			return
		}
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, result)
}

func RegisterDeviceRoutes(rg *gin.RouterGroup) {
	h := NewDeviceHandler()
	devices := rg.Group("/devices")
	{
		devices.GET("", h.List)
		devices.POST("", h.Create)
		devices.GET("/:id", h.Get)
		devices.PUT("/:id", h.Update)
		devices.DELETE("/:id", h.Delete)
		devices.PUT("/:id/status", h.UpdateStatus)
		devices.GET("/:id/usage-logs", h.ListUsageLogs)
		devices.GET("/:id/maintenance-records", h.ListMaintenanceRecords)
		devices.GET("/:id/disinfections", h.ListDisinfections)
	}
}

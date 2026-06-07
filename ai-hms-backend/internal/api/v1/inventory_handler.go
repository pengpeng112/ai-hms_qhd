package v1

import (
	"github.com/elliotxin/ai-hms-backend/internal/middleware"
	"github.com/elliotxin/ai-hms-backend/internal/services"
	"github.com/elliotxin/ai-hms-backend/pkg/response"
	"github.com/gin-gonic/gin"
)

type InventoryHandler struct {
	service *services.InventoryService
}

func NewInventoryHandler() *InventoryHandler {
	return &InventoryHandler{service: services.NewInventoryService()}
}

func (h *InventoryHandler) ListItems(c *gin.Context) {
	var req services.InventoryListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "无效的请求参数")
		return
	}
	result, err := h.service.ListItems(req)
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

func (h *InventoryHandler) CreateItem(c *gin.Context) {
	var req services.CreateInventoryItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "无效的请求参数")
		return
	}
	tenantID := middleware.GetTenantID(c)
	creatorID := middleware.GetCreatorID(c)
	item, err := h.service.CreateItem(req, tenantID, creatorID)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.SuccessCreated(c, item)
}

func (h *InventoryHandler) UpdateItem(c *gin.Context) {
	response.BadRequest(c, "库存修改暂不支持")
}

func (h *InventoryHandler) DeleteItem(c *gin.Context) {
	response.BadRequest(c, "库存删除暂不支持")
}

func (h *InventoryHandler) ListLogs(c *gin.Context) {
	var req services.StockLogRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "无效的请求参数")
		return
	}
	result, err := h.service.ListLogs(req)
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

func (h *InventoryHandler) AdjustStock(c *gin.Context) {
	response.BadRequest(c, "库存调整暂不支持")
}

func (h *InventoryHandler) ListLabelTasks(c *gin.Context) {
	response.Success(c, gin.H{"items": []interface{}{}, "total": 0, "page": 1, "pageSize": 50, "totalPage": 0})
}

func (h *InventoryHandler) CreateLabelTask(c *gin.Context) {
	response.BadRequest(c, "标签打印暂不支持")
}

func (h *InventoryHandler) UpdateLabelTaskStatus(c *gin.Context) {
	response.BadRequest(c, "标签打印暂不支持")
}

func RegisterInventoryRoutes(rg *gin.RouterGroup) {
	h := NewInventoryHandler()
	inv := rg.Group("/inventory")
	{
		inv.GET("/items", h.ListItems)
		inv.POST("/items", h.CreateItem)
		inv.PUT("/items/:id", h.UpdateItem)
		inv.DELETE("/items/:id", h.DeleteItem)
		inv.POST("/adjust", h.AdjustStock)
		inv.GET("/logs", h.ListLogs)
		inv.GET("/labels", h.ListLabelTasks)
		inv.POST("/labels", h.CreateLabelTask)
		inv.PUT("/labels/:id/status", h.UpdateLabelTaskStatus)
	}
}

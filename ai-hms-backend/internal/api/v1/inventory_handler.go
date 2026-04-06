package v1

import (
	"github.com/elliotxin/ai-hms-backend/internal/middleware"
	"github.com/elliotxin/ai-hms-backend/internal/services"
	"github.com/elliotxin/ai-hms-backend/pkg/response"
	"github.com/gin-gonic/gin"
)

// InventoryHandler 库存处理器
type InventoryHandler struct {
	service *services.InventoryService
}

func NewInventoryHandler() *InventoryHandler {
	return &InventoryHandler{service: services.NewInventoryService()}
}

// ─────────────── InventoryItem ───────────────

// ListItems GET /api/v1/inventory/items
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

// CreateItem POST /api/v1/inventory/items
func (h *InventoryHandler) CreateItem(c *gin.Context) {
	var req services.InventoryCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "无效的请求参数: "+err.Error())
		return
	}

	tenantId := middleware.GetTenantID(c)
	creatorId := middleware.GetCreatorID(c)

	item, err := h.service.CreateItem(req, tenantId, creatorId)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.SuccessCreated(c, item)
}

// UpdateItem PUT /api/v1/inventory/items/:id
func (h *InventoryHandler) UpdateItem(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.BadRequest(c, "库存ID不能为空")
		return
	}

	var req services.InventoryUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "无效的请求参数")
		return
	}

	item, err := h.service.UpdateItem(id, req)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, item)
}

// DeleteItem DELETE /api/v1/inventory/items/:id
func (h *InventoryHandler) DeleteItem(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.BadRequest(c, "库存ID不能为空")
		return
	}

	if err := h.service.DeleteItem(id); err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, gin.H{"message": "删除成功"})
}

// ─────────────── StockLog ───────────────

// ListLogs GET /api/v1/inventory/logs
func (h *InventoryHandler) ListLogs(c *gin.Context) {
	var req services.StockLogListRequest
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

// AdjustStock POST /api/v1/inventory/adjust
func (h *InventoryHandler) AdjustStock(c *gin.Context) {
	var req services.AdjustStockRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "无效的请求参数: "+err.Error())
		return
	}

	tenantId := middleware.GetTenantID(c)

	log, err := h.service.AdjustStock(req, tenantId)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.SuccessCreated(c, log)
}

// ─────────────── LabelTask ───────────────

// ListLabelTasks GET /api/v1/inventory/labels
func (h *InventoryHandler) ListLabelTasks(c *gin.Context) {
	var req services.LabelTaskListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "无效的请求参数")
		return
	}

	result, err := h.service.ListLabelTasks(req)
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

// CreateLabelTask POST /api/v1/inventory/labels
func (h *InventoryHandler) CreateLabelTask(c *gin.Context) {
	var req services.LabelTaskCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "无效的请求参数: "+err.Error())
		return
	}

	tenantId := middleware.GetTenantID(c)
	creatorId := middleware.GetCreatorID(c)

	task, err := h.service.CreateLabelTask(req, tenantId, creatorId)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.SuccessCreated(c, task)
}

// UpdateLabelTaskStatus PUT /api/v1/inventory/labels/:id/status
func (h *InventoryHandler) UpdateLabelTaskStatus(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.BadRequest(c, "标签任务ID不能为空")
		return
	}

	var req struct {
		Status string `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "无效的请求参数")
		return
	}

	if err := h.service.UpdateLabelTaskStatus(id, req.Status); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, gin.H{"message": "状态已更新"})
}

// RegisterInventoryRoutes 注册库存管理路由
func RegisterInventoryRoutes(rg *gin.RouterGroup) {
	h := NewInventoryHandler()
	inv := rg.Group("/inventory")
	{
		// 库存品目
		inv.GET("/items", h.ListItems)
		inv.POST("/items", h.CreateItem)
		inv.PUT("/items/:id", h.UpdateItem)
		inv.DELETE("/items/:id", h.DeleteItem)

		// 出入库操作 + 记录
		inv.POST("/adjust", h.AdjustStock)
		inv.GET("/logs", h.ListLogs)

		// 标签打印任务
		inv.GET("/labels", h.ListLabelTasks)
		inv.POST("/labels", h.CreateLabelTask)
		inv.PUT("/labels/:id/status", h.UpdateLabelTaskStatus)
	}
}

package v1

import (
	"net/http"
	"strconv"

	"github.com/elliotxin/ai-hms-backend/internal/services"
	"github.com/elliotxin/ai-hms-backend/pkg/response"
	"github.com/gin-gonic/gin"
)

// ===== 方案模板控制器 =====

// PlanTemplateHandler 方案模板控制器
type PlanTemplateHandler struct {
	service *services.PlanTemplateService
}

// NewPlanTemplateHandler 创建方案模板控制器
func NewPlanTemplateHandler() *PlanTemplateHandler {
	return &PlanTemplateHandler{
		service: services.NewPlanTemplateService(),
	}
}

// List 获取方案模板列表
// @Summary 获取方案模板列表
// @Tags 诊疗配置
// @Accept json
// @Produce json
// @Param page query int false "页码" default(1)
// @Param pageSize query int false "每页数量" default(20)
// @Param search query string false "搜索关键词"
// @Param mode query string false "透析模式 (HD, HDF, HP, HF, HFD)"
// @Param category query string false "分类"
// @Param isEnabled query boolean false "启用状态"
// @Success 200 {object} response.PaginationResponse
// @Router /api/v1/treatment-templates [get]
func (h *PlanTemplateHandler) List(c *gin.Context) {
	var req services.PlanTemplateListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "无效的请求参数")
		return
	}

	result, err := h.service.LegacyList(req)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Paginated(c, result.Items, result.Page, result.PageSize, result.Total)
}

// Get 获取方案模板详情
// @Summary 获取方案模板详情
// @Tags 诊疗配置
// @Accept json
// @Produce json
// @Param id path string true "模板ID"
// @Success 200 {object} models.PlanTemplate
// @Router /api/v1/treatment-templates/{id} [get]
func (h *PlanTemplateHandler) Get(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.BadRequest(c, "模板ID不能为空")
		return
	}

	template, err := h.service.LegacyGet(id)
	if err != nil {
		if err.Error() == "plan template not found" {
			response.NotFound(c, "方案模板不存在")
			return
		}
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, template)
}

// Create 创建方案模板
// @Summary 创建方案模板
// @Tags 诊疗配置
// @Accept json
// @Produce json
// @Param request body services.PlanTemplateCreateRequest true "创建方案模板请求"
// @Success 201 {object} models.PlanTemplate
// @Router /api/v1/treatment-templates [post]
func (h *PlanTemplateHandler) Create(c *gin.Context) {
	var req services.PlanTemplateCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "无效的请求参数: "+err.Error())
		return
	}

	template, err := h.service.LegacyCreate(req)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, template)
}

// Update 更新方案模板
// @Summary 更新方案模板
// @Tags 诊疗配置
// @Accept json
// @Produce json
// @Param id path string true "模板ID"
// @Param request body services.PlanTemplateUpdateRequest true "更新方案模板请求"
// @Success 200 {object} models.PlanTemplate
// @Router /api/v1/treatment-templates/{id} [put]
func (h *PlanTemplateHandler) Update(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.BadRequest(c, "模板ID不能为空")
		return
	}

	var req services.PlanTemplateUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "无效的请求参数")
		return
	}

	template, err := h.service.LegacyUpdate(id, req)
	if err != nil {
		if err.Error() == "plan template not found" {
			response.NotFound(c, "方案模板不存在")
			return
		}
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, template)
}

// Delete 删除方案模板
// @Summary 删除方案模板
// @Tags 诊疗配置
// @Accept json
// @Produce json
// @Param id path string true "模板ID"
// @Success 204
// @Router /api/v1/treatment-templates/{id} [delete]
func (h *PlanTemplateHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.BadRequest(c, "模板ID不能为空")
		return
	}

	if err := h.service.LegacyDelete(id); err != nil {
		if err.Error() == "plan template not found" {
			response.NotFound(c, "方案模板不存在")
			return
		}
		response.InternalError(c, err.Error())
		return
	}

	c.Status(http.StatusNoContent)
}

// ToggleEnabled 切换启用状态
// @Summary 切换方案模板启用状态
// @Tags 诊疗配置
// @Accept json
// @Produce json
// @Param id path string true "模板ID"
// @Success 200 {object} response.SuccessResponse
// @Router /api/v1/treatment-templates/{id}/toggle [post]
func (h *PlanTemplateHandler) ToggleEnabled(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.BadRequest(c, "模板ID不能为空")
		return
	}

	isEnabled, err := h.service.LegacyToggleEnabled(id)
	if err != nil {
		if err.Error() == "plan template not found" {
			response.NotFound(c, "方案模板不存在")
			return
		}
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, gin.H{
		"id":        id,
		"isEnabled": isEnabled,
	})
}

// SetDefault 设置默认模板
// @Summary 设置默认方案模板
// @Tags 诊疗配置
// @Accept json
// @Produce json
// @Param id path string true "模板ID"
// @Success 200 {object} response.SuccessResponse
// @Router /api/v1/treatment-templates/{id}/set-default [post]
func (h *PlanTemplateHandler) SetDefault(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.BadRequest(c, "模板ID不能为空")
		return
	}

	if err := h.service.LegacySetDefault(id); err != nil {
		if err.Error() == "plan template not found" {
			response.NotFound(c, "方案模板不存在")
			return
		}
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, gin.H{
		"id":      id,
		"message": "已设置为默认模板",
	})
}

// ===== 材料目录控制器 =====

// MaterialCatalogHandler 材料目录控制器
type MaterialCatalogHandler struct {
	service *services.MaterialCatalogService
}

// NewMaterialCatalogHandler 创建材料目录控制器
func NewMaterialCatalogHandler() *MaterialCatalogHandler {
	return &MaterialCatalogHandler{
		service: services.NewMaterialCatalogService(),
	}
}

// List 获取材料目录列表
// @Summary 获取材料目录列表
// @Tags 诊疗配置
// @Accept json
// @Produce json
// @Param page query int false "页码" default(1)
// @Param pageSize query int false "每页数量" default(20)
// @Param search query string false "搜索关键词"
// @Param category query string false "材料分类"
// @Param isEnabled query boolean false "启用状态"
// @Success 200 {object} response.PaginationResponse
// @Router /api/v1/materials/catalog [get]
func (h *MaterialCatalogHandler) List(c *gin.Context) {
	var req services.MaterialCatalogListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "无效的请求参数")
		return
	}

	result, err := h.service.LegacyList(req)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Paginated(c, result.Items, result.Page, result.PageSize, result.Total)
}

// Get 获取材料目录详情
// @Summary 获取材料目录详情
// @Tags 诊疗配置
// @Accept json
// @Produce json
// @Param id path string true "材料ID"
// @Success 200 {object} models.MaterialCatalog
// @Router /api/v1/materials/catalog/{id} [get]
func (h *MaterialCatalogHandler) Get(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.BadRequest(c, "材料ID不能为空")
		return
	}

	catalog, err := h.service.LegacyGet(id)
	if err != nil {
		if err.Error() == "material catalog not found" {
			response.NotFound(c, "材料目录不存在")
			return
		}
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, catalog)
}

// Create 创建材料目录
// @Summary 创建材料目录
// @Tags 诊疗配置
// @Accept json
// @Produce json
// @Param request body services.MaterialCatalogCreateRequest true "创建材料目录请求"
// @Success 201 {object} models.MaterialCatalog
// @Router /api/v1/materials/catalog [post]
func (h *MaterialCatalogHandler) Create(c *gin.Context) {
	var req services.MaterialCatalogCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "无效的请求参数: "+err.Error())
		return
	}
	if err := services.ValidateMaterialCatalogCreateRequest(req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	catalog, err := h.service.LegacyCreate(req)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, catalog)
}

// Update 更新材料目录
// @Summary 更新材料目录
// @Tags 诊疗配置
// @Accept json
// @Produce json
// @Param id path string true "材料ID"
// @Param request body services.MaterialCatalogUpdateRequest true "更新材料目录请求"
// @Success 200 {object} models.MaterialCatalog
// @Router /api/v1/materials/catalog/{id} [put]
func (h *MaterialCatalogHandler) Update(c *gin.Context) {
	idStr := c.Param("id")
	if idStr == "" {
		response.BadRequest(c, "材料ID不能为空")
		return
	}

	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的材料ID")
		return
	}

	var req services.MaterialCatalogUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "无效的请求参数")
		return
	}

	catalog, err := h.service.LegacyUpdate(uint(id), req)
	if err != nil {
		if err.Error() == "material catalog not found" {
			response.NotFound(c, "材料目录不存在")
			return
		}
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, catalog)
}

// Delete 删除材料目录
// @Summary 删除材料目录
// @Tags 诊疗配置
// @Accept json
// @Produce json
// @Param id path string true "材料ID"
// @Success 204
// @Router /api/v1/materials/catalog/{id} [delete]
func (h *MaterialCatalogHandler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	if idStr == "" {
		response.BadRequest(c, "材料ID不能为空")
		return
	}

	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的材料ID")
		return
	}

	if err := h.service.LegacyDelete(uint(id)); err != nil {
		if err.Error() == "material catalog not found" {
			response.NotFound(c, "材料目录不存在")
			return
		}
		response.InternalError(c, err.Error())
		return
	}

	c.Status(http.StatusNoContent)
}

// ToggleEnabled 切换启用状态
// @Summary 切换材料启用状态
// @Tags 诊疗配置
// @Accept json
// @Produce json
// @Param id path string true "材料ID"
// @Success 200 {object} response.SuccessResponse
// @Router /api/v1/materials/catalog/{id}/toggle [post]
func (h *MaterialCatalogHandler) ToggleEnabled(c *gin.Context) {
	idStr := c.Param("id")
	if idStr == "" {
		response.BadRequest(c, "材料ID不能为空")
		return
	}

	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的材料ID")
		return
	}

	isEnabled, err := h.service.LegacyToggleEnabled(uint(id))
	if err != nil {
		if err.Error() == "material catalog not found" {
			response.NotFound(c, "材料目录不存在")
			return
		}
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, gin.H{
		"id":        id,
		"isEnabled": isEnabled,
	})
}

// GetCategories 获取材料分类列表
// @Summary 获取材料分类列表
// @Tags 诊疗配置
// @Accept json
// @Produce json
// @Success 200 {object} response.SuccessResponse
// @Router /api/v1/materials/categories [get]
func (h *MaterialCatalogHandler) GetCategories(c *gin.Context) {
	categories, err := h.service.LegacyGetCategories()
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, categories)
}

// ===== 药品目录控制器 =====

// DrugCatalogHandler 药品目录控制器
type DrugCatalogHandler struct {
	service *services.DrugCatalogService
}

// NewDrugCatalogHandler 创建药品目录控制器
func NewDrugCatalogHandler() *DrugCatalogHandler {
	return &DrugCatalogHandler{
		service: services.NewDrugCatalogService(),
	}
}

// List 获取药品目录列表
// @Summary 获取药品目录列表
// @Tags 诊疗配置
// @Accept json
// @Produce json
// @Param page query int false "页码" default(1)
// @Param pageSize query int false "每页数量" default(20)
// @Param search query string false "搜索关键词"
// @Param category query string false "药品分类"
// @Param isEnabled query boolean false "启用状态"
// @Success 200 {object} response.PaginationResponse
// @Router /api/v1/drugs/catalog [get]
func (h *DrugCatalogHandler) List(c *gin.Context) {
	var req services.DrugCatalogListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "无效的请求参数")
		return
	}

	result, err := h.service.LegacyList(req)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Paginated(c, result.Items, result.Page, result.PageSize, result.Total)
}

// Get 获取药品目录详情
// @Summary 获取药品目录详情
// @Tags 诊疗配置
// @Accept json
// @Produce json
// @Param id path string true "药品ID"
// @Success 200 {object} models.DrugCatalog
// @Router /api/v1/drugs/catalog/{id} [get]
func (h *DrugCatalogHandler) Get(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.BadRequest(c, "药品ID不能为空")
		return
	}

	catalog, err := h.service.LegacyGet(id)
	if err != nil {
		if err.Error() == "drug catalog not found" {
			response.NotFound(c, "药品目录不存在")
			return
		}
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, catalog)
}

// Create 创建药品目录
// @Summary 创建药品目录
// @Tags 诊疗配置
// @Accept json
// @Produce json
// @Param request body services.DrugCatalogCreateRequest true "创建药品目录请求"
// @Success 201 {object} models.DrugCatalog
// @Router /api/v1/drugs/catalog [post]
func (h *DrugCatalogHandler) Create(c *gin.Context) {
	var req services.DrugCatalogCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "无效的请求参数: "+err.Error())
		return
	}

	catalog, err := h.service.LegacyCreate(req)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, catalog)
}

// Update 更新药品目录
// @Summary 更新药品目录
// @Tags 诊疗配置
// @Accept json
// @Produce json
// @Param id path string true "药品ID"
// @Param request body services.DrugCatalogUpdateRequest true "更新药品目录请求"
// @Success 200 {object} models.DrugCatalog
// @Router /api/v1/drugs/catalog/{id} [put]
func (h *DrugCatalogHandler) Update(c *gin.Context) {
	idStr := c.Param("id")
	if idStr == "" {
		response.BadRequest(c, "药品ID不能为空")
		return
	}

	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的药品ID")
		return
	}

	var req services.DrugCatalogUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "无效的请求参数")
		return
	}

	catalog, err := h.service.LegacyUpdate(uint(id), req)
	if err != nil {
		if err.Error() == "drug catalog not found" {
			response.NotFound(c, "药品目录不存在")
			return
		}
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, catalog)
}

// Delete 删除药品目录
// @Summary 删除药品目录
// @Tags 诊疗配置
// @Accept json
// @Produce json
// @Param id path string true "药品ID"
// @Success 204
// @Router /api/v1/drugs/catalog/{id} [delete]
func (h *DrugCatalogHandler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	if idStr == "" {
		response.BadRequest(c, "药品ID不能为空")
		return
	}

	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的药品ID")
		return
	}

	if err := h.service.LegacyDelete(uint(id)); err != nil {
		if err.Error() == "drug catalog not found" {
			response.NotFound(c, "药品目录不存在")
			return
		}
		response.InternalError(c, err.Error())
		return
	}

	c.Status(http.StatusNoContent)
}

// ToggleEnabled 切换启用状态
// @Summary 切换药品启用状态
// @Tags 诊疗配置
// @Accept json
// @Produce json
// @Param id path string true "药品ID"
// @Success 200 {object} response.SuccessResponse
// @Router /api/v1/drugs/catalog/{id}/toggle [post]
func (h *DrugCatalogHandler) ToggleEnabled(c *gin.Context) {
	idStr := c.Param("id")
	if idStr == "" {
		response.BadRequest(c, "药品ID不能为空")
		return
	}

	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的药品ID")
		return
	}

	isEnabled, err := h.service.LegacyToggleEnabled(uint(id))
	if err != nil {
		if err.Error() == "drug catalog not found" {
			response.NotFound(c, "药品目录不存在")
			return
		}
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, gin.H{
		"id":        id,
		"isEnabled": isEnabled,
	})
}

// GetCategories 获取药品分类列表
// @Summary 获取药品分类列表
// @Tags 诊疗配置
// @Accept json
// @Produce json
// @Success 200 {object} response.SuccessResponse
// @Router /api/v1/drugs/categories [get]
func (h *DrugCatalogHandler) GetCategories(c *gin.Context) {
	categories, err := h.service.LegacyGetCategories()
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, categories)
}

// ===== 医嘱模板控制器 =====

// OrderTemplateHandler 医嘱模板控制器
type OrderTemplateHandler struct {
	service *services.OrderTemplateService
}

// NewOrderTemplateHandler 创建医嘱模板控制器
func NewOrderTemplateHandler() *OrderTemplateHandler {
	return &OrderTemplateHandler{
		service: services.NewOrderTemplateService(),
	}
}

// List 获取医嘱模板列表
// @Summary 获取医嘱模板列表
// @Tags 诊疗配置
// @Accept json
// @Produce json
// @Param page query int false "页码" default(1)
// @Param pageSize query int false "每页数量" default(20)
// @Param search query string false "搜索关键词"
// @Param type query string false "医嘱类型 (长期, 临时)"
// @Param category query string false "医嘱分类"
// @Param isEnabled query boolean false "启用状态"
// @Success 200 {object} response.PaginationResponse
// @Router /api/v1/order-templates [get]
func (h *OrderTemplateHandler) List(c *gin.Context) {
	var req services.OrderTemplateListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "无效的请求参数")
		return
	}

	result, err := h.service.LegacyList(req)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Paginated(c, result.Items, result.Page, result.PageSize, result.Total)
}

// Get 获取医嘱模板详情
// @Summary 获取医嘱模板详情
// @Tags 诊疗配置
// @Accept json
// @Produce json
// @Param id path string true "模板ID"
// @Success 200 {object} models.OrderTemplate
// @Router /api/v1/order-templates/{id} [get]
func (h *OrderTemplateHandler) Get(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.BadRequest(c, "模板ID不能为空")
		return
	}

	template, err := h.service.LegacyGet(id)
	if err != nil {
		if err.Error() == "order template not found" {
			response.NotFound(c, "医嘱模板不存在")
			return
		}
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, template)
}

// Create 创建医嘱模板
// @Summary 创建医嘱模板
// @Tags 诊疗配置
// @Accept json
// @Produce json
// @Param request body services.OrderTemplateCreateRequest true "创建医嘱模板请求"
// @Success 201 {object} models.OrderTemplate
// @Router /api/v1/order-templates [post]
func (h *OrderTemplateHandler) Create(c *gin.Context) {
	var req services.OrderTemplateCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "无效的请求参数: "+err.Error())
		return
	}

	template, err := h.service.LegacyCreate(req)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, template)
}

// Update 更新医嘱模板
// @Summary 更新医嘱模板
// @Tags 诊疗配置
// @Accept json
// @Produce json
// @Param id path string true "模板ID"
// @Param request body services.OrderTemplateUpdateRequest true "更新医嘱模板请求"
// @Success 200 {object} models.OrderTemplate
// @Router /api/v1/order-templates/{id} [put]
func (h *OrderTemplateHandler) Update(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.BadRequest(c, "模板ID不能为空")
		return
	}

	var req services.OrderTemplateUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "无效的请求参数")
		return
	}

	template, err := h.service.LegacyUpdate(id, req)
	if err != nil {
		if err.Error() == "order template not found" {
			response.NotFound(c, "医嘱模板不存在")
			return
		}
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, template)
}

// Delete 删除医嘱模板
// @Summary 删除医嘱模板
// @Tags 诊疗配置
// @Accept json
// @Produce json
// @Param id path string true "模板ID"
// @Success 204
// @Router /api/v1/order-templates/{id} [delete]
func (h *OrderTemplateHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.BadRequest(c, "模板ID不能为空")
		return
	}

	if err := h.service.LegacyDelete(id); err != nil {
		if err.Error() == "order template not found" {
			response.NotFound(c, "医嘱模板不存在")
			return
		}
		response.InternalError(c, err.Error())
		return
	}

	c.Status(http.StatusNoContent)
}

// ToggleEnabled 切换启用状态
// @Summary 切换医嘱模板启用状态
// @Tags 诊疗配置
// @Accept json
// @Produce json
// @Param id path string true "模板ID"
// @Success 200 {object} response.SuccessResponse
// @Router /api/v1/order-templates/{id}/toggle [post]
func (h *OrderTemplateHandler) ToggleEnabled(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.BadRequest(c, "模板ID不能为空")
		return
	}

	isEnabled, err := h.service.LegacyToggleEnabled(id)
	if err != nil {
		if err.Error() == "order template not found" {
			response.NotFound(c, "医嘱模板不存在")
			return
		}
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, gin.H{
		"id":        id,
		"isEnabled": isEnabled,
	})
}

// SetDefault 设置默认模板
// @Summary 设置默认医嘱模板
// @Tags 诊疗配置
// @Accept json
// @Produce json
// @Param id path string true "模板ID"
// @Success 200 {object} response.SuccessResponse
// @Router /api/v1/order-templates/{id}/set-default [post]
func (h *OrderTemplateHandler) SetDefault(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.BadRequest(c, "模板ID不能为空")
		return
	}

	if err := h.service.LegacySetDefault(id); err != nil {
		if err.Error() == "order template not found" {
			response.NotFound(c, "医嘱模板不存在")
			return
		}
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, gin.H{
		"id":      id,
		"message": "已设置为默认模板",
	})
}

// ===== 注册路由 =====

// RegisterTreatmentConfigRoutes 注册诊疗配置路由
func RegisterTreatmentConfigRoutes(r *gin.RouterGroup) {
	// 方案模板
	planTemplateHandler := NewPlanTemplateHandler()
	planTemplates := r.Group("/treatment-templates")
	{
		planTemplates.GET("", planTemplateHandler.List)                // 获取方案模板列表
		planTemplates.POST("", planTemplateHandler.Create)             // 创建方案模板
		planTemplates.GET("/:id", planTemplateHandler.Get)             // 获取方案模板详情
		planTemplates.PUT("/:id", planTemplateHandler.Update)          // 更新方案模板
		planTemplates.DELETE("/:id", planTemplateHandler.Delete)       // 删除方案模板
		planTemplates.POST("/:id/toggle", planTemplateHandler.ToggleEnabled)   // 切换启用状态
		planTemplates.POST("/:id/set-default", planTemplateHandler.SetDefault) // 设置默认模板
	}

	// 材料目录
	materialCatalogHandler := NewMaterialCatalogHandler()
	materials := r.Group("/materials")
	{
		materials.GET("/catalog", materialCatalogHandler.List)           // 获取材料目录列表
		materials.POST("/catalog", materialCatalogHandler.Create)        // 创建材料目录
		materials.GET("/catalog/:id", materialCatalogHandler.Get)        // 获取材料目录详情
		materials.PUT("/catalog/:id", materialCatalogHandler.Update)     // 更新材料目录
		materials.DELETE("/catalog/:id", materialCatalogHandler.Delete)  // 删除材料目录
		materials.POST("/catalog/:id/toggle", materialCatalogHandler.ToggleEnabled) // 切换启用状态
		materials.GET("/categories", materialCatalogHandler.GetCategories) // 获取材料分类列表
	}

	// 药品目录
	drugCatalogHandler := NewDrugCatalogHandler()
	drugs := r.Group("/drugs")
	{
		drugs.GET("/catalog", drugCatalogHandler.List)           // 获取药品目录列表
		drugs.POST("/catalog", drugCatalogHandler.Create)        // 创建药品目录
		drugs.GET("/catalog/:id", drugCatalogHandler.Get)        // 获取药品目录详情
		drugs.PUT("/catalog/:id", drugCatalogHandler.Update)     // 更新药品目录
		drugs.DELETE("/catalog/:id", drugCatalogHandler.Delete)  // 删除药品目录
		drugs.POST("/catalog/:id/toggle", drugCatalogHandler.ToggleEnabled) // 切换启用状态
		drugs.GET("/categories", drugCatalogHandler.GetCategories) // 获取药品分类列表
	}

	// 医嘱模板
	orderTemplateHandler := NewOrderTemplateHandler()
	orderTemplates := r.Group("/order-templates")
	{
		orderTemplates.GET("", orderTemplateHandler.List)                // 获取医嘱模板列表
		orderTemplates.POST("", orderTemplateHandler.Create)             // 创建医嘱模板
		orderTemplates.GET("/:id", orderTemplateHandler.Get)             // 获取医嘱模板详情
		orderTemplates.PUT("/:id", orderTemplateHandler.Update)          // 更新医嘱模板
		orderTemplates.DELETE("/:id", orderTemplateHandler.Delete)       // 删除医嘱模板
		orderTemplates.POST("/:id/toggle", orderTemplateHandler.ToggleEnabled)   // 切换启用状态
		orderTemplates.POST("/:id/set-default", orderTemplateHandler.SetDefault) // 设置默认模板
	}
}

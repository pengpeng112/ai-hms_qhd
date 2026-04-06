package v1

import (
	"net/http"

	"github.com/elliotxin/ai-hms-backend/internal/models"
	"github.com/elliotxin/ai-hms-backend/internal/services"
	"github.com/elliotxin/ai-hms-backend/pkg/response"
	"github.com/gin-gonic/gin"
)

// DictHandler 字典控制器
type DictHandler struct {
	service *services.DictService
}

// NewDictHandler 创建字典控制器
func NewDictHandler() *DictHandler {
	return &DictHandler{
		service: services.NewDictService(),
	}
}

// ListTypes 获取字典类型列表
// @Summary 获取字典类型列表
// @Tags 字典管理
// @Accept json
// @Produce json
// @Success 200 {object} response.Response{data=[]models.DictType}
// @Router /api/v1/dict/types [get]
func (h *DictHandler) ListTypes(c *gin.Context) {
	result, err := h.service.ListTypes()
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, result.Items)
}

// GetType 获取字典类型详情
// @Summary 获取字典类型详情
// @Tags 字典管理
// @Accept json
// @Produce json
// @Param code path string true "字典类型代码"
// @Success 200 {object} response.Response{data=models.DictType}
// @Router /api/v1/dict/types/{code} [get]
func (h *DictHandler) GetType(c *gin.Context) {
	code := c.Param("code")
	if code == "" {
		response.BadRequest(c, "字典类型代码不能为空")
		return
	}

	result, err := h.service.GetTypeByCode(code)
	if err != nil {
		response.NotFound(c, "字典类型不存在")
		return
	}

	response.Success(c, result)
}

// GetItems 获取字典项列表
// @Summary 获取字典项列表
// @Tags 字典管理
// @Accept json
// @Produce json
// @Param typeCode path string true "字典类型代码"
// @Param isEnabled query bool false "仅启用项"
// @Success 200 {object} response.Response{data=[]models.DictItem}
// @Router /api/v1/dict/items/{typeCode} [get]
func (h *DictHandler) GetItems(c *gin.Context) {
	typeCode := c.Param("typeCode")
	if typeCode == "" {
		response.BadRequest(c, "字典类型代码不能为空")
		return
	}

	isEnabledOnly := c.DefaultQuery("isEnabled", "true") == "true"

	result, err := h.service.GetItemsByTypeCode(typeCode, isEnabledOnly)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, result.Items)
}

// GetItemsTree 获取字典项树形结构（用于级联选择）
// @Summary 获取字典项树形结构
// @Tags 字典管理
// @Accept json
// @Produce json
// @Param typeCode path string true "字典类型代码"
// @Param isEnabled query bool false "仅启用项"
// @Success 200 {object} response.Response{data=[]models.DictItem}
// @Router /api/v1/dict/items/{typeCode}/tree [get]
func (h *DictHandler) GetItemsTree(c *gin.Context) {
	typeCode := c.Param("typeCode")
	if typeCode == "" {
		response.BadRequest(c, "字典类型代码不能为空")
		return
	}

	isEnabledOnly := c.DefaultQuery("isEnabled", "true") == "true"

	result, err := h.service.GetItemsByTypeCodeTree(typeCode, isEnabledOnly)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, result)
}

// CreateType 创建字典类型
// @Summary 创建字典类型
// @Tags 字典管理
// @Accept json
// @Produce json
// @Param request body models.DictType true "字典类型信息"
// @Success 200 {object} response.Response{data=models.DictType}
// @Router /api/v1/dict/types [post]
func (h *DictHandler) CreateType(c *gin.Context) {
	var req models.DictType
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "无效的请求参数")
		return
	}

	if err := h.service.CreateType(&req); err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, req)
}

// UpdateType 更新字典类型
// @Summary 更新字典类型
// @Tags 字典管理
// @Accept json
// @Produce json
// @Param id path string true "字典类型ID"
// @Param request body map[string]interface{} true "更新内容"
// @Success 200 {object} response.Response
// @Router /api/v1/dict/types/{id} [put]
func (h *DictHandler) UpdateType(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.BadRequest(c, "字典类型ID不能为空")
		return
	}

	var raw map[string]interface{}
	if err := c.ShouldBindJSON(&raw); err != nil {
		response.BadRequest(c, "无效的请求参数")
		return
	}

	updates := sanitizeDictTypeUpdates(raw)
	if len(updates) == 0 {
		response.BadRequest(c, "没有可更新的字段")
		return
	}

	if err := h.service.UpdateType(id, updates); err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, nil)
}

// DeleteType 删除字典类型
// @Summary 删除字典类型
// @Tags 字典管理
// @Accept json
// @Produce json
// @Param id path string true "字典类型ID"
// @Success 200 {object} response.Response
// @Router /api/v1/dict/types/{id} [delete]
func (h *DictHandler) DeleteType(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.BadRequest(c, "字典类型ID不能为空")
		return
	}

	if err := h.service.DeleteType(id); err != nil {
		response.InternalError(c, err.Error())
		return
	}

	c.Status(http.StatusNoContent)
}

// CreateItem 创建字典项
// @Summary 创建字典项
// @Tags 字典管理
// @Accept json
// @Produce json
// @Param request body models.DictItem true "字典项信息"
// @Success 200 {object} response.Response{data=models.DictItem}
// @Router /api/v1/dict/items [post]
func (h *DictHandler) CreateItem(c *gin.Context) {
	var req models.DictItem
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "无效的请求参数")
		return
	}

	if err := h.service.CreateItem(&req); err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, req)
}

// UpdateItem 更新字典项
// @Summary 更新字典项
// @Tags 字典管理
// @Accept json
// @Produce json
// @Param id path string true "字典项ID"
// @Param request body map[string]interface{} true "更新内容"
// @Success 200 {object} response.Response
// @Router /api/v1/dict/items/{id} [put]
func (h *DictHandler) UpdateItem(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.BadRequest(c, "字典项ID不能为空")
		return
	}

	var raw map[string]interface{}
	if err := c.ShouldBindJSON(&raw); err != nil {
		response.BadRequest(c, "无效的请求参数")
		return
	}

	updates := sanitizeDictItemUpdates(raw)
	if len(updates) == 0 {
		response.BadRequest(c, "没有可更新的字段")
		return
	}

	if err := h.service.UpdateItem(id, updates); err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, nil)
}

// DeleteItem 删除字典项
// @Summary 删除字典项
// @Tags 字典管理
// @Accept json
// @Produce json
// @Param id path string true "字典项ID"
// @Success 200 {object} response.Response
// @Router /api/v1/dict/items/{id} [delete]
func (h *DictHandler) DeleteItem(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.BadRequest(c, "字典项ID不能为空")
		return
	}

	if err := h.service.DeleteItem(id); err != nil {
		response.InternalError(c, err.Error())
		return
	}

	c.Status(http.StatusNoContent)
}

// ToggleItemEnabled 切换字典项启用状态
// @Summary 切换字典项启用状态
// @Tags 字典管理
// @Accept json
// @Produce json
// @Param id path string true "字典项ID"
// @Success 200 {object} response.Response{data=map[string]interface{}}
// @Router /api/v1/dict/items/{id}/toggle [post]
func (h *DictHandler) ToggleItemEnabled(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.BadRequest(c, "字典项ID不能为空")
		return
	}

	if err := h.service.ToggleItemEnabled(id); err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, gin.H{"id": id})
}

// ImportDicts 批量导入字典数据
// @Summary 批量导入字典数据
// @Tags 字典管理
// @Accept json
// @Produce json
// @Param request body services.ImportData true "导入数据"
// @Success 200 {object} response.Response{data=services.ImportResult}
// @Router /api/v1/dict/import [post]
func (h *DictHandler) ImportDicts(c *gin.Context) {
	var data services.ImportData
	if err := c.ShouldBindJSON(&data); err != nil {
		response.BadRequest(c, "无效的请求参数")
		return
	}

	result, err := h.service.ImportDicts(&data)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, result)
}

// RegisterDictRoutes 注册字典管理路由
func RegisterDictRoutes(r *gin.RouterGroup) {
	dictHandler := NewDictHandler()

	// 字典类型路由
	types := r.Group("/dict/types")
	{
		types.GET("", dictHandler.ListTypes)               // 获取字典类型列表
		types.GET("/:code", dictHandler.GetType)            // 获取字典类型详情
		types.POST("", dictHandler.CreateType)              // 创建字典类型
		types.PUT("/:id", dictHandler.UpdateType)           // 更新字典类型
		types.DELETE("/:id", dictHandler.DeleteType)        // 删除字典类型
	}

	// 字典项路由
	items := r.Group("/dict/items")
	{
		// 注意：更具体的路由必须在前面注册
		items.GET("/:typeCode/tree", dictHandler.GetItemsTree)  // 获取字典项树形结构（必须在 /:typeCode 之前）
		items.GET("/:typeCode", dictHandler.GetItems)           // 获取字典项列表
		items.POST("", dictHandler.CreateItem)                  // 创建字典项
		items.PUT("/:id", dictHandler.UpdateItem)               // 更新字典项
		items.DELETE("/:id", dictHandler.DeleteItem)            // 删除字典项
		items.POST("/:id/toggle", dictHandler.ToggleItemEnabled) // 切换启用状态
	}

	// 字典数据管理
	dict := r.Group("/dict")
	{
		dict.POST("/import", dictHandler.ImportDicts) // 批量导入字典数据
	}

	// 初始化字典数据（仅开发环境 + ADMIN 角色）
	items.POST("/init", func(c *gin.Context) {
		// 生产环境保护：仅 debug 模式允许通过 API 初始化字典
		if gin.Mode() == gin.ReleaseMode {
			response.BadRequest(c, "生产环境禁止通过 API 初始化字典数据，请使用启动时自动初始化")
			return
		}
		// 初始化默认字典
		if err := dictHandler.service.InitDefaultDicts(); err != nil {
			response.InternalError(c, err.Error())
			return
		}
		// 初始化临床诊疗字典
		if err := dictHandler.service.InitClinicalDicts(); err != nil {
			response.InternalError(c, err.Error())
			return
		}
		// 初始化转归字典
		if err := dictHandler.service.InitOutcomeDicts(); err != nil {
			response.InternalError(c, err.Error())
			return
		}
		response.Success(c, gin.H{"message": "字典数据初始化成功"})
	})
}

// camelToSnake 映射表：前端 camelCase JSON key → GORM snake_case 列名
var dictTypeKeyMap = map[string]string{
	"code":        "code",
	"name":        "name",
	"description": "description",
	"icon":        "icon",
	"sortOrder":   "sort_order",
	"sort_order":  "sort_order",
	"isEnabled":   "is_enabled",
	"is_enabled":  "is_enabled",
}

var dictItemKeyMap = map[string]string{
	"code":        "code",
	"name":        "name",
	"description": "description",
	"sortOrder":   "sort_order",
	"sort_order":  "sort_order",
	"isEnabled":   "is_enabled",
	"is_enabled":  "is_enabled",
	"extra":       "extra",
	"parentCode":  "parent_code",
	"parent_code": "parent_code",
}

// sanitizeDictTypeUpdates 白名单过滤 + key 映射
func sanitizeDictTypeUpdates(raw map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	for k, v := range raw {
		if col, ok := dictTypeKeyMap[k]; ok {
			result[col] = v
		}
	}
	return result
}

// sanitizeDictItemUpdates 白名单过滤 + key 映射
func sanitizeDictItemUpdates(raw map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	for k, v := range raw {
		if col, ok := dictItemKeyMap[k]; ok {
			result[col] = v
		}
	}
	return result
}

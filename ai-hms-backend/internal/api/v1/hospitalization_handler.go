package v1

import (
	"net/http"
	"strconv"

	"github.com/elliotxin/ai-hms-backend/internal/services"
	"github.com/elliotxin/ai-hms-backend/pkg/response"
	"github.com/gin-gonic/gin"
)

// HospitalizationHandler 住院信息控制器
type HospitalizationHandler struct {
	service *services.HospitalizationService
}

// NewHospitalizationHandler 创建住院信息控制器
func NewHospitalizationHandler() *HospitalizationHandler {
	return &HospitalizationHandler{
		service: services.NewHospitalizationService(),
	}
}

// List 获取住院信息列表
// @Summary 获取住院信息列表
// @Tags 住院管理
// @Accept json
// @Produce json
// @Param page query int false "页码" default(1)
// @Param pageSize query int false "每页数量" default(20)
// @Param patientId query int false "患者ID"
// @Param status query int false "状态"
// @Param hospWard query string false "病房"
// @Success 200 {object} services.HospitalizationListResponse
// @Router /api/v1/hospitalizations [get]
func (h *HospitalizationHandler) List(c *gin.Context) {
	var req services.HospitalizationListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "无效的请求参数")
		return
	}

	result, err := h.service.List(req)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, result)
}

// Get 获取住院信息详情
// @Summary 获取住院信息详情
// @Tags 住院管理
// @Accept json
// @Produce json
// @Param id path int true "住院信息ID"
// @Success 200 {object} models.Hospitalization
// @Router /api/v1/hospitalizations/{id} [get]
func (h *HospitalizationHandler) Get(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的ID")
		return
	}

	hospitalization, err := h.service.Get(id)
	if err != nil {
		if err.Error() == "hospitalization not found" {
			response.NotFound(c, "住院信息不存在")
			return
		}
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, hospitalization)
}

// Create 创建住院信息
// @Summary 创建住院信息
// @Tags 住院管理
// @Accept json
// @Produce json
// @Param request body services.HospitalizationCreateRequest true "创建住院信息请求"
// @Success 201 {object} models.Hospitalization
// @Router /api/v1/hospitalizations [post]
func (h *HospitalizationHandler) Create(c *gin.Context) {
	var req services.HospitalizationCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "无效的请求参数")
		return
	}

	hospitalization, err := h.service.Create(req)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	c.JSON(http.StatusCreated, hospitalization)
}

// Update 更新住院信息
// @Summary 更新住院信息
// @Tags 住院管理
// @Accept json
// @Produce json
// @Param id path int true "住院信息ID"
// @Param request body services.HospitalizationUpdateRequest true "更新住院信息请求"
// @Success 200 {object} models.Hospitalization
// @Router /api/v1/hospitalizations/{id} [put]
func (h *HospitalizationHandler) Update(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的ID")
		return
	}

	var req services.HospitalizationUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "无效的请求参数")
		return
	}

	hospitalization, err := h.service.Update(id, req)
	if err != nil {
		if err.Error() == "hospitalization not found" {
			response.NotFound(c, "住院信息不存在")
			return
		}
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, hospitalization)
}

// Delete 删除住院信息
// @Summary 删除住院信息
// @Tags 住院管理
// @Accept json
// @Produce json
// @Param id path int true "住院信息ID"
// @Success 204
// @Router /api/v1/hospitalizations/{id} [delete]
func (h *HospitalizationHandler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的ID")
		return
	}

	if err := h.service.Delete(id); err != nil {
		if err.Error() == "hospitalization not found" {
			response.NotFound(c, "住院信息不存在")
			return
		}
		response.InternalError(c, err.Error())
		return
	}

	c.Status(http.StatusNoContent)
}

// GetByPatient 获取患者的当前住院信息
// @Summary 获取患者的当前住院信息
// @Tags 住院管理
// @Accept json
// @Produce json
// @Param id path int true "患者ID"
// @Success 200 {object} models.Hospitalization
// @Router /api/v1/patients/{id}/hospitalization [get]
func (h *HospitalizationHandler) GetByPatient(c *gin.Context) {
	patientIdStr := c.Param("id")
	patientId, err := strconv.ParseInt(patientIdStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的患者ID")
		return
	}

	hospitalization, err := h.service.GetByPatientId(patientId)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	if hospitalization == nil {
		response.NotFound(c, "患者当前无在院记录")
		return
	}

	response.Success(c, hospitalization)
}

// RegisterHospitalizationRoutes 注册住院信息路由
func RegisterHospitalizationRoutes(r *gin.RouterGroup) {
	handler := NewHospitalizationHandler()

	hospitalizations := r.Group("/hospitalizations")
	{
		hospitalizations.GET("", handler.List)           // 获取住院信息列表
		hospitalizations.POST("", handler.Create)        // 创建住院信息
		hospitalizations.GET("/:id", handler.Get)        // 获取住院信息详情
		hospitalizations.PUT("/:id", handler.Update)     // 更新住院信息
		hospitalizations.DELETE("/:id", handler.Delete)  // 删除住院信息
	}
}

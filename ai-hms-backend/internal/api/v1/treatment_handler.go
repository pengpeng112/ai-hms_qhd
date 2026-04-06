package v1

import (
	"strconv"
	"time"

	"github.com/elliotxin/ai-hms-backend/internal/middleware"
	"github.com/elliotxin/ai-hms-backend/internal/services"
	"github.com/elliotxin/ai-hms-backend/pkg/response"
	"github.com/gin-gonic/gin"
)

// TreatmentHandler 透析治疗控制器
type TreatmentHandler struct {
	service *services.TreatmentService
}

// NewTreatmentHandler 创建透析治疗控制器
func NewTreatmentHandler() *TreatmentHandler {
	return &TreatmentHandler{
		service: services.NewTreatmentService(),
	}
}

// List 获取治疗记录列表
// @Summary 获取治疗记录列表
// @Tags 治疗管理
// @Accept json
// @Produce json
// @Param page query int false "页码" default(1)
// @Param pageSize query int false "每页数量" default(20)
// @Param patientId query int false "患者ID"
// @Param status query int false "状态"
// @Param type query int false "治疗类型"
// @Param treatmentDate query string false "治疗日期 (YYYY-MM-DD)"
// @Success 200 {object} services.TreatmentListResponse
// @Router /api/v1/treatments [get]
func (h *TreatmentHandler) List(c *gin.Context) {
	var req services.TreatmentListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "无效的请求参数")
		return
	}

	// 解析日期参数
	if dateStr := c.Query("treatmentDate"); dateStr != "" {
		if t, err := time.Parse("2006-01-02", dateStr); err == nil {
			req.TreatmentDate = &t
		}
	}
	if dateStr := c.Query("treatmentDateStart"); dateStr != "" {
		if t, err := time.Parse("2006-01-02", dateStr); err == nil {
			req.TreatmentDateStart = &t
		}
	}
	if dateStr := c.Query("treatmentDateEnd"); dateStr != "" {
		if t, err := time.Parse("2006-01-02", dateStr); err == nil {
			req.TreatmentDateEnd = &t
		}
	}

	result, err := h.service.List(req)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, result)
}

// Get 获取治疗记录详情
// @Summary 获取治疗记录详情
// @Tags 治疗管理
// @Accept json
// @Produce json
// @Param id path int true "治疗记录ID"
// @Success 200 {object} models.Treatment
// @Router /api/v1/treatments/{id} [get]
func (h *TreatmentHandler) Get(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的ID")
		return
	}

	treatment, err := h.service.Get(id)
	if err != nil {
		if err.Error() == "treatment not found" {
			response.NotFound(c, "治疗记录不存在")
			return
		}
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, treatment)
}

// Create 创建治疗记录
// @Summary 创建治疗记录
// @Tags 治疗管理
// @Accept json
// @Produce json
// @Param request body services.TreatmentCreateRequest true "创建治疗记录请求"
// @Success 201 {object} models.Treatment
// @Router /api/v1/treatments [post]
func (h *TreatmentHandler) Create(c *gin.Context) {
	var req services.TreatmentCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "无效的请求参数")
		return
	}

	tenantId := middleware.GetTenantID(c)
	creatorId := middleware.GetCreatorID(c)

	treatment, err := h.service.Create(req, tenantId, creatorId)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.SuccessCreated(c, treatment)
}

// Update 更新治疗记录
// @Summary 更新治疗记录
// @Tags 治疗管理
// @Accept json
// @Produce json
// @Param id path int true "治疗记录ID"
// @Param request body services.TreatmentUpdateRequest true "更新治疗记录请求"
// @Success 200 {object} models.Treatment
// @Router /api/v1/treatments/{id} [put]
func (h *TreatmentHandler) Update(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的ID")
		return
	}

	var req services.TreatmentUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "无效的请求参数")
		return
	}

	treatment, err := h.service.Update(id, req)
	if err != nil {
		if err.Error() == "treatment not found" {
			response.NotFound(c, "治疗记录不存在")
			return
		}
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, treatment)
}

// Delete 删除治疗记录
// @Summary 删除治疗记录
// @Tags 治疗管理
// @Accept json
// @Produce json
// @Param id path int true "治疗记录ID"
// @Success 200
// @Router /api/v1/treatments/{id} [delete]
func (h *TreatmentHandler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的ID")
		return
	}

	if err := h.service.Delete(id); err != nil {
		if err.Error() == "treatment not found" {
			response.NotFound(c, "治疗记录不存在")
			return
		}
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, gin.H{"message": "删除成功"})
}

// UpdateStatus 更新治疗状态
// @Summary 更新治疗状态
// @Tags 治疗管理
// @Accept json
// @Produce json
// @Param id path int true "治疗记录ID"
// @Param request body map[string]int true "状态" Example({"status": 1})
// @Success 200
// @Router /api/v1/treatments/{id}/status [put]
func (h *TreatmentHandler) UpdateStatus(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的ID")
		return
	}

	var req struct {
		Status int `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "无效的请求参数")
		return
	}

	if err := h.service.UpdateStatus(id, req.Status); err != nil {
		if err.Error() == "treatment not found" {
			response.NotFound(c, "治疗记录不存在")
			return
		}
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, gin.H{"message": "状态更新成功"})
}

// GetByPatientAndDate 获取患者在指定日期的治疗记录
// @Summary 获取患者在指定日期的治疗记录
// @Tags 治疗管理
// @Accept json
// @Produce json
// @Param id path int true "患者ID"
// @Param date query string true "治疗日期 (YYYY-MM-DD)"
// @Success 200 {object} models.Treatment
// @Router /api/v1/patients/{id}/treatment [get]
func (h *TreatmentHandler) GetByPatientAndDate(c *gin.Context) {
	patientIdStr := c.Param("id")
	patientId, err := strconv.ParseInt(patientIdStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的患者ID")
		return
	}

	dateStr := c.Query("date")
	if dateStr == "" {
		response.BadRequest(c, "请提供治疗日期")
		return
	}

	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		response.BadRequest(c, "无效的日期格式，请使用 YYYY-MM-DD 格式")
		return
	}

	treatment, err := h.service.GetByPatientAndDate(patientId, date)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	if treatment == nil {
		response.NotFound(c, "该日期没有治疗记录")
		return
	}

	response.Success(c, treatment)
}

// RegisterTreatmentRoutes 注册治疗管理路由
func RegisterTreatmentRoutes(r *gin.RouterGroup) {
	handler := NewTreatmentHandler()

	treatments := r.Group("/treatments")
	{
		treatments.GET("", handler.List)                    // 获取治疗记录列表
		treatments.POST("", handler.Create)                 // 创建治疗记录
		treatments.GET("/:id", handler.Get)                 // 获取治疗记录详情
		treatments.PUT("/:id", handler.Update)              // 更新治疗记录
		treatments.DELETE("/:id", handler.Delete)           // 删除治疗记录
		treatments.PUT("/:id/status", handler.UpdateStatus) // 更新治疗状态
	}

	// 患者相关的治疗查询
	r.GET("/patients/:id/treatment", handler.GetByPatientAndDate)
}

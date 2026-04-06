package v1

import (
	"net/http"
	"strings"

	"github.com/elliotxin/ai-hms-backend/internal/middleware"
	"github.com/elliotxin/ai-hms-backend/internal/services"
	"github.com/elliotxin/ai-hms-backend/pkg/response"
	"github.com/gin-gonic/gin"
)

// PatientHandler 患者控制器
type PatientHandler struct {
	service *services.PatientService
}

// NewPatientHandler 创建患者控制器
func NewPatientHandler() *PatientHandler {
	return &PatientHandler{
		service: services.NewPatientService(),
	}
}

// List 获取患者列表
// @Summary 获取患者列表
// @Tags 患者管理
// @Accept json
// @Produce json
// @Param page query int false "页码" default(1)
// @Param pageSize query int false "每页数量" default(20)
// @Param status query string false "患者状态"
// @Param bedNumber query string false "床位号"
// @Param name query string false "患者姓名"
// @Param riskLevel query string false "风险等级"
// @Success 200 {object} response.SuccessResponse
// @Router /api/v1/patients [get]
func (h *PatientHandler) List(c *gin.Context) {
	var req services.ListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "无效的请求参数")
		return
	}

	result, err := h.service.List(req)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	// 使用标准分页响应格式
	response.Paginated(c, result.Items, result.Page, result.PageSize, result.Total)
}

// Get 获取患者详情
// @Summary 获取患者详情
// @Tags 患者管理
// @Accept json
// @Produce json
// @Param id path string true "患者ID"
// @Success 200 {object} models.Patient
// @Router /api/v1/patients/{id} [get]
func (h *PatientHandler) Get(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.BadRequest(c, "患者ID不能为空")
		return
	}

	patient, err := h.service.Get(id)
	if err != nil {
		if err.Error() == "patient not found" {
			response.NotFound(c, "患者不存在")
			return
		}
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, patient)
}

// Create 创建患者
// @Summary 创建患者
// @Tags 患者管理
// @Accept json
// @Produce json
// @Param request body services.CreateRequest true "创建患者请求"
// @Success 201 {object} models.Patient
// @Router /api/v1/patients [post]
func (h *PatientHandler) Create(c *gin.Context) {
	var req services.CreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "无效的请求参数: "+err.Error())
		return
	}

	creatorID := middleware.GetUserID(c)
	if strings.TrimSpace(creatorID) == "" {
		response.Unauthorized(c, "未获取到用户身份")
		return
	}

	tenantID := creatorID
	if rawTenantID, exists := c.Get("tenant_id"); exists {
		if t, ok := rawTenantID.(string); ok && strings.TrimSpace(t) != "" {
			tenantID = t
		}
	}

	patient, err := h.service.Create(req, tenantID, creatorID)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, patient)
}

// Update 更新患者
// @Summary 更新患者
// @Tags 患者管理
// @Accept json
// @Produce json
// @Param id path string true "患者ID"
// @Param request body services.UpdateRequest true "更新患者请求"
// @Success 200 {object} models.Patient
// @Router /api/v1/patients/{id} [put]
func (h *PatientHandler) Update(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.BadRequest(c, "患者ID不能为空")
		return
	}

	var req services.UpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "无效的请求参数")
		return
	}

	patient, err := h.service.Update(id, req)
	if err != nil {
		if err.Error() == "patient not found" {
			response.NotFound(c, "患者不存在")
			return
		}
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, patient)
}

// Delete 删除患者
// @Summary 删除患者
// @Tags 患者管理
// @Accept json
// @Produce json
// @Param id path string true "患者ID"
// @Success 204
// @Router /api/v1/patients/{id} [delete]
func (h *PatientHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.BadRequest(c, "患者ID不能为空")
		return
	}

	if err := h.service.Delete(id); err != nil {
		if err.Error() == "patient not found" {
			response.NotFound(c, "患者不存在")
			return
		}
		response.InternalError(c, err.Error())
		return
	}

	c.Status(http.StatusNoContent)
}

// Me 获取当前用户信息
// @Summary 获取当前用户信息
// @Tags 用户
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/me [get]
func (h *PatientHandler) Me(c *gin.Context) {
	response.Success(c, gin.H{
		"user_id":  middleware.GetUserID(c),
		"username": middleware.GetUsername(c),
		"roles":    middleware.GetRoles(c),
	})
}

// GetTreatmentPlans 获取患者的所有治疗方案
// @Summary 获取患者的所有治疗方案
// @Tags 患者管理
// @Accept json
// @Produce json
// @Param id path string true "患者ID"
// @Success 200 {object} []models.TreatmentPlan
// @Router /api/v1/patients/{id}/treatment-plans [get]
func (h *PatientHandler) GetTreatmentPlans(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.BadRequest(c, "患者ID不能为空")
		return
	}

	plans, err := h.service.GetTreatmentPlans(id)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, plans)
}

// GetTreatmentPlan 获取患者治疗方案
// @Summary 获取患者治疗方案
// @Tags 患者管理
// @Accept json
// @Produce json
// @Param id path string true "患者ID"
// @Param mode query string false "透析模式（HD, HDF, HF, HP, HFD）"
// @Success 200 {object} models.TreatmentPlan
// @Router /api/v1/patients/{id}/treatment-plan [get]
func (h *PatientHandler) GetTreatmentPlan(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.BadRequest(c, "患者ID不能为空")
		return
	}

	mode := c.Query("mode") // 可选的模式参数

	plan, err := h.service.GetTreatmentPlan(id, mode)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	if plan == nil {
		response.NotFound(c, "治疗方案不存在")
		return
	}

	response.Success(c, plan)
}

// CreateTreatmentPlan 创建患者治疗方案
// @Summary 创建患者治疗方案
// @Tags 患者管理
// @Accept json
// @Produce json
// @Param id path string true "患者ID"
// @Param request body services.CreateTreatmentPlanRequest true "创建治疗方案请求"
// @Success 201 {object} models.TreatmentPlan
// @Router /api/v1/patients/{id}/treatment-plan [post]
func (h *PatientHandler) CreateTreatmentPlan(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.BadRequest(c, "患者ID不能为空")
		return
	}

	var req services.CreateTreatmentPlanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "无效的请求参数: "+err.Error())
		return
	}

	// 设置默认值
	if req.Status == "" {
		req.Status = "启用"
	}
	if req.WeeklyFrequency == 0 {
		req.WeeklyFrequency = 3
	}
	if req.Duration == 0 {
		req.Duration = 4
	}

	plan, err := h.service.CreateTreatmentPlan(id, req)
	if err != nil {
		if err.Error() == "patient not found" {
			response.NotFound(c, "患者不存在")
			return
		}
		if err.Error() == "treatment plan already exists for this patient" {
			response.BadRequest(c, "该患者已存在治疗方案")
			return
		}
		response.InternalError(c, err.Error())
		return
	}

	response.SuccessCreated(c, plan)
}

// UpdateTreatmentPlan 更新患者治疗方案
// @Summary 更新患者治疗方案
// @Tags 患者管理
// @Accept json
// @Produce json
// @Param id path string true "患者ID"
// @Param request body services.UpdateTreatmentPlanRequest true "更新治疗方案请求"
// @Success 200 {object} models.TreatmentPlan
// @Router /api/v1/patients/{id}/treatment-plan [put]
func (h *PatientHandler) UpdateTreatmentPlan(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.BadRequest(c, "患者ID不能为空")
		return
	}

	var req services.UpdateTreatmentPlanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "无效的请求参数: "+err.Error())
		return
	}

	plan, err := h.service.UpdateTreatmentPlan(id, req)
	if err != nil {
		if err.Error() == "treatment plan not found" {
			response.NotFound(c, "治疗方案不存在")
			return
		}
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, plan)
}

// DeleteTreatmentPlan 删除患者治疗方案
// @Summary 删除患者治疗方案
// @Tags 患者管理
// @Accept json
// @Produce json
// @Param id path string true "患者ID"
// @Success 204
// @Router /api/v1/patients/{id}/treatment-plan [delete]
func (h *PatientHandler) DeleteTreatmentPlan(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.BadRequest(c, "患者ID不能为空")
		return
	}

	if err := h.service.DeleteTreatmentPlan(id); err != nil {
		response.InternalError(c, err.Error())
		return
	}

	c.Status(http.StatusNoContent)
}

// GetAdjustmentRecords 获取患者方案调整记录列表
func (h *PatientHandler) GetAdjustmentRecords(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.BadRequest(c, "患者ID不能为空")
		return
	}

	records, err := h.service.GetAdjustmentRecords(id)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, records)
}

// CreateAdjustmentRecord 创建患者方案调整记录
func (h *PatientHandler) CreateAdjustmentRecord(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.BadRequest(c, "患者ID不能为空")
		return
	}

	var req services.CreateAdjustmentRecordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "无效的请求参数: "+err.Error())
		return
	}

	// 如果未指定操作者，尝试从 token 获取
	if req.Operator == "" {
		if username, exists := c.Get("username"); exists {
			if name, ok := username.(string); ok {
				req.Operator = name
			}
		}
	}

	record, err := h.service.CreateAdjustmentRecord(id, req)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, record)
}

// RegisterPatientRoutes 注册患者路由
func RegisterPatientRoutes(r *gin.RouterGroup) {
	handler := NewPatientHandler()
	hospitalizationHandler := NewHospitalizationHandler()

	patients := r.Group("/patients")
	{
		patients.GET("", handler.List)    // 获取患者列表
		patients.POST("", handler.Create) // 创建患者
		// 注意：带子路径的路由必须在 :id 之前注册，避免 Gin 路由冲突
		patients.GET("/:id/hospitalization", hospitalizationHandler.GetByPatient) // 患者的当前住院信息
		patients.GET("/:id/treatment-plan", handler.GetTreatmentPlan)             // 获取患者治疗方案
		patients.POST("/:id/treatment-plan", handler.CreateTreatmentPlan)         // 创建患者治疗方案
		patients.PUT("/:id/treatment-plan", handler.UpdateTreatmentPlan)          // 更新患者治疗方案
		patients.DELETE("/:id/treatment-plan", handler.DeleteTreatmentPlan)       // 删除患者治疗方案
		patients.GET("/:id", handler.Get)                                         // 获取患者详情
		patients.PUT("/:id", handler.Update)                                      // 更新患者
		patients.DELETE("/:id", handler.Delete)                                   // 删除患者
	}
}

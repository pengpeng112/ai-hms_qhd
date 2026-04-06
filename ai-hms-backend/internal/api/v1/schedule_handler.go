package v1

import (
	"net/http"
	"strconv"
	"time"

	"github.com/elliotxin/ai-hms-backend/internal/middleware"
	"github.com/elliotxin/ai-hms-backend/internal/services"
	"github.com/elliotxin/ai-hms-backend/pkg/response"
	"github.com/gin-gonic/gin"
)

// ShiftHandler 班次控制器
type ShiftHandler struct {
	service *services.ShiftService
}

// NewShiftHandler 创建班次控制器
func NewShiftHandler() *ShiftHandler {
	return &ShiftHandler{
		service: services.NewShiftService(),
	}
}

// List 获取班次列表
// @Summary 获取班次列表
// @Tags 排班管理
// @Accept json
// @Produce json
// @Success 200 {array} models.Shift
// @Router /api/v1/shifts [get]
func (h *ShiftHandler) List(c *gin.Context) {
	shifts, err := h.service.List()
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, shifts)
}

// Get 获取班次详情
// @Summary 获取班次详情
// @Tags 排班管理
// @Accept json
// @Produce json
// @Param id path int true "班次ID"
// @Success 200 {object} models.Shift
// @Router /api/v1/shifts/{id} [get]
func (h *ShiftHandler) Get(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的ID")
		return
	}

	shift, err := h.service.Get(id)
	if err != nil {
		if err.Error() == "shift not found" {
			response.NotFound(c, "班次不存在")
			return
		}
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, shift)
}

// Create 创建班次
// @Summary 创建班次
// @Tags 排班管理
// @Accept json
// @Produce json
// @Param request body services.ShiftCreateRequest true "创建班次请求"
// @Success 201 {object} models.Shift
// @Router /api/v1/shifts [post]
func (h *ShiftHandler) Create(c *gin.Context) {
	var req services.ShiftCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "无效的请求参数")
		return
	}

	tenantId := middleware.GetTenantID(c)
	creatorId := middleware.GetCreatorID(c)

	shift, err := h.service.Create(req, tenantId, creatorId)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	c.JSON(http.StatusCreated, shift)
}

// Update 更新班次
// @Summary 更新班次
// @Tags 排班管理
// @Accept json
// @Produce json
// @Param id path int true "班次ID"
// @Param request body services.ShiftUpdateRequest true "更新班次请求"
// @Success 200 {object} models.Shift
// @Router /api/v1/shifts/{id} [put]
func (h *ShiftHandler) Update(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的ID")
		return
	}

	var req services.ShiftUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "无效的请求参数")
		return
	}

	shift, err := h.service.Update(id, req)
	if err != nil {
		if err.Error() == "shift not found" {
			response.NotFound(c, "班次不存在")
			return
		}
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, shift)
}

// Delete 删除班次
// @Summary 删除班次
// @Tags 排班管理
// @Accept json
// @Produce json
// @Param id path int true "班次ID"
// @Success 204
// @Router /api/v1/shifts/{id} [delete]
func (h *ShiftHandler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的ID")
		return
	}

	if err := h.service.Delete(id); err != nil {
		if err.Error() == "shift not found" {
			response.NotFound(c, "班次不存在")
			return
		}
		response.InternalError(c, err.Error())
		return
	}

	c.Status(http.StatusNoContent)
}

// PatientShiftHandler 患者排班控制器
type PatientShiftHandler struct {
	service *services.PatientShiftService
}

// NewPatientShiftHandler 创建患者排班控制器
func NewPatientShiftHandler() *PatientShiftHandler {
	return &PatientShiftHandler{
		service: services.NewPatientShiftService(),
	}
}

// List 获取患者排班列表
// @Summary 获取患者排班列表
// @Tags 排班管理
// @Accept json
// @Produce json
// @Param page query int false "页码" default(1)
// @Param pageSize query int false "每页数量" default(20)
// @Param patientId query int false "患者ID"
// @Param shiftId query int false "班次ID"
// @Param wardId query int false "病房ID"
// @Param bedId query int false "床位ID"
// @Param status query int false "状态"
// @Success 200 {object} services.PatientShiftListResponse
// @Router /api/v1/patient-shifts [get]
func (h *PatientShiftHandler) List(c *gin.Context) {
	var req services.PatientShiftListRequest
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

// Get 获取患者排班详情
// @Summary 获取患者排班详情
// @Tags 排班管理
// @Accept json
// @Produce json
// @Param id path int true "排班ID"
// @Success 200 {object} models.PatientShift
// @Router /api/v1/patient-shifts/{id} [get]
func (h *PatientShiftHandler) Get(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的ID")
		return
	}

	patientShift, err := h.service.Get(id)
	if err != nil {
		if err.Error() == "patient shift not found" {
			response.NotFound(c, "排班记录不存在")
			return
		}
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, patientShift)
}

// Create 创建患者排班
// @Summary 创建患者排班
// @Tags 排班管理
// @Accept json
// @Produce json
// @Param request body services.PatientShiftCreateRequest true "创建患者排班请求"
// @Success 201 {object} models.PatientShift
// @Router /api/v1/patient-shifts [post]
func (h *PatientShiftHandler) Create(c *gin.Context) {
	var req services.PatientShiftCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "无效的请求参数")
		return
	}

	// 检查冲突
	hasConflict, err := h.service.CheckConflict(
		req.PatientId,
		req.ScheduleDate,
		req.ShiftId,
		nil,
	)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	if hasConflict {
		response.BadRequest(c, "该患者在该日期的该班次已有排班")
		return
	}

	tenantId := middleware.GetTenantID(c)
	creatorId := middleware.GetCreatorID(c)

	patientShift, err := h.service.Create(req, tenantId, creatorId)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	c.JSON(http.StatusCreated, patientShift)
}

// Update 更新患者排班
// @Summary 更新患者排班
// @Tags 排班管理
// @Accept json
// @Produce json
// @Param id path int true "排班ID"
// @Param request body services.PatientShiftUpdateRequest true "更新患者排班请求"
// @Success 200 {object} models.PatientShift
// @Router /api/v1/patient-shifts/{id} [put]
func (h *PatientShiftHandler) Update(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的ID")
		return
	}

	var req services.PatientShiftUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "无效的请求参数")
		return
	}

	patientShift, err := h.service.Update(id, req)
	if err != nil {
		if err.Error() == "patient shift not found" {
			response.NotFound(c, "排班记录不存在")
			return
		}
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, patientShift)
}

// Delete 删除患者排班
// @Summary 删除患者排班
// @Tags 排班管理
// @Accept json
// @Produce json
// @Param id path int true "排班ID"
// @Success 204
// @Router /api/v1/patient-shifts/{id} [delete]
func (h *PatientShiftHandler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的ID")
		return
	}

	if err := h.service.Delete(id); err != nil {
		if err.Error() == "patient shift not found" {
			response.NotFound(c, "排班记录不存在")
			return
		}
		response.InternalError(c, err.Error())
		return
	}

	c.Status(http.StatusNoContent)
}

// GetByPatientAndDate 根据患者ID和日期获取排班
// @Summary 根据患者ID和日期获取排班
// @Tags 排班管理
// @Accept json
// @Produce json
// @Param id path int true "患者ID"
// @Param date query string true "日期 (YYYY-MM-DD)"
// @Success 200 {object} models.PatientShift
// @Router /api/v1/patients/{id}/shift [get]
func (h *PatientShiftHandler) GetByPatientAndDate(c *gin.Context) {
	patientIdStr := c.Param("id")
	patientId, err := strconv.ParseInt(patientIdStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的患者ID")
		return
	}

	dateStr := c.Query("date")
	if dateStr == "" {
		response.BadRequest(c, "日期不能为空")
		return
	}

	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		response.BadRequest(c, "无效的日期格式，请使用 YYYY-MM-DD")
		return
	}

	patientShift, err := h.service.GetByPatientAndDate(patientId, date)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	if patientShift == nil {
		response.NotFound(c, "该患者该日期无排班记录")
		return
	}

	response.Success(c, patientShift)
}

// RegisterScheduleRoutes 注册排班管理路由
func RegisterScheduleRoutes(r *gin.RouterGroup) {
	shiftHandler := NewShiftHandler()
	patientShiftHandler := NewPatientShiftHandler()

	// 班次路由
	shifts := r.Group("/shifts")
	{
		shifts.GET("", shiftHandler.List)           // 获取班次列表
		shifts.POST("", shiftHandler.Create)        // 创建班次
		shifts.GET("/:id", shiftHandler.Get)         // 获取班次详情
		shifts.PUT("/:id", shiftHandler.Update)      // 更新班次
		shifts.DELETE("/:id", shiftHandler.Delete)   // 删除班次
	}

	// 患者排班路由
	patientShifts := r.Group("/patient-shifts")
	{
		patientShifts.GET("", patientShiftHandler.List)           // 获取患者排班列表
		patientShifts.POST("", patientShiftHandler.Create)        // 创建患者排班
		patientShifts.GET("/:id", patientShiftHandler.Get)         // 获取患者排班详情
		patientShifts.PUT("/:id", patientShiftHandler.Update)      // 更新患者排班
		patientShifts.DELETE("/:id", patientShiftHandler.Delete)   // 删除患者排班
	}

	// 患者相关的排班查询
	r.GET("/patients/:id/shift", patientShiftHandler.GetByPatientAndDate)
}

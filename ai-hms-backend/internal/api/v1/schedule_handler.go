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
	tenantId := middleware.GetTenantID(c)
	shifts, err := h.service.List(tenantId)
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

	tenantId := middleware.GetTenantID(c)
	shift, err := h.service.Get(id, tenantId)
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

	tenantId := middleware.GetTenantID(c)
	shift, err := h.service.Update(id, tenantId, req)
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

	tenantId := middleware.GetTenantID(c)
	if err := h.service.Delete(id, tenantId); err != nil {
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

	req.TenantId = middleware.GetTenantID(c)

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

	tenantId := middleware.GetTenantID(c)
	patientShift, err := h.service.Get(id, tenantId)
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

	// parse date for conflict check
	scheduleDate, err := services.ParseScheduleDate(req.ScheduleDate)
	if err != nil {
		response.BadRequest(c, "无效的日期格式，应为 YYYY-MM-DD")
		return
	}

	// 检查冲突
	tenantId := middleware.GetTenantID(c)
	hasConflict, err := h.service.CheckConflict(
		req.PatientId,
		tenantId,
		scheduleDate,
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

	if req.BedId != nil {
		hasBedConflict, err := h.service.CheckBedConflict(*req.BedId, tenantId, scheduleDate, req.ShiftId, nil)
		if err != nil {
			response.InternalError(c, err.Error())
			return
		}
		if hasBedConflict {
			response.BadRequest(c, "该床位该班次已被占用")
			return
		}
	}

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

	tenantId := middleware.GetTenantID(c)

	// P0-7: 修改前先校验
	existingShift, err := h.service.Get(id, tenantId)
	if err != nil {
		if err.Error() == "patient shift not found" {
			response.NotFound(c, "排班记录不存在")
			return
		}
		response.InternalError(c, err.Error())
		return
	}
	if err := h.service.ValidateShiftEdit(existingShift, tenantId); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	patientShift, err := h.service.Update(id, tenantId, req)
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

	tenantId := middleware.GetTenantID(c)

	// P0-7: 取消前先校验
	shift, err := h.service.Get(id, tenantId)
	if err != nil {
		if err.Error() == "patient shift not found" {
			response.NotFound(c, "排班记录不存在")
			return
		}
		response.InternalError(c, err.Error())
		return
	}
	if err := h.service.ValidateShiftEdit(shift, tenantId); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if err := h.service.Delete(id, tenantId); err != nil {
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

	tenantId := middleware.GetTenantID(c)
	patientShift, err := h.service.GetByPatientAndDate(patientId, tenantId, date)
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

// MoveRequest 换床请求
type MoveRequest struct {
	BedID         *int64    `json:"bedId"`
	WardID        *int64    `json:"wardId"`
	TreatmentTime *string   `json:"treatmentTime"`
	ShiftID       *int64    `json:"shiftId"`
}

// Move 换床/移动排班（支持跨日期/跨班次）
func (h *PatientShiftHandler) Move(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的ID")
		return
	}

	var req MoveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "无效的请求参数")
		return
	}

	tenantId := middleware.GetTenantID(c)
	patientShift, err := h.service.Get(id, tenantId)
	if err != nil {
		if err.Error() == "patient shift not found" {
			response.NotFound(c, "排班记录不存在")
			return
		}
		response.InternalError(c, err.Error())
		return
	}

	// P0-7 校验：历史保护/已过班次保护/治疗中保护
	if err := h.service.ValidateShiftEdit(patientShift, tenantId); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	// 计算目标值
	targetDate := patientShift.ScheduleDate
	targetShiftID := patientShift.ShiftId
	targetBedID := patientShift.BedId
	targetWardID := patientShift.WardId

	if req.TreatmentTime != nil {
		parsed, err := services.ParseScheduleDate(*req.TreatmentTime)
		if err != nil {
			response.BadRequest(c, "无效的 treatmentTime 格式，应为 YYYY-MM-DD")
			return
		}
		targetDate = parsed
	}
	if req.ShiftID != nil {
		targetShiftID = *req.ShiftID
	}
	if req.BedID != nil {
		targetBedID = req.BedID
	}
	if req.WardID != nil {
		targetWardID = req.WardID
	}

	// 冲突校验
	if targetBedID != nil {
		hasConflict, err := h.service.CheckBedConflict(*targetBedID, tenantId, targetDate, targetShiftID, &id)
		if err != nil {
			response.InternalError(c, err.Error())
			return
		}
		if hasConflict {
			response.BadRequest(c, "目标床位该班次已被占用")
			return
		}
	}

	updateReq := services.PatientShiftUpdateRequest{
		BedId:         targetBedID,
		WardId:        targetWardID,
		ShiftId:       &targetShiftID,
		TreatmentTime: &targetDate,
	}
	updated, err := h.service.Update(id, tenantId, updateReq)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, updated)
}

// SwapRequest 互换请求
type SwapRequest struct {
	SourceID int64 `json:"sourceId"`
	TargetID int64 `json:"targetId"`
}

// Swap 互换排班（事务 + 完整交换）
func (h *PatientShiftHandler) Swap(c *gin.Context) {
	var req SwapRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "无效的请求参数")
		return
	}

	tenantId := middleware.GetTenantID(c)

	// P0-7: 互换前校验双方
	sourceShift, err := h.service.Get(req.SourceID, tenantId)
	if err != nil {
		response.NotFound(c, "源排班记录不存在")
		return
	}
	targetShift, err := h.service.Get(req.TargetID, tenantId)
	if err != nil {
		response.NotFound(c, "目标排班记录不存在")
		return
	}
	if err := h.service.ValidateShiftEdit(sourceShift, tenantId); err != nil {
		response.BadRequest(c, "源排班: "+err.Error())
		return
	}
	if err := h.service.ValidateShiftEdit(targetShift, tenantId); err != nil {
		response.BadRequest(c, "目标排班: "+err.Error())
		return
	}

	if err := h.service.Swap(req.SourceID, req.TargetID, tenantId); err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, gin.H{"success": true})
}

// RegisterScheduleRoutes 注册排班管理路由
func RegisterScheduleRoutes(r *gin.RouterGroup) {
	shiftHandler := NewShiftHandler()
	patientShiftHandler := NewPatientShiftHandler()
	weekHandler := NewScheduleWeekHandler()

	// 班次路由
	shifts := r.Group("/shifts")
	{
		shifts.GET("", shiftHandler.List)
		shifts.POST("", shiftHandler.Create)
		shifts.GET("/:id", shiftHandler.Get)
		shifts.PUT("/:id", shiftHandler.Update)
		shifts.DELETE("/:id", shiftHandler.Delete)
	}

	// 患者排班路由
	patientShifts := r.Group("/patient-shifts")
	{
		patientShifts.GET("", patientShiftHandler.List)
		patientShifts.POST("", patientShiftHandler.Create)
		patientShifts.GET("/:id", patientShiftHandler.Get)
		patientShifts.PUT("/:id", patientShiftHandler.Update)
		patientShifts.DELETE("/:id", patientShiftHandler.Delete)
		patientShifts.POST("/:id/move", patientShiftHandler.Move)
		patientShifts.POST("/swap", patientShiftHandler.Swap)
	}

	// 患者相关的排班查询
	r.GET("/patients/:id/shift", patientShiftHandler.GetByPatientAndDate)

	// 周视图聚合
	r.GET("/schedule/week", weekHandler.GetWeek)
}

// ScheduleWeekHandler 周视图聚合
type ScheduleWeekHandler struct {
	service *services.ScheduleWeekService
}

func NewScheduleWeekHandler() *ScheduleWeekHandler {
	return &ScheduleWeekHandler{service: services.NewScheduleWeekService()}
}

func (h *ScheduleWeekHandler) GetWeek(c *gin.Context) {
	startDate := c.Query("startDate")
	endDate := c.Query("endDate")
	if startDate == "" || endDate == "" {
		response.BadRequest(c, "startDate 和 endDate 必填")
		return
	}

	var wardID *int64
	if wardStr := c.Query("wardId"); wardStr != "" {
		parsed, err := strconv.ParseInt(wardStr, 10, 64)
		if err == nil {
			wardID = &parsed
		}
	}

	tenantID := middleware.GetTenantID(c)
	result, err := h.service.GetWeek(startDate, endDate, tenantID, wardID)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, result)
}

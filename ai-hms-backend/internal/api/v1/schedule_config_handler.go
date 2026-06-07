package v1

import (
	"github.com/elliotxin/ai-hms-backend/internal/middleware"
	"github.com/elliotxin/ai-hms-backend/internal/services"
	"github.com/elliotxin/ai-hms-backend/pkg/response"
	"github.com/gin-gonic/gin"
)

// ScheduleConfigHandler 排班配置控制器
type ScheduleConfigHandler struct {
	svc *services.ScheduleConfigService
}

func NewScheduleConfigHandler() *ScheduleConfigHandler {
	return &ScheduleConfigHandler{svc: services.NewScheduleConfigService()}
}

// ===================== WardExt 病区扩展 =====================

func (h *ScheduleConfigHandler) ListWardExts(c *gin.Context) {
	items, err := h.svc.ListWardExts(middleware.GetTenantID(c))
	if err != nil {
		response.InternalErrorSafe(c)
		return
	}
	response.Success(c, items)
}

func (h *ScheduleConfigHandler) UpsertWardExt(c *gin.Context) {
	var req services.WardExtRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数无效: "+err.Error())
		return
	}
	ext, err := h.svc.UpsertWardExt(middleware.GetTenantID(c), middleware.GetCreatorID(c), req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.Success(c, ext)
}

// ===================== BedMachineExt 机器扩展 =====================

func (h *ScheduleConfigHandler) ListBedMachineExts(c *gin.Context) {
	items, err := h.svc.ListBedMachineExts(middleware.GetTenantID(c))
	if err != nil {
		response.InternalErrorSafe(c)
		return
	}
	response.Success(c, items)
}

func (h *ScheduleConfigHandler) UpsertBedMachineExt(c *gin.Context) {
	var req services.BedMachineExtRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数无效: "+err.Error())
		return
	}
	ext, err := h.svc.UpsertBedMachineExt(middleware.GetTenantID(c), middleware.GetCreatorID(c), req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.Success(c, ext)
}

// ===================== PatientProfile 患者骨架 =====================

func (h *ScheduleConfigHandler) ListPatientProfiles(c *gin.Context) {
	items, err := h.svc.ListPatientProfiles(middleware.GetTenantID(c))
	if err != nil {
		response.InternalErrorSafe(c)
		return
	}
	response.Success(c, items)
}

func (h *ScheduleConfigHandler) UpsertPatientProfile(c *gin.Context) {
	var req services.PatientProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数无效: "+err.Error())
		return
	}
	profile, err := h.svc.UpsertPatientProfile(middleware.GetTenantID(c), middleware.GetCreatorID(c), req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.Success(c, profile)
}

// ===================== TenantSetting 排班配置 =====================

func (h *ScheduleConfigHandler) ListTenantSettings(c *gin.Context) {
	items, err := h.svc.ListTenantSettings(middleware.GetTenantID(c))
	if err != nil {
		response.InternalErrorSafe(c)
		return
	}
	response.Success(c, items)
}

func (h *ScheduleConfigHandler) UpsertTenantSetting(c *gin.Context) {
	var req services.TenantSettingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数无效: "+err.Error())
		return
	}
	setting, err := h.svc.UpsertTenantSetting(middleware.GetTenantID(c), middleware.GetCreatorID(c), req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.Success(c, setting)
}

// ===================== Calendar 日历 =====================

func (h *ScheduleConfigHandler) ListCalendars(c *gin.Context) {
	items, err := h.svc.ListCalendars(middleware.GetTenantID(c))
	if err != nil {
		response.InternalErrorSafe(c)
		return
	}
	response.Success(c, items)
}

func (h *ScheduleConfigHandler) UpsertCalendar(c *gin.Context) {
	var req services.CalendarRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数无效: "+err.Error())
		return
	}
	if req.IsDialysisDay == nil {
		response.BadRequest(c, "isDialysisDay 必填")
		return
	}
	cal, openWards, openBeds, err := h.svc.UpsertCalendar(middleware.GetTenantID(c), middleware.GetCreatorID(c), req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.Success(c, gin.H{
		"calendar":  cal,
		"openWards": openWards,
		"openBeds":  openBeds,
	})
}

// ===================== 路由注册 =====================

// RegisterScheduleConfigRoutes 注册排班配置路由（写操作需要管理员角色）
func RegisterScheduleConfigRoutes(r *gin.RouterGroup) {
	h := NewScheduleConfigHandler()

	// 读接口：已登录可访问
	config := r.Group("/schedule/config")
	{
		config.GET("/wards", h.ListWardExts)
		config.GET("/machines", h.ListBedMachineExts)
		config.GET("/patient-profiles", h.ListPatientProfiles)
		config.GET("/settings", h.ListTenantSettings)
		config.GET("/calendar", h.ListCalendars)
	}

	// 写接口：需要管理员角色
	adminConfig := r.Group("/schedule/config")
	adminConfig.Use(middleware.RequireRoles("ADMIN", "管理员", "安全管理员", "运维管理员"))
	{
		adminConfig.PUT("/wards", h.UpsertWardExt)
		adminConfig.PUT("/machines", h.UpsertBedMachineExt)
		adminConfig.PUT("/patient-profiles", h.UpsertPatientProfile)
		adminConfig.PUT("/settings", h.UpsertTenantSetting)
		adminConfig.PUT("/calendar", h.UpsertCalendar)
	}
}

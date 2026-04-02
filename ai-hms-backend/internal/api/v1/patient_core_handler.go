package v1

import (
	"github.com/elliotxin/ai-hms-backend/internal/services"
	"github.com/elliotxin/ai-hms-backend/pkg/response"
	"github.com/gin-gonic/gin"
)

// PatientCoreHandler 患者核心信息控制器
type PatientCoreHandler struct {
	service *services.PatientCoreService
}

// NewPatientCoreHandler 创建患者核心信息控制器
func NewPatientCoreHandler() *PatientCoreHandler {
	return &PatientCoreHandler{
		service: services.NewPatientCoreService(),
	}
}

// GetCore 获取患者核心信息聚合数据
// @Summary 获取患者核心信息（首屏聚合接口）
// @Description 返回患者详情页首屏所需的所有核心数据，包括页面头部、Overview Tab、临床焦点面板
// @Tags 患者管理
// @Accept json
// @Produce json
// @Param id path string true "患者ID"
// @Success 200 {object} response.SuccessResponse
// @Failure 404 {object} response.ErrorResponse "患者不存在"
// @Failure 500 {object} response.ErrorResponse "服务器错误"
// @Router /api/v1/patients/{id}/core [get]
func (h *PatientCoreHandler) GetCore(c *gin.Context) {
	patientID := c.Param("id")
	if patientID == "" {
		response.BadRequest(c, "患者ID不能为空")
		return
	}

	result, err := h.service.GetCore(patientID)
	if err != nil {
		if err.Error() == "patient not found" {
			response.NotFound(c, "患者不存在")
			return
		}
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, result)
}

// RegisterPatientCoreRoutes 注册患者核心信息路由
// 注意：此函数应该在 RegisterPatientRoutes 之前调用，或者将路由合并到 RegisterPatientRoutes 中
func RegisterPatientCoreRoutes(patients *gin.RouterGroup) {
	handler := NewPatientCoreHandler()

	// 注意：/:id/core 必须在 /:id 之前注册
	patients.GET("/:id/core", handler.GetCore)
}

// RegisterPatientRoutesWithCore 注册患者路由（包含 core 路由和 basic-info 路由）
func RegisterPatientRoutesWithCore(r *gin.RouterGroup) {
	handler := NewPatientHandler()
	coreHandler := NewPatientCoreHandler()
	hospitalizationHandler := NewHospitalizationHandler()
	basicInfoHandler := NewPatientBasicInfoHandler()
	vascularAccessHandler := NewVascularAccessHandler()
	orderHandler := NewOrderHandler()
	prescriptionHandler := NewPrescriptionHandler()

	patients := r.Group("/patients")
	{
		patients.GET("", handler.List)           // 获取患者列表
		patients.POST("", handler.Create)        // 创建患者
		// 注意：带子路径的路由必须在 :id 之前注册，避免 Gin 路由冲突
		patients.GET("/:id/core", coreHandler.GetCore)            // 患者核心信息（必须在 /:id 之前）
		patients.GET("/:id/basic-info", basicInfoHandler.GetBasicInfo)    // 患者基本信息档案
		patients.PUT("/:id/basic-info", basicInfoHandler.UpdateBasicInfo)  // 更新患者基本信息档案
		patients.GET("/:id/hospitalization", hospitalizationHandler.GetByPatient) // 患者的当前住院信息
		patients.GET("/:id/treatment-plans", handler.GetTreatmentPlans)   // 获取患者所有治疗方案（必须在 treatment-plan 之前）
		patients.GET("/:id/treatment-plan", handler.GetTreatmentPlan)     // 获取患者治疗方案
		patients.POST("/:id/treatment-plan", handler.CreateTreatmentPlan) // 创建患者治疗方案
		patients.PUT("/:id/treatment-plan", handler.UpdateTreatmentPlan) // 更新患者治疗方案
		patients.DELETE("/:id/treatment-plan", handler.DeleteTreatmentPlan) // 删除患者治疗方案
		// 方案调整记录
		patients.GET("/:id/adjustment-records", handler.GetAdjustmentRecords)    // 获取方案调整记录
		patients.POST("/:id/adjustment-records", handler.CreateAdjustmentRecord) // 创建方案调整记录
		// 血管通路 CRUD
		patients.GET("/:id/vascular-accesses", vascularAccessHandler.List)
		patients.POST("/:id/vascular-accesses", vascularAccessHandler.Create)
		patients.PUT("/:id/vascular-accesses/:rid", vascularAccessHandler.Update)
		patients.DELETE("/:id/vascular-accesses/:rid", vascularAccessHandler.Delete)
		// 血管通路干预记录
		patients.GET("/:id/vascular-access-interventions", vascularAccessHandler.ListInterventions)
		patients.POST("/:id/vascular-access-interventions", vascularAccessHandler.CreateIntervention)
		patients.DELETE("/:id/vascular-access-interventions/:iid", vascularAccessHandler.DeleteIntervention)
		// 医嘱 CRUD（静态子路径优先于参数路由）
		patients.POST("/:id/orders/from-template", orderHandler.CreateFromTemplate)
		patients.POST("/:id/orders/group", orderHandler.Group)
		patients.POST("/:id/orders/ungroup", orderHandler.Ungroup)
		patients.GET("/:id/orders", orderHandler.List)
		patients.POST("/:id/orders", orderHandler.Create)
		patients.POST("/:id/orders/:oid/revise", orderHandler.Revise)
		patients.POST("/:id/orders/:oid/copy", orderHandler.Copy)
		patients.PUT("/:id/orders/:oid", orderHandler.Update)
		patients.POST("/:id/orders/:oid/stop", orderHandler.Stop)
		// 处方 CRUD（静态子路径优先于参数路由）
		patients.POST("/:id/prescriptions/extract", prescriptionHandler.Extract)
		patients.GET("/:id/prescriptions", prescriptionHandler.List)
		patients.POST("/:id/prescriptions", prescriptionHandler.Create)
		patients.GET("/:id/prescriptions/:pid", prescriptionHandler.Get)
		patients.PUT("/:id/prescriptions/:pid", prescriptionHandler.Update)
		patients.POST("/:id/prescriptions/:pid/execute", prescriptionHandler.Execute)
		patients.POST("/:id/prescriptions/:pid/cancel", prescriptionHandler.Cancel)
		patients.GET("/:id", handler.Get)        // 获取患者详情
		patients.PUT("/:id", handler.Update)     // 更新患者
		patients.DELETE("/:id", handler.Delete)  // 删除患者
	}
}

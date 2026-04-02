package v1

import (
	"github.com/elliotxin/ai-hms-backend/internal/services"
	"github.com/elliotxin/ai-hms-backend/pkg/response"
	"github.com/gin-gonic/gin"
)

// PatientBasicInfoHandler 患者基本信息档案控制器
type PatientBasicInfoHandler struct {
	service *services.PatientBasicService
}

// NewPatientBasicInfoHandler 创建患者基本信息档案控制器
func NewPatientBasicInfoHandler() *PatientBasicInfoHandler {
	return &PatientBasicInfoHandler{
		service: services.NewPatientBasicService(),
	}
}

// GetBasicInfo 获取患者基本信息档案
// @Summary 获取患者基本信息档案
// @Description 返回患者详细基本信息，包括身份信息、医疗登记、生命体征社会信息、联系信息
// @Tags 患者管理
// @Accept json
// @Produce json
// @Param id path string true "患者ID"
// @Success 200 {object} response.SuccessResponse
// @Failure 404 {object} response.ErrorResponse "患者不存在"
// @Failure 500 {object} response.ErrorResponse "服务器错误"
// @Router /api/v1/patients/{id}/basic-info [get]
func (h *PatientBasicInfoHandler) GetBasicInfo(c *gin.Context) {
	patientID := c.Param("id")
	if patientID == "" {
		response.BadRequest(c, "患者ID不能为空")
		return
	}

	result, err := h.service.GetBasicInfo(patientID)
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

// UpdateBasicInfo 更新患者基本信息档案
// @Summary 更新患者基本信息档案
// @Description 更新患者的详细基本信息
// @Tags 患者管理
// @Accept json
// @Produce json
// @Param id path string true "患者ID"
// @Param request body services.PatientBasicInfoRequest true "基本信息更新请求"
// @Success 200 {object} response.SuccessResponse
// @Failure 400 {object} response.ErrorResponse "无效的请求参数"
// @Failure 404 {object} response.ErrorResponse "患者不存在"
// @Failure 500 {object} response.ErrorResponse "服务器错误"
// @Router /api/v1/patients/{id}/basic-info [put]
func (h *PatientBasicInfoHandler) UpdateBasicInfo(c *gin.Context) {
	patientID := c.Param("id")
	if patientID == "" {
		response.BadRequest(c, "患者ID不能为空")
		return
	}

	var req services.PatientBasicInfoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "无效的请求参数: "+err.Error())
		return
	}

	err := h.service.UpdateBasicInfo(patientID, &req)
	if err != nil {
		if err.Error() == "patient not found" {
			response.NotFound(c, "患者不存在")
			return
		}
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, gin.H{
		"message": "基本信息更新成功",
		"id":      patientID,
	})
}

// RegisterPatientBasicInfoRoutes 注册患者基本信息档案路由
func RegisterPatientBasicInfoRoutes(patients *gin.RouterGroup) {
	handler := NewPatientBasicInfoHandler()

	// 注意：带子路径的路由必须在 :id 之前注册
	patients.GET("/:id/basic-info", handler.GetBasicInfo)
	patients.PUT("/:id/basic-info", handler.UpdateBasicInfo)
}

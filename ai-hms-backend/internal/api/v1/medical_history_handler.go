package v1

import (
	"github.com/elliotxin/ai-hms-backend/internal/services"
	"github.com/elliotxin/ai-hms-backend/pkg/response"
	"github.com/gin-gonic/gin"
)

// MedicalHistoryHandler 临床病史控制器
type MedicalHistoryHandler struct {
	service *services.MedicalHistoryService
}

// NewMedicalHistoryHandler 创建临床病史控制器
func NewMedicalHistoryHandler() *MedicalHistoryHandler {
	return &MedicalHistoryHandler{
		service: services.NewMedicalHistoryService(),
	}
}

// GetMedicalHistory 获取临床病史
func (h *MedicalHistoryHandler) GetMedicalHistory(c *gin.Context) {
	patientID := c.Param("id")

	result, err := h.service.GetMedicalHistory(patientID)
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

// SaveMedicalHistory 保存/更新临床病史
func (h *MedicalHistoryHandler) SaveMedicalHistory(c *gin.Context) {
	patientID := c.Param("id")

	var req services.MedicalHistoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "无效的请求参数")
		return
	}

	result, err := h.service.SaveMedicalHistory(patientID, &req)
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

// ListOutcomeRecords 获取转归记录列表
func (h *MedicalHistoryHandler) ListOutcomeRecords(c *gin.Context) {
	patientID := c.Param("id")

	result, err := h.service.ListOutcomeRecords(patientID)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, result)
}

// CreateOutcomeRecord 创建转归记录
func (h *MedicalHistoryHandler) CreateOutcomeRecord(c *gin.Context) {
	patientID := c.Param("id")

	var req services.OutcomeRecordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "无效的请求参数")
		return
	}

	result, err := h.service.CreateOutcomeRecord(patientID, &req)
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

// UpdateOutcomeRecord 更新转归记录
func (h *MedicalHistoryHandler) UpdateOutcomeRecord(c *gin.Context) {
	patientID := c.Param("id")
	recordID := c.Param("rid")

	var req services.OutcomeRecordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "无效的请求参数")
		return
	}

	result, err := h.service.UpdateOutcomeRecord(patientID, recordID, &req)
	if err != nil {
		if err.Error() == "outcome record not found" {
			response.NotFound(c, "转归记录不存在")
			return
		}
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, result)
}

// DeleteOutcomeRecord 删除转归记录
func (h *MedicalHistoryHandler) DeleteOutcomeRecord(c *gin.Context) {
	patientID := c.Param("id")
	recordID := c.Param("rid")

	if err := h.service.DeleteOutcomeRecord(patientID, recordID); err != nil {
		if err.Error() == "outcome record not found" {
			response.NotFound(c, "转归记录不存在")
			return
		}
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, gin.H{"message": "删除成功"})
}

// RegisterMedicalHistoryRoutes 注册临床病史路由
func RegisterMedicalHistoryRoutes(r *gin.RouterGroup) {
	handler := NewMedicalHistoryHandler()

	// 临床病史（挂载在 /patients/:id 下）
	r.GET("/patients/:id/medical-history", handler.GetMedicalHistory)
	r.PUT("/patients/:id/medical-history", handler.SaveMedicalHistory)

	// 治疗转归记录
	r.GET("/patients/:id/outcome-records", handler.ListOutcomeRecords)
	r.POST("/patients/:id/outcome-records", handler.CreateOutcomeRecord)
	r.PUT("/patients/:id/outcome-records/:rid", handler.UpdateOutcomeRecord)
	r.DELETE("/patients/:id/outcome-records/:rid", handler.DeleteOutcomeRecord)
}

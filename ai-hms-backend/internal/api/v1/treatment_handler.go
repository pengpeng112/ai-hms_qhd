package v1

import (
	"strconv"
	"time"

	"github.com/elliotxin/ai-hms-backend/internal/middleware"
	modeltypes "github.com/elliotxin/ai-hms-backend/internal/models/types"
	"github.com/elliotxin/ai-hms-backend/internal/services"
	"github.com/elliotxin/ai-hms-backend/pkg/response"
	"github.com/gin-gonic/gin"
)

type TreatmentHandler struct {
	service *services.TreatmentService
}

func NewTreatmentHandler() *TreatmentHandler {
	return &TreatmentHandler{service: services.NewTreatmentService()}
}

func (h *TreatmentHandler) List(c *gin.Context) {
	var req services.TreatmentListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "invalid request parameters")
		return
	}

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

func (h *TreatmentHandler) Get(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid treatment id")
		return
	}

	treatment, err := h.service.Get(id)
	if err != nil {
		if err.Error() == "treatment not found" {
			response.NotFound(c, "treatment not found")
			return
		}
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, treatment)
}

func (h *TreatmentHandler) Create(c *gin.Context) {
	var req services.TreatmentCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request parameters")
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

func (h *TreatmentHandler) Update(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid treatment id")
		return
	}

	var req services.TreatmentUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request parameters")
		return
	}

	treatment, err := h.service.Update(id, req)
	if err != nil {
		if err.Error() == "treatment not found" {
			response.NotFound(c, "treatment not found")
			return
		}
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, treatment)
}

func (h *TreatmentHandler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid treatment id")
		return
	}

	if err := h.service.Delete(id); err != nil {
		if err.Error() == "treatment not found" {
			response.NotFound(c, "treatment not found")
			return
		}
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, gin.H{"message": "deleted"})
}

func (h *TreatmentHandler) UpdateStatus(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid treatment id")
		return
	}

	var req struct {
		Status int `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request parameters")
		return
	}

	if err := h.service.UpdateStatus(id, req.Status); err != nil {
		if err.Error() == "treatment not found" {
			response.NotFound(c, "treatment not found")
			return
		}
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, gin.H{"message": "updated"})
}

func (h *TreatmentHandler) GetByPatientAndDate(c *gin.Context) {
	patientIDStr := c.Param("id")
	patientIDInt, err := strconv.ParseInt(patientIDStr, 10, 64)
	if err != nil || patientIDInt <= 0 {
		response.BadRequest(c, "invalid patient id")
		return
	}

	dateStr := c.Query("date")
	if dateStr == "" {
		response.BadRequest(c, "date is required")
		return
	}

	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		response.BadRequest(c, "invalid date format, expected YYYY-MM-DD")
		return
	}

	treatment, err := h.service.GetByPatientAndDate(modeltypes.LegacyID(patientIDInt), date)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	if treatment == nil {
		response.NotFound(c, "暂无治疗记录")
		return
	}

	response.Success(c, treatment)
}

func (h *TreatmentHandler) CreateDuringParam(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid treatment id")
		return
	}

	var req services.TreatmentDuringParamRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request parameters")
		return
	}

	item, err := h.service.CreateDuringParam(id, req, middleware.GetCreatorID(c))
	if err != nil {
		if err.Error() == "treatment not found" {
			response.NotFound(c, "treatment not found")
			return
		}
		response.InternalError(c, err.Error())
		return
	}

	response.SuccessCreated(c, item)
}

func (h *TreatmentHandler) UpdateDuringParam(c *gin.Context) {
	treatmentID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid treatment id")
		return
	}
	paramID, err := strconv.ParseInt(c.Param("paramId"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid during param id")
		return
	}

	var req services.TreatmentDuringParamRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request parameters")
		return
	}

	item, err := h.service.UpdateDuringParam(treatmentID, paramID, req)
	if err != nil {
		if err.Error() == "during param not found" {
			response.NotFound(c, "during param not found")
			return
		}
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, item)
}

func (h *TreatmentHandler) DeleteDuringParam(c *gin.Context) {
	treatmentID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid treatment id")
		return
	}
	paramID, err := strconv.ParseInt(c.Param("paramId"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid during param id")
		return
	}

	if err := h.service.DeleteDuringParam(treatmentID, paramID); err != nil {
		if err.Error() == "during param not found" {
			response.NotFound(c, "during param not found")
			return
		}
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, gin.H{"message": "deleted"})
}

func (h *TreatmentHandler) SaveBeforeSigns(c *gin.Context) {
	treatmentID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid treatment id")
		return
	}
	var req services.TreatmentBeforeSignsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request parameters")
		return
	}
	item, err := h.service.SaveBeforeSigns(treatmentID, req, middleware.GetCreatorID(c))
	if err != nil {
		if err.Error() == "treatment not found" {
			response.NotFound(c, "treatment not found")
			return
		}
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, item)
}

func (h *TreatmentHandler) SaveAfterSigns(c *gin.Context) {
	treatmentID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid treatment id")
		return
	}
	var req services.TreatmentAfterSignsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request parameters")
		return
	}
	item, err := h.service.SaveAfterSigns(treatmentID, req, middleware.GetCreatorID(c))
	if err != nil {
		if err.Error() == "treatment not found" {
			response.NotFound(c, "treatment not found")
			return
		}
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, item)
}

func (h *TreatmentHandler) SaveFirstCheck(c *gin.Context) {
	treatmentID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid treatment id")
		return
	}
	var req services.TreatmentFirstCheckRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request parameters")
		return
	}
	item, err := h.service.SaveFirstCheck(treatmentID, req, middleware.GetCreatorID(c))
	if err != nil {
		if err.Error() == "treatment not found" {
			response.NotFound(c, "treatment not found")
			return
		}
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, item)
}

func (h *TreatmentHandler) SaveSecondCheck(c *gin.Context) {
	treatmentID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid treatment id")
		return
	}
	var req services.TreatmentSecondCheckRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request parameters")
		return
	}
	item, err := h.service.SaveSecondCheck(treatmentID, req, middleware.GetCreatorID(c))
	if err != nil {
		if err.Error() == "treatment not found" {
			response.NotFound(c, "treatment not found")
			return
		}
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, item)
}

func RegisterTreatmentRoutes(r *gin.RouterGroup) {
	handler := NewTreatmentHandler()

	treatments := r.Group("/treatments")
	{
		treatments.GET("", handler.List)
		treatments.POST("", handler.Create)
		treatments.GET("/:id", handler.Get)
		treatments.PUT("/:id", handler.Update)
		treatments.DELETE("/:id", handler.Delete)
		treatments.PUT("/:id/status", handler.UpdateStatus)
		treatments.POST("/:id/during-params", handler.CreateDuringParam)
		treatments.PUT("/:id/during-params/:paramId", handler.UpdateDuringParam)
		treatments.DELETE("/:id/during-params/:paramId", handler.DeleteDuringParam)
		treatments.PUT("/:id/before-signs", handler.SaveBeforeSigns)
		treatments.PUT("/:id/after-signs", handler.SaveAfterSigns)
		treatments.PUT("/:id/first-check", handler.SaveFirstCheck)
		treatments.PUT("/:id/second-check", handler.SaveSecondCheck)
	}

	r.GET("/patients/:id/treatment", handler.GetByPatientAndDate)
}

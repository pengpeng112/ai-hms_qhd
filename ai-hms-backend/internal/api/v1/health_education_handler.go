package v1

import (
	"github.com/elliotxin/ai-hms-backend/internal/services"
	"github.com/elliotxin/ai-hms-backend/pkg/response"
	"github.com/gin-gonic/gin"
)

type HealthEducationHandler struct {
	service *services.HealthEducationService
}

func NewHealthEducationHandler() *HealthEducationHandler {
	return &HealthEducationHandler{service: services.NewHealthEducationService()}
}

func (h *HealthEducationHandler) ListContents(c *gin.Context) {
	contents, err := h.service.ListContents()
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, contents)
}

func (h *HealthEducationHandler) ListPatientEducations(c *gin.Context) {
	patientID := c.Param("id")
	if patientID == "" {
		response.BadRequest(c, "patient id is required")
		return
	}
	records, err := h.service.ListPatientEducations(patientID)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, records)
}

func (h *HealthEducationHandler) CreatePatientEducation(c *gin.Context) {
	patientID := c.Param("id")
	if patientID == "" {
		response.BadRequest(c, "patient id is required")
		return
	}
	var req services.CreatePatientHealthEducationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request: "+err.Error())
		return
	}
	record, err := h.service.CreatePatientEducation(patientID, req)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.SuccessCreated(c, record)
}

func RegisterHealthEducationRoutes(rg *gin.RouterGroup) {
	h := NewHealthEducationHandler()
	rg.GET("/health-educations", h.ListContents)
	rg.POST("/health-educations", h.CreateContent)
	rg.PUT("/health-educations/:id", h.UpdateContent)
	rg.DELETE("/health-educations/:id", h.DeleteContent)
	patients := rg.Group("/patients")
	{
		patients.GET("/:id/health-educations", h.ListPatientEducations)
		patients.POST("/:id/health-educations", h.CreatePatientEducation)
	}
}

func (h *HealthEducationHandler) CreateContent(c *gin.Context) {
	var req services.CreateHealthEducationContentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request: name is required")
		return
	}
	item, err := h.service.CreateContent(req)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.SuccessCreated(c, item)
}

func (h *HealthEducationHandler) UpdateContent(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.BadRequest(c, "id is required")
		return
	}
	var req services.UpdateHealthEducationContentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request")
		return
	}
	item, err := h.service.UpdateContent(id, req)
	if err != nil {
		if err.Error() == "content not found" || err.Error() == "content not found after update" {
			response.NotFound(c, "宣教内容不存在")
			return
		}
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, item)
}

func (h *HealthEducationHandler) DeleteContent(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.BadRequest(c, "id is required")
		return
	}
	if err := h.service.DeleteContent(id); err != nil {
		if err.Error() == "content not found" {
			response.NotFound(c, "宣教内容不存在")
			return
		}
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, gin.H{"id": id})
}
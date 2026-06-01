package v1

import (
	"strconv"

	"github.com/elliotxin/ai-hms-backend/internal/middleware"
	"github.com/elliotxin/ai-hms-backend/internal/services"
	"github.com/elliotxin/ai-hms-backend/pkg/response"
	"github.com/gin-gonic/gin"
)

func RegisterMonthlySummaryRoutes(r *gin.RouterGroup) {
	h := &MonthlySummaryHandler{service: services.NewMonthlySummaryService()}
	r.GET("/patients/:id/monthly-summaries", h.Get)
	r.PUT("/patients/:id/monthly-summaries", h.Save)
}

type MonthlySummaryHandler struct {
	service *services.MonthlySummaryService
}

func (h *MonthlySummaryHandler) Get(c *gin.Context) {
	patientID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid patient id")
		return
	}
	year, err := strconv.Atoi(c.DefaultQuery("year", "0"))
	if err != nil || year <= 0 {
		response.BadRequest(c, "year is required")
		return
	}
	month, err := strconv.Atoi(c.DefaultQuery("month", "0"))
	if err != nil || month < 1 || month > 12 {
		response.BadRequest(c, "month is required (1-12)")
		return
	}

	tenantID := middleware.GetTenantID(c)
	if tenantID <= 0 {
		response.Unauthorized(c, "tenant id missing")
		return
	}

	data, err := h.service.Get(patientID, tenantID, year, month)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	if data == nil {
		response.Success(c, gin.H{"id": 0, "patientId": patientID, "year": year, "month": month, "content": gin.H{}})
		return
	}
	response.Success(c, data)
}

func (h *MonthlySummaryHandler) Save(c *gin.Context) {
	patientID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid patient id")
		return
	}
	year, err := strconv.Atoi(c.Query("year"))
	if err != nil || year <= 0 {
		response.BadRequest(c, "year is required")
		return
	}
	month, err := strconv.Atoi(c.Query("month"))
	if err != nil || month < 1 || month > 12 {
		response.BadRequest(c, "month is required (1-12)")
		return
	}

	var req services.SaveMonthlySummaryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request: "+err.Error())
		return
	}

	tenantID := middleware.GetTenantID(c)
	if tenantID <= 0 {
		response.Unauthorized(c, "tenant id missing")
		return
	}

	userID, exists := c.Get("user_id")
	creatorID := int64(0)
	if exists {
		if uid, ok := userID.(int64); ok {
			creatorID = uid
		}
	}

	data, err := h.service.Save(patientID, tenantID, year, month, req, creatorID)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, data)
}

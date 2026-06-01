package v1

import (
	"strconv"

	"github.com/elliotxin/ai-hms-backend/internal/middleware"
	"github.com/elliotxin/ai-hms-backend/internal/services"
	"github.com/elliotxin/ai-hms-backend/pkg/response"
	"github.com/gin-gonic/gin"
)

func RegisterConsumableRoutes(r *gin.RouterGroup) {
	h := &ConsumableHandler{service: services.NewConsumableService()}
	r.GET("/treatments/:id/consumables", h.List)
	r.POST("/treatments/:id/consumables", h.Create)
	r.DELETE("/treatments/:id/consumables/:cid", h.Delete)
}

type ConsumableHandler struct {
	service *services.ConsumableService
}

func (h *ConsumableHandler) List(c *gin.Context) {
	treatmentID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid treatment id")
		return
	}
	tenantID := middleware.GetTenantID(c)
	if tenantID <= 0 {
		response.Unauthorized(c, "tenant id missing")
		return
	}
	data, err := h.service.ListByTreatment(treatmentID, tenantID)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, data)
}

func (h *ConsumableHandler) Create(c *gin.Context) {
	treatmentID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid treatment id")
		return
	}
	var req services.CreateConsumableRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request: "+err.Error())
		return
	}
	tenantID := middleware.GetTenantID(c)
	if tenantID <= 0 {
		response.Unauthorized(c, "tenant id missing")
		return
	}
	userID, _ := c.Get("user_id")
	creatorID := int64(0)
	if uid, ok := userID.(int64); ok {
		creatorID = uid
	}
	data, err := h.service.Create(treatmentID, tenantID, creatorID, req)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.SuccessCreated(c, data)
}

func (h *ConsumableHandler) Delete(c *gin.Context) {
	cid, err := strconv.ParseInt(c.Param("cid"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid consumable id")
		return
	}
	tenantID := middleware.GetTenantID(c)
	if tenantID <= 0 {
		response.Unauthorized(c, "tenant id missing")
		return
	}
	if err := h.service.Delete(cid, tenantID); err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, gin.H{"id": cid})
}

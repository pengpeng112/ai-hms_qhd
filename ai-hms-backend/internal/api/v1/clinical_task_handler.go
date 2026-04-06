package v1

import (
	"strconv"

	"github.com/elliotxin/ai-hms-backend/internal/middleware"
	"github.com/elliotxin/ai-hms-backend/internal/services"
	"github.com/elliotxin/ai-hms-backend/pkg/response"
	"github.com/gin-gonic/gin"
)

type ClinicalTaskHandler struct {
	service *services.ClinicalTaskService
}

func NewClinicalTaskHandler() *ClinicalTaskHandler {
	return &ClinicalTaskHandler{service: services.NewClinicalTaskService()}
}

func (h *ClinicalTaskHandler) List(c *gin.Context) {
	status := c.Query("status")
	tenantId := middleware.GetTenantID(c)
	items, total, err := h.service.List(status, tenantId)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, gin.H{
		"items": items,
		"total": total,
	})
}

func (h *ClinicalTaskHandler) UpdateStatus(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}
	var req struct {
		Status string `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request")
		return
	}
	if req.Status != "handled" && req.Status != "dismissed" && req.Status != "pending" {
		response.BadRequest(c, "invalid status")
		return
	}
	if err := h.service.UpdateStatus(id, req.Status, middleware.GetCreatorID(c), middleware.GetTenantID(c)); err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, gin.H{})
}

func RegisterClinicalTaskRoutes(rg *gin.RouterGroup) {
	h := NewClinicalTaskHandler()
	group := rg.Group("/clinical-tasks")
	{
		group.GET("", h.List)
		group.PUT("/:id/status", h.UpdateStatus)
	}
}

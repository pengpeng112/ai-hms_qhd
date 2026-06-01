package v1

import (
	"strconv"

	"github.com/elliotxin/ai-hms-backend/internal/services"
	"github.com/elliotxin/ai-hms-backend/pkg/response"
	"github.com/gin-gonic/gin"
)

type BedHandler struct {
	service *services.BedService
}

func NewBedHandler() *BedHandler {
	return &BedHandler{service: services.NewBedService()}
}

func (h *BedHandler) List(c *gin.Context) {
	includeDisabled := c.DefaultQuery("includeDisabled", "true") == "true"
	items, err := h.service.List(includeDisabled)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, gin.H{
		"items":    items,
		"total":    len(items),
		"page":     1,
		"pageSize": len(items),
	})
}

func (h *BedHandler) Get(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid bed id")
		return
	}
	item, err := h.service.GetByID(id)
	if err != nil {
		if err.Error() == "bed not found" {
			response.NotFound(c, "床位不存在")
			return
		}
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, item)
}

func (h *BedHandler) Create(c *gin.Context) {
	var req services.BedCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request: name and wardId are required")
		return
	}
	item, err := h.service.Create(req)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.SuccessCreated(c, item)
}

func (h *BedHandler) Update(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid bed id")
		return
	}
	var req services.BedUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request")
		return
	}
	item, err := h.service.Update(id, req)
	if err != nil {
		if err.Error() == "bed not found" {
			response.NotFound(c, "床位不存在")
			return
		}
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, item)
}

func (h *BedHandler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid bed id")
		return
	}
	if err := h.service.Delete(id); err != nil {
		if err.Error() == "bed not found" {
			response.NotFound(c, "床位不存在")
			return
		}
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, gin.H{"id": idStr})
}

func RegisterBedRoutes(rg *gin.RouterGroup) {
	h := NewBedHandler()
	rg.GET("/beds", h.List)
	rg.POST("/beds", h.Create)
	rg.GET("/beds/:id", h.Get)
	rg.PUT("/beds/:id", h.Update)
	rg.DELETE("/beds/:id", h.Delete)
}

package v1

import (
	"strconv"

	"github.com/elliotxin/ai-hms-backend/internal/services"
	"github.com/elliotxin/ai-hms-backend/pkg/response"
	"github.com/gin-gonic/gin"
)

type WardHandler struct {
	service *services.WardService
}

func NewWardHandler() *WardHandler {
	return &WardHandler{service: services.NewWardService()}
}

func (h *WardHandler) List(c *gin.Context) {
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

func (h *WardHandler) Get(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid ward id")
		return
	}
	item, err := h.service.GetByID(id)
	if err != nil {
		if err.Error() == "ward not found" {
			response.NotFound(c, "病区不存在")
			return
		}
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, item)
}

func (h *WardHandler) Create(c *gin.Context) {
	var req services.WardCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request: name is required")
		return
	}
	item, err := h.service.Create(req)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.SuccessCreated(c, item)
}

func (h *WardHandler) Update(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid ward id")
		return
	}
	var req services.WardUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request")
		return
	}
	item, err := h.service.Update(id, req)
	if err != nil {
		if err.Error() == "ward not found" {
			response.NotFound(c, "病区不存在")
			return
		}
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, item)
}

func (h *WardHandler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid ward id")
		return
	}
	if err := h.service.Delete(id); err != nil {
		if err.Error() == "ward not found" {
			response.NotFound(c, "病区不存在")
			return
		}
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, gin.H{"id": idStr})
}

func RegisterWardRoutes(rg *gin.RouterGroup) {
	h := NewWardHandler()
	rg.GET("/wards", h.List)
	rg.POST("/wards", h.Create)
	rg.GET("/wards/:id", h.Get)
	rg.PUT("/wards/:id", h.Update)
	rg.DELETE("/wards/:id", h.Delete)
}

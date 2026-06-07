package v1

import (
	"fmt"

	"github.com/elliotxin/ai-hms-backend/internal/services"
	"github.com/elliotxin/ai-hms-backend/pkg/response"
	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	service *services.UserService
}

func NewUserHandler() *UserHandler {
	return &UserHandler{service: services.NewUserService()}
}

func (h *UserHandler) List(c *gin.Context) {
	var req services.UserListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "无效的请求参数")
		return
	}

	dtos, err := h.service.List(req)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, gin.H{"items": dtos, "total": len(dtos)})
}

func (h *UserHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.BadRequest(c, "user id is required")
		return
	}
	dto, err := h.service.GetByID(id)
	if err != nil {
		if err.Error() == "user not found" {
			response.NotFound(c, "用户不存在")
			return
		}
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, dto)
}

func (h *UserHandler) Create(c *gin.Context) {
	var req services.UserCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "无效的请求参数: "+err.Error())
		return
	}
	dto, err := h.service.Create(req)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.SuccessCreated(c, dto)
}

func (h *UserHandler) Update(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.BadRequest(c, "user id is required")
		return
	}
	var req services.UserUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "无效的请求参数: "+err.Error())
		return
	}
	dto, err := h.service.Update(id, req)
	if err != nil {
		if err.Error() == "user not found" {
			response.NotFound(c, "用户不存在")
			return
		}
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, dto)
}

func (h *UserHandler) UpdateStatus(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.BadRequest(c, "user id is required")
		return
	}
	var req services.UserUpdateStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "无效的请求参数")
		return
	}
	dto, err := h.service.UpdateStatus(id, req)
	if err != nil {
		if err.Error() == "user not found" {
			response.NotFound(c, "用户不存在")
			return
		}
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, dto)
}

func (h *UserHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.BadRequest(c, "user id is required")
		return
	}
	if err := h.service.Delete(id); err != nil {
		if err.Error() == "user not found" {
			response.NotFound(c, "用户不存在")
			return
		}
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, gin.H{"id": id})
}

func (h *UserHandler) ResetPassword(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.BadRequest(c, "user id is required")
		return
	}
	var req struct {
		NewPassword string `json:"newPassword" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "newPassword is required")
		return
	}
	if err := h.service.ResetPassword(id, req.NewPassword); err != nil {
		if err.Error() == "user not found" {
			response.NotFound(c, "用户不存在")
			return
		}
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, gin.H{"id": id})
}

func (h *UserHandler) GetRoles(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.BadRequest(c, "user id is required")
		return
	}
	roles, err := h.service.GetUserRoles(id)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, gin.H{"userId": id, "roles": roles})
}

func (h *UserHandler) SetRoles(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.BadRequest(c, "user id is required")
		return
	}
	var req struct {
		RoleIDs []string `json:"roleIds"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "无效的请求参数")
		return
	}
	if err := h.service.SetUserRoles(id, req.RoleIDs); err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, gin.H{"userId": id, "roleIds": req.RoleIDs})
}

func (h *UserHandler) GetMyRoles(c *gin.Context) {
	userIDStr, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "未登录")
		return
	}
	roles, err := h.service.GetUserRoles(fmt.Sprintf("%v", userIDStr))
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, gin.H{"userId": userIDStr, "roles": roles})
}

func RegisterUserRoutes(rg *gin.RouterGroup) {
	h := NewUserHandler()
	users := rg.Group("/users")
	{
		users.GET("", h.List)
		users.POST("", h.Create)
		users.GET("/:id", h.GetByID)
		users.PUT("/:id", h.Update)
		users.DELETE("/:id", h.Delete)
		users.PUT("/:id/status", h.UpdateStatus)
		users.PUT("/:id/password", h.ResetPassword)
		users.GET("/:id/roles", h.GetRoles)
		users.PUT("/:id/roles", h.SetRoles)
	}
	rg.GET("/me/roles", h.GetMyRoles)
}
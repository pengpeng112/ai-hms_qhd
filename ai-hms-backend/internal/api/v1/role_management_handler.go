package v1

import (
	"github.com/elliotxin/ai-hms-backend/internal/services"
	"github.com/elliotxin/ai-hms-backend/pkg/response"
	"github.com/gin-gonic/gin"
)

type RoleManagementHandler struct {
	service *services.PermissionService
}

func NewRoleManagementHandler() *RoleManagementHandler {
	return &RoleManagementHandler{service: services.NewPermissionService()}
}

func (h *RoleManagementHandler) ListRoles(c *gin.Context) {
	roles, err := h.service.ListRoles()
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, roles)
}

func (h *RoleManagementHandler) CreateRole(c *gin.Context) {
	var req struct {
		Name string `json:"name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "name is required")
		return
	}
	role, err := h.service.CreateRole(req.Name)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.SuccessCreated(c, role)
}

func (h *RoleManagementHandler) UpdateRole(c *gin.Context) {
	code := c.Param("code")
	if code == "" {
		response.BadRequest(c, "role code is required")
		return
	}
	var req struct {
		Name string `json:"name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "name is required")
		return
	}
	role, err := h.service.UpdateRole(code, req.Name)
	if err != nil {
		if err.Error() == "role not found" {
			response.NotFound(c, "角色不存在")
			return
		}
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, role)
}

func (h *RoleManagementHandler) DeleteRole(c *gin.Context) {
	code := c.Param("code")
	if code == "" {
		response.BadRequest(c, "role code is required")
		return
	}
	if err := h.service.DeleteRole(code); err != nil {
		if err.Error() == "role not found" {
			response.NotFound(c, "角色不存在")
			return
		}
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, gin.H{"code": code})
}

func (h *RoleManagementHandler) GetPermissionTree(c *gin.Context) {
	tree, err := h.service.GetPermissionTree()
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, tree)
}

func RegisterRoleManagementRoutes(rg *gin.RouterGroup) {
	h := NewRoleManagementHandler()
	roles := rg.Group("/app-roles")
	{
		roles.GET("", h.ListRoles)
		roles.POST("", h.CreateRole)
		roles.PUT("/:code", h.UpdateRole)
		roles.DELETE("/:code", h.DeleteRole)
	}
	rg.GET("/app-permissions/tree", h.GetPermissionTree)
}
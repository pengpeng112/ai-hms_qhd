package v1

import (
	"strings"

	"github.com/elliotxin/ai-hms-backend/internal/models"
	"github.com/elliotxin/ai-hms-backend/internal/services"
	"github.com/elliotxin/ai-hms-backend/pkg/response"
	"github.com/gin-gonic/gin"
)

type PermissionHandler struct {
	service *services.PermissionService
}

func NewPermissionHandler() *PermissionHandler {
	return &PermissionHandler{service: services.NewPermissionService()}
}

func (h *PermissionHandler) ListPermissions(c *gin.Context) {
	items, err := h.service.ListPermissions()
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, gin.H{"items": items, "total": len(items)})
}

func (h *PermissionHandler) SavePermission(c *gin.Context) {
	var req models.Permission
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request")
		return
	}
	req.Code = strings.TrimSpace(req.Code)
	req.Name = strings.TrimSpace(req.Name)
	if req.Code == "" || req.Name == "" {
		response.BadRequest(c, "code and name are required")
		return
	}
	item, err := h.service.SavePermission(req)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, item)
}

func (h *PermissionHandler) GetRolePermissions(c *gin.Context) {
	role := strings.TrimSpace(c.Param("role"))
	if role == "" {
		response.BadRequest(c, "role is required")
		return
	}
	items, err := h.service.GetRolePermissionCodes(role)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, gin.H{"role": role, "permissionCodes": items})
}

func (h *PermissionHandler) SetRolePermissions(c *gin.Context) {
	role := strings.TrimSpace(c.Param("role"))
	if role == "" {
		response.BadRequest(c, "role is required")
		return
	}
	var req struct {
		PermissionCodes []string `json:"permissionCodes"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request")
		return
	}
	unique := make([]string, 0, len(req.PermissionCodes))
	seen := map[string]struct{}{}
	for _, code := range req.PermissionCodes {
		code = strings.TrimSpace(code)
		if code == "" {
			continue
		}
		if _, ok := seen[code]; ok {
			continue
		}
		seen[code] = struct{}{}
		unique = append(unique, code)
	}
	if err := h.service.SetRolePermissionCodes(role, unique); err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, gin.H{"role": role, "permissionCodes": unique})
}

func RegisterPermissionRoutes(rg *gin.RouterGroup) {
	h := NewPermissionHandler()
	rg.GET("/permissions", h.ListPermissions)
	rg.POST("/permissions", h.SavePermission)
	rg.GET("/role-permissions/:role", h.GetRolePermissions)
	rg.PUT("/role-permissions/:role", h.SetRolePermissions)
}

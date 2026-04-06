package v1

import (
	"github.com/elliotxin/ai-hms-backend/internal/services"
	"github.com/elliotxin/ai-hms-backend/pkg/response"
	"github.com/gin-gonic/gin"
)

// UserHandler 用户处理器
type UserHandler struct {
	service *services.UserService
}

// NewUserHandler 创建用户处理器
func NewUserHandler() *UserHandler {
	return &UserHandler{
		service: services.NewUserService(),
	}
}

// List 获取用户列表
// @Summary 获取用户列表（角色选择）
// @Param role query string false "角色过滤"
// @Param status query string false "状态过滤"
// @Success 200 {object} response.SuccessResponse
// @Router /api/v1/users [get]
func (h *UserHandler) List(c *gin.Context) {
	var req services.UserListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "无效的请求参数")
		return
	}

	users, err := h.service.List(req)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, users)
}

// RegisterUserRoutes 注册用户路由
func RegisterUserRoutes(rg *gin.RouterGroup) {
	h := NewUserHandler()
	users := rg.Group("/users")
	{
		users.GET("", h.List)
	}
}

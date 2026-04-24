package v1

import (
	"strings"

	"github.com/elliotxin/ai-hms-backend/internal/services"
	"github.com/elliotxin/ai-hms-backend/internal/utils"
	"github.com/elliotxin/ai-hms-backend/pkg/response"
	"github.com/gin-gonic/gin"
)

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse 登录响应
type LoginResponse struct {
	Token    string `json:"token"`
	UserID   string `json:"userId"`
	Username string `json:"username"`
	RealName string `json:"realName"`
	Role     string `json:"role"`
}

// AuthHandler 认证处理器
type AuthHandler struct {
	jwtManager  *utils.JWTManager
	authService *services.AuthService
}

// NewAuthHandler 创建认证处理器
func NewAuthHandler(jwtManager *utils.JWTManager) *AuthHandler {
	return &AuthHandler{
		jwtManager:  jwtManager,
		authService: services.NewAuthService(),
	}
}

// Login 登录
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "无效的请求参数")
		return
	}

	username := strings.TrimSpace(req.Username)
	user, err := h.authService.Authenticate(username, req.Password)
	if err != nil {
		response.Unauthorized(c, "用户名或密码错误")
		return
	}

	token, err := h.jwtManager.GenerateToken(
		services.FormatUserID(user.UserID),
		user.Username,
		user.EmployeeName,
		user.Roles,
		user.TenantID,
	)
	if err != nil {
		response.InternalError(c, "登录失败")
		return
	}

	response.Success(c, LoginResponse{
		Token:    token,
		UserID:   services.FormatUserID(user.UserID),
		Username: user.Username,
		RealName: user.EmployeeName,
		Role:     user.Role,
	})
}

// RegisterAuthRoutes 注册认证路由
func RegisterAuthRoutes(rg *gin.RouterGroup, jwtManager *utils.JWTManager) {
	h := NewAuthHandler(jwtManager)
	rg.POST("/auth/login", h.Login)
}

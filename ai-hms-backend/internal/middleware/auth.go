package middleware

import (
	"strconv"
	"strings"

	"github.com/elliotxin/ai-hms-backend/internal/utils"
	"github.com/elliotxin/ai-hms-backend/pkg/response"
	"github.com/gin-gonic/gin"
)

// AuthMiddleware JWT 认证中间件
func AuthMiddleware(jwtManager *utils.JWTManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从 Authorization header 获取 token
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Unauthorized(c, "缺少认证令牌")
			c.Abort()
			return
		}

		// 解析 Bearer token
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			response.Unauthorized(c, "无效的认证令牌格式")
			c.Abort()
			return
		}

		// 验证 token
		claims, err := jwtManager.ParseToken(parts[1])
		if err != nil {
			response.Unauthorized(c, "认证令牌无效或已过期")
			c.Abort()
			return
		}

		// 将用户信息存入上下文
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("roles", claims.Roles)

		c.Next()
	}
}

// OptionalAuthMiddleware 可选的认证中间件（不强制要求登录）
func OptionalAuthMiddleware(jwtManager *utils.JWTManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.Next()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) == 2 && parts[0] == "Bearer" {
			if claims, err := jwtManager.ParseToken(parts[1]); err == nil {
				c.Set("user_id", claims.UserID)
				c.Set("username", claims.Username)
				c.Set("roles", claims.Roles)
			}
		}

		c.Next()
	}
}

// RequireRoles 要求特定角色的中间件
func RequireRoles(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRoles, exists := c.Get("roles")
		if !exists {
			response.Forbidden(c, "无权限访问")
			c.Abort()
			return
		}

		userRoleList := userRoles.([]string)
		for _, requiredRole := range roles {
			for _, userRole := range userRoleList {
				if userRole == requiredRole {
					c.Next()
					return
				}
			}
		}

		response.Forbidden(c, "权限不足")
		c.Abort()
	}
}

// GetUserID 从上下文获取用户 ID
func GetUserID(c *gin.Context) string {
	userID, _ := c.Get("user_id")
	if userID == nil {
		return ""
	}
	return userID.(string)
}

// GetUsername 从上下文获取用户名
func GetUsername(c *gin.Context) string {
	username, _ := c.Get("username")
	if username == nil {
		return ""
	}
	return username.(string)
}

// GetRoles 从上下文获取用户角色
func GetRoles(c *gin.Context) []string {
	roles, _ := c.Get("roles")
	if roles == nil {
		return []string{}
	}
	return roles.([]string)
}

// GetCreatorID 从上下文获取创建者 ID（int64）
// JWT 中 user_id 为 string，若可解析为 int64 则返回，否则返回默认值 1
func GetCreatorID(c *gin.Context) int64 {
	userID := GetUserID(c)
	if userID == "" {
		return 1
	}
	id, err := strconv.ParseInt(userID, 10, 64)
	if err != nil {
		return 1
	}
	return id
}

// GetTenantID 从上下文获取租户 ID
// 当前为单租户模式，固定返回 1；后续多租户时从 JWT 或请求头提取
func GetTenantID(c *gin.Context) int64 {
	return 1
}

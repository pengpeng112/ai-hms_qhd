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
		if claims.EmployeeName != "" {
			c.Set("employee_name", claims.EmployeeName)
		}
		c.Set("roles", claims.Roles)
		if claims.TenantID <= 0 {
			response.Forbidden(c, "缺少租户信息")
			c.Abort()
			return
		}
		c.Set("tenant_id", claims.TenantID)

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
				if claims.EmployeeName != "" {
					c.Set("employee_name", claims.EmployeeName)
				}
				c.Set("roles", claims.Roles)
				if claims.TenantID > 0 {
					c.Set("tenant_id", claims.TenantID)
				}
			}
		}

		c.Next()
	}
}

// RequireRoles 要求特定角色的中间件
func RequireRoles(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if _, exists := c.Get("roles"); !exists {
			response.Forbidden(c, "无权限访问")
			c.Abort()
			return
		}
		if hasAnyRole(GetRoles(c), roles) {
			c.Next()
			return
		}
		response.Forbidden(c, "权限不足")
		c.Abort()
	}
}

// AdminRoles 系统管理类操作的兜底角色集合（单一真值源）。
// main.go 的 admin 路由组与 smart_schedule 的角色映射均引用此处，避免角色名字面量散落重复。
var AdminRoles = []string{"ADMIN", "管理员", "安全管理员", "运维管理员"}

// AdminPermissionCodes 管理类操作期望的权限码（过渡占位）。
// 需在老库 Authorization_Permissions 中配置该权限码并赋予相应角色后方才生效；
// 在此之前由 AdminRoles 兜底放行。确认落库后可细化拆分，并逐步移除角色兜底完成权限码迁移。
var AdminPermissionCodes = []string{"system:manage"}

// IsAdminRole 判断单个角色名是否属于管理员角色集合。
func IsAdminRole(role string) bool {
	for _, r := range AdminRoles {
		if r == role {
			return true
		}
	}
	return false
}

// hasAnyRole 判断 have 中是否包含 want 里的任一角色。
func hasAnyRole(have, want []string) bool {
	for _, h := range have {
		for _, w := range want {
			if h == w {
				return true
			}
		}
	}
	return false
}

// PermissionResolver 给定用户角色列表，返回其拥有的权限码集合（并集）。
type PermissionResolver func(roles []string) (map[string]bool, error)

// RequirePermissions 基于权限码的门禁中间件，带角色兜底。
//
// 放行条件（任一满足即可）：
//  1. 用户命中 fallbackRoles 中任一角色（兜底，保证权限码尚未在老库配置时管理员不被锁死）；
//  2. 用户经 resolver 解析出的权限码命中 required 中任一项。
//
// 这是从"硬编码角色名门禁"向"细粒度权限码门禁"过渡的安全形态：
// 一旦老库 Authorization_Permissions 配置了 required 中的权限码并赋予相应角色，
// 即可逐步收窄/移除 fallbackRoles 完成迁移，期间不会造成越权或锁死。
// resolver 为 nil 或 required 为空时，退化为纯 fallbackRoles 角色校验。
func RequirePermissions(resolver PermissionResolver, fallbackRoles []string, required ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if _, exists := c.Get("roles"); !exists {
			response.Forbidden(c, "无权限访问")
			c.Abort()
			return
		}
		roles := GetRoles(c)
		if hasAnyRole(roles, fallbackRoles) {
			c.Next()
			return
		}
		if resolver != nil && len(required) > 0 {
			if codes, err := resolver(roles); err == nil {
				for _, req := range required {
					if codes[req] {
						c.Next()
						return
					}
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
// 如果无法解析到有效用户 ID 则返回 0，调用方应决定如何处理（拒绝操作或使用系统用户 ID）
func GetCreatorID(c *gin.Context) int64 {
	userID := GetUserID(c)
	if userID == "" {
		return 0
	}
	id, err := strconv.ParseInt(userID, 10, 64)
	if err != nil {
		return 0
	}
	return id
}

// GetTenantID 从上下文获取租户 ID
// 从上下文中解析 int64/int/float64/string 形式的 tenant_id，缺失或非法返回 0
func GetTenantID(c *gin.Context) int64 {
	v, exists := c.Get("tenant_id")
	if !exists || v == nil {
		return 0
	}

	switch tenantID := v.(type) {
	case int64:
		if tenantID > 0 {
			return tenantID
		}
	case int:
		if tenantID > 0 {
			return int64(tenantID)
		}
	case float64:
		if tenantID > 0 {
			return int64(tenantID)
		}
	case string:
		if tenantID == "" {
			return 0
		}
		parsed, err := strconv.ParseInt(tenantID, 10, 64)
		if err == nil && parsed > 0 {
			return parsed
		}
	}

	return 0
}

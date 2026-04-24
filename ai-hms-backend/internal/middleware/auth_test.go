package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/elliotxin/ai-hms-backend/config"
	"github.com/elliotxin/ai-hms-backend/internal/utils"
	"github.com/gin-gonic/gin"
)

func TestAuthMiddlewareRejectsMissingTenantClaim(t *testing.T) {
	gin.SetMode(gin.TestMode)

	jwtManager := utils.NewJWTManager(&config.JWTConfig{Secret: "test-secret", ExpirationHours: 1})
	token, err := jwtManager.GenerateToken("1", "tester", "", []string{"ADMIN"}, 0)
	if err != nil {
		t.Fatalf("GenerateToken() error = %v", err)
	}

	r := gin.New()
	r.Use(AuthMiddleware(jwtManager))
	r.GET("/protected", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusForbidden)
	}
	if !strings.Contains(w.Body.String(), "缺少租户信息") {
		t.Fatalf("response body = %q, want missing tenant message", w.Body.String())
	}
}

func TestGetTenantIDReturnsZeroWhenMissing(t *testing.T) {
	c, _ := gin.CreateTestContext(httptest.NewRecorder())

	if got := GetTenantID(c); got != 0 {
		t.Fatalf("GetTenantID() = %d, want 0", got)
	}
}

func TestGetTenantIDParsesStringValue(t *testing.T) {
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Set("tenant_id", "123")

	if got := GetTenantID(c); got != 123 {
		t.Fatalf("GetTenantID() = %d, want 123", got)
	}
}

func TestAuthMiddlewareSetsContextOnValidToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	jwtManager := utils.NewJWTManager(&config.JWTConfig{Secret: "test-secret", ExpirationHours: 1})
	token, err := jwtManager.GenerateToken("12", "tester", "张三", []string{"NURSE"}, 3)
	if err != nil {
		t.Fatalf("GenerateToken() error = %v", err)
	}

	r := gin.New()
	r.Use(AuthMiddleware(jwtManager))
	r.GET("/protected", func(c *gin.Context) {
		if got := GetUserID(c); got != "12" {
			t.Fatalf("GetUserID() = %q, want %q", got, "12")
		}
		if got := GetUsername(c); got != "tester" {
			t.Fatalf("GetUsername() = %q, want %q", got, "tester")
		}
		roles := GetRoles(c)
		if len(roles) != 1 || roles[0] != "NURSE" {
			t.Fatalf("GetRoles() = %#v, want [NURSE]", roles)
		}
		tenantID := GetTenantID(c)
		if tenantID != 3 {
			t.Fatalf("GetTenantID() = %d, want 3", tenantID)
		}
		employeeName, exists := c.Get("employee_name")
		if !exists || employeeName != "张三" {
			t.Fatalf("employee_name context = %#v, exists = %v, want 张三", employeeName, exists)
		}
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestAuthMiddlewareSkipsEmptyEmployeeName(t *testing.T) {
	gin.SetMode(gin.TestMode)

	jwtManager := utils.NewJWTManager(&config.JWTConfig{Secret: "test-secret", ExpirationHours: 1})
	token, err := jwtManager.GenerateToken("12", "tester", "", []string{"NURSE"}, 3)
	if err != nil {
		t.Fatalf("GenerateToken() error = %v", err)
	}

	r := gin.New()
	r.Use(AuthMiddleware(jwtManager))
	r.GET("/protected", func(c *gin.Context) {
		_, exists := c.Get("employee_name")
		if exists {
			t.Fatal("employee_name should not be set when claim is empty")
		}
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
}

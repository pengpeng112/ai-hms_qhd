package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestLoginRateLimiter(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(LoginRateLimiter())
	r.POST("/auth/login", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	makeRequest := func() int {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/auth/login", nil)
		req.RemoteAddr = "192.168.1.100:12345"
		r.ServeHTTP(w, req)
		return w.Code
	}

	// 前 5 次应成功
	for i := 0; i < 5; i++ {
		if code := makeRequest(); code != http.StatusOK {
			t.Fatalf("attempt %d: expected 200, got %d", i+1, code)
		}
	}

	// 第 6 次应返回 429
	if code := makeRequest(); code != http.StatusTooManyRequests {
		t.Fatalf("attempt 6: expected 429, got %d", code)
	}
}

func TestLoginRateLimiterDifferentIP(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(LoginRateLimiter())
	r.POST("/auth/login", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	makeRequestIP := func(ip string) int {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/auth/login", nil)
		req.RemoteAddr = ip + ":12345"
		r.ServeHTTP(w, req)
		return w.Code
	}

	// IP1: 消耗 5 次额度
	for i := 0; i < 5; i++ {
		makeRequestIP("10.0.0.1")
	}
	// IP1 第六次被限
	if code := makeRequestIP("10.0.0.1"); code != http.StatusTooManyRequests {
		t.Fatalf("IP1 attempt 6: expected 429, got %d", code)
	}

	// IP2: 不受 IP1 影响
	if code := makeRequestIP("10.0.0.2"); code != http.StatusOK {
		t.Fatalf("IP2 attempt: expected 200, got %d", code)
	}
}

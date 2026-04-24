package middleware

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const requestIDHeader = "X-Request-ID"

// RequestLogger 记录请求摘要，便于定位接口与异常响应。
func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		requestID := c.GetHeader(requestIDHeader)
		if requestID == "" {
			requestID = uuid.NewString()
		}
		c.Writer.Header().Set(requestIDHeader, requestID)
		c.Set("request_id", requestID)

		c.Next()

		path := c.Request.URL.Path
		if rawQuery := c.Request.URL.RawQuery; rawQuery != "" {
			path += "?" + rawQuery
		}

		attrs := []any{
			"request_id", requestID,
			"method", c.Request.Method,
			"path", path,
			"status", c.Writer.Status(),
			"latency_ms", time.Since(start).Milliseconds(),
			"client_ip", c.ClientIP(),
			"user_agent", c.Request.UserAgent(),
			"response_bytes", c.Writer.Size(),
		}

		if userID := GetUserID(c); userID != "" {
			attrs = append(attrs, "user_id", userID)
		}
		if tenantID := GetTenantID(c); tenantID > 0 {
			attrs = append(attrs, "tenant_id", tenantID)
		}
		if len(c.Errors) > 0 {
			attrs = append(attrs, "errors", c.Errors.String())
		}

		switch {
		case c.Writer.Status() >= 500:
			slog.Error("http request completed", attrs...)
		case c.Writer.Status() >= 400:
			slog.Warn("http request completed", attrs...)
		default:
			slog.Info("http request completed", attrs...)
		}
	}
}

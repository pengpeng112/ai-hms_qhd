package middleware

import "github.com/gin-gonic/gin"

// SecurityHeaders 为所有响应附加一组安全响应头，降低 XSS / 点击劫持 / MIME 嗅探等风险。
//
// 设计原则：只加不破坏。
//   - 这里的头部均为"防御性"且对现有前端无副作用。
//   - CSP 采用 Report-Only 模式：仅上报违规、不拦截资源，因此不会破坏 Antd/Tailwind/Vite
//     可能产生的内联样式/脚本。待前端联调确认无误后，可将其升级为强制 Content-Security-Policy。
//
// 注：HSTS（Strict-Transport-Security）依赖 HTTPS，应由前置 Nginx 在 TLS 终止处统一下发，
// 不在应用层强加，以免本地/内网 HTTP 环境异常。
func SecurityHeaders() gin.HandlerFunc {
	// Report-Only CSP：先观察、不拦截。后续要强制时，去掉 -Report-Only 并按上报结果收敛白名单，
	// 同时可补 report-uri/report-to 指向收集端点。
	const cspReportOnly = "default-src 'self'; " +
		"script-src 'self'; " +
		"style-src 'self' 'unsafe-inline'; " +
		"img-src 'self' data:; " +
		"font-src 'self' data:; " +
		"connect-src 'self'; " +
		"object-src 'none'; " +
		"base-uri 'self'; " +
		"frame-ancestors 'none'"

	return func(c *gin.Context) {
		h := c.Writer.Header()
		// 禁止浏览器 MIME 嗅探
		h.Set("X-Content-Type-Options", "nosniff")
		// 禁止被任意页面以 iframe 嵌入，防点击劫持
		h.Set("X-Frame-Options", "DENY")
		// 不在跨源跳转时泄露完整 URL（可能含敏感查询参数）
		h.Set("Referrer-Policy", "no-referrer")
		// 关闭老旧且有缺陷的浏览器 XSS 过滤器（现代最佳实践，交由 CSP 负责）
		h.Set("X-XSS-Protection", "0")
		// 限制部分高风险浏览器特性
		h.Set("Permissions-Policy", "geolocation=(), microphone=(), camera=()")
		// CSP 观察模式
		h.Set("Content-Security-Policy-Report-Only", cspReportOnly)

		c.Next()
	}
}

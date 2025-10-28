package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"http-services/api/response"
)

// BodySizeLimit 请求体大小限制中间件
// 默认限制为 10MB，防止过大的请求导致服务器资源耗尽
func BodySizeLimit(maxSize int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 设置请求体最大大小
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxSize)

		c.Next()

		// 检查是否因为请求体过大而出错
		if c.Writer.Status() == http.StatusRequestEntityTooLarge {
			response.ReturnError(c, response.INVALID_ARGUMENT, "request body too large")
			c.Abort()
		}
	}
}

// SecurityHeaders 安全响应头中间件
// 添加常见的安全响应头，防止 XSS、点击劫持等攻击
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 防止浏览器进行 MIME 类型嗅探
		c.Header("X-Content-Type-Options", "nosniff")

		// 防止点击劫持攻击
		c.Header("X-Frame-Options", "DENY")

		// XSS 防护
		c.Header("X-XSS-Protection", "1; mode=block")

		// 严格的传输安全（仅在 HTTPS 时有效）
		// c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")

		// 内容安全策略（根据实际需求调整）
		// c.Header("Content-Security-Policy", "default-src 'self'")

		c.Next()
	}
}

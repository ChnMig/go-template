package middleware

import (
	"http-services/utils/id"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

const (
	// RequestIDKey 是请求 ID 在 context 中的 key
	RequestIDKey = "X-Request-ID"
)

// RequestID 请求 ID 追踪中间件
// 为每个请求生成唯一 ID，并创建带上下文的 logger，方便日志追踪和问题排查
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 尝试从请求头获取 Request ID
		requestID := c.GetHeader(RequestIDKey)

		// 如果请求头中没有，则生成新的 Request ID
		if requestID == "" {
			requestID = id.GenerateID()
		}

		// 设置 Request ID 到 context
		c.Set(RequestIDKey, requestID)

		// 在响应头中返回 Request ID
		c.Header(RequestIDKey, requestID)

		// 创建带上下文信息的 logger 并存入 context
		contextLogger := zap.L().With(
			zap.String("trace_id", requestID),
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.String("client_ip", c.ClientIP()),
		)
		c.Set("logger", contextLogger)

		// 记录请求开始（调试级别）
		contextLogger.Debug("Request started")

		c.Next()

		// 记录请求完成（包含状态码，调试级别）
		contextLogger.Debug("Request completed",
			zap.Int("status_code", c.Writer.Status()),
		)
	}
}

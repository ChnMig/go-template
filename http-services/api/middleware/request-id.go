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
// 为每个请求生成唯一 ID，方便日志追踪和问题排查
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

		// 在日志中记录请求信息
		zap.L().Info("Request",
			zap.String("request_id", requestID),
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.String("ip", c.ClientIP()),
		)

		c.Next()
	}
}

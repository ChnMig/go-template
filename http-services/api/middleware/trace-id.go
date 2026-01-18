package middleware

import (
	"http-services/utils/id"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

const (
	TraceIDHeaderKey  = "X-Trace-ID"
	TraceIDContextKey = "trace_id"
)

func TraceID() gin.HandlerFunc {
	return func(c *gin.Context) {
		traceID := c.GetHeader(TraceIDHeaderKey)
		if traceID == "" {
			traceID = id.GenerateID()
		}

		c.Set(TraceIDContextKey, traceID)
		c.Header(TraceIDHeaderKey, traceID)

		// 创建带上下文信息的 logger 并存入 context
		contextLogger := zap.L().With(
			zap.String("trace_id", traceID),
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

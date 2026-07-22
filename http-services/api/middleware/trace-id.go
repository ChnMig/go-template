package middleware

import (
	"http-services/utils/contextkey"
	"http-services/utils/id"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

const (
	TraceIDHeaderKey  = contextkey.TraceIDHeader
	TraceIDContextKey = contextkey.TraceID
)

func TraceID() gin.HandlerFunc {
	return TraceIDWithDependencies(id.GenerateID, zap.L)
}

// TraceIDWithDependencies 使用显式依赖注入追踪 ID factory 和上下文 logger。
func TraceIDWithDependencies(factory TraceIDFactory, loggerProvider LoggerProvider) gin.HandlerFunc {
	if factory == nil {
		factory = id.GenerateID
	}
	return func(c *gin.Context) {
		traceID := c.GetHeader(TraceIDHeaderKey)
		if !validTraceID(traceID) {
			traceID = factory()
			if !validTraceID(traceID) {
				traceID = id.GenerateID()
			}
		}

		c.Set(TraceIDContextKey, traceID)
		c.Request = c.Request.WithContext(contextkey.WithTraceID(c.Request.Context(), traceID))
		c.Header(TraceIDHeaderKey, traceID)

		// 创建带上下文信息的 logger 并存入 context
		contextLogger := loggerFromProvider(loggerProvider).With(
			zap.String("trace_id", traceID),
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.String("client_ip", c.ClientIP()),
		)
		c.Set(contextkey.Logger, contextLogger)

		// 记录请求开始（调试级别）
		contextLogger.Debug("Request started")

		c.Next()

		// 记录请求完成（包含状态码，调试级别）
		contextLogger.Debug("Request completed",
			zap.Int("status_code", c.Writer.Status()),
		)
	}
}

func validTraceID(traceID string) bool {
	if len(traceID) != 32 {
		return false
	}
	for _, character := range traceID {
		if character < '0' || character > '9' {
			if character < 'a' || character > 'f' {
				return false
			}
		}
	}
	return true
}

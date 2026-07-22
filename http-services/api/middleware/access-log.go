package middleware

import (
	"time"

	"http-services/utils/contextkey"
	httplog "http-services/utils/log"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// AccessLog 记录结构化 Gin access log。
// 成功路径只记录请求摘要字段，避免记录 body 或大量业务参数。
func AccessLog() gin.HandlerFunc {
	return AccessLogWithLogger(httplog.GetGinLogger)
}

// AccessLogWithLogger 使用调用方注入的动态 logger provider 记录访问摘要。
func AccessLogWithLogger(loggerProvider LoggerProvider) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		defer func() {
			responseBytes := c.Writer.Size()
			if responseBytes < 0 {
				responseBytes = 0
			}
			fields := []zap.Field{
				zap.String("method", c.Request.Method),
				zap.String("path", c.Request.URL.Path),
				zap.Int("status", c.Writer.Status()),
				zap.Int("response_bytes", responseBytes),
				zap.Duration("latency", time.Since(start)),
				zap.String("client_ip", c.ClientIP()),
				zap.String("trace_id", traceIDFromContext(c)),
			}

			loggerFromProvider(loggerProvider).Info("HTTP access", fields...)
		}()

		c.Next()
	}
}

func traceIDFromContext(c *gin.Context) string {
	if traceID, exists := c.Get(contextkey.TraceID); exists {
		if id, ok := traceID.(string); ok {
			return id
		}
	}
	return ""
}

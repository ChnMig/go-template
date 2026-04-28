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
	return func(c *gin.Context) {
		start := time.Now()
		defer func() {
			errText := c.Errors.ByType(gin.ErrorTypeAny).String()
			fields := []zap.Field{
				zap.String("method", c.Request.Method),
				zap.String("path", c.Request.URL.Path),
				zap.String("raw_query", c.Request.URL.RawQuery),
				zap.Int("status", c.Writer.Status()),
				zap.Duration("latency", time.Since(start)),
				zap.String("client_ip", c.ClientIP()),
				zap.String("user_agent", c.Request.UserAgent()),
				zap.String("trace_id", traceIDFromContext(c)),
				zap.String("error", errText),
			}

			httplog.GetGinLogger().Info("HTTP access", fields...)
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

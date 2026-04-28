package middleware

import (
	"runtime/debug"

	"http-services/api/response"
	httplog "http-services/utils/log"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Recovery 将 panic 转换为统一响应，并写入带请求上下文的错误日志。
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if recovered := recover(); recovered != nil {
				httplog.GetGinErrorLogger().Error("HTTP panic recovered",
					zap.Any("panic", recovered),
					zap.String("method", c.Request.Method),
					zap.String("path", c.Request.URL.Path),
					zap.String("raw_query", c.Request.URL.RawQuery),
					zap.String("client_ip", c.ClientIP()),
					zap.String("user_agent", c.Request.UserAgent()),
					zap.String("trace_id", traceIDFromContext(c)),
					zap.ByteString("stack", debug.Stack()),
				)
				response.ReturnError(c, response.INTERNAL, "服务内部错误")
			}
		}()

		c.Next()
	}
}

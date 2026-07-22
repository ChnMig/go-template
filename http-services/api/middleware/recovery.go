package middleware

import (
	"errors"
	"fmt"
	"net/http"
	"runtime/debug"
	"syscall"

	"http-services/api/response"
	httplog "http-services/utils/log"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Recovery 将 panic 转换为统一响应，并写入带请求上下文的错误日志。
func Recovery() gin.HandlerFunc {
	return RecoveryWithLogger(httplog.GetGinErrorLogger)
}

// RecoveryWithLogger 使用调用方注入的动态错误 logger provider 恢复 panic。
func RecoveryWithLogger(loggerProvider LoggerProvider) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if recovered := recover(); recovered != nil {
				if isAbortedConnection(recovered) {
					loggerFromProvider(loggerProvider).Warn("HTTP connection aborted",
						zap.String("method", c.Request.Method),
						zap.String("path", c.Request.URL.Path),
						zap.String("client_ip", c.ClientIP()),
						zap.String("trace_id", traceIDFromContext(c)),
						zap.String("panic_type", fmt.Sprintf("%T", recovered)),
					)
					c.Abort()
					return
				}

				committed := c.Writer.Written()
				loggerFromProvider(loggerProvider).Error("HTTP panic recovered",
					zap.String("method", c.Request.Method),
					zap.String("path", c.Request.URL.Path),
					zap.Int("status", recoveryStatus(c, committed)),
					zap.String("client_ip", c.ClientIP()),
					zap.String("trace_id", traceIDFromContext(c)),
					zap.String("panic_type", fmt.Sprintf("%T", recovered)),
					zap.ByteString("stack", debug.Stack()),
				)
				c.Abort()
				if !committed {
					response.ReturnError(c, response.INTERNAL, "服务内部错误")
				}
			}
		}()

		c.Next()
	}
}

func recoveryStatus(c *gin.Context, committed bool) int {
	if committed {
		return c.Writer.Status()
	}
	return http.StatusInternalServerError
}

func isAbortedConnection(recovered any) bool {
	err, ok := recovered.(error)
	return ok && (errors.Is(err, syscall.EPIPE) || errors.Is(err, syscall.ECONNRESET) || errors.Is(err, http.ErrAbortHandler))
}

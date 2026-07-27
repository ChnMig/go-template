package middleware

import (
	"errors"
	"fmt"
	"net/http"
	"runtime/debug"
	"syscall"

	"http-services/api/response"
	serviceLog "http-services/utils/log"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Recovery converts pre-write panics to the internal envelope and preserves committed responses.
func Recovery() gin.HandlerFunc {
	return RecoveryWithLogger(serviceLog.GetGinErrorLogger)
}

// RecoveryWithLogger recovers panics through a dynamic global error logger provider.
func RecoveryWithLogger(loggerProvider LoggerProvider) gin.HandlerFunc {
	return func(context *gin.Context) {
		defer func() {
			recovered := recover()
			if recovered == nil {
				return
			}

			if isAbortedConnection(recovered) {
				fields := []zap.Field{
					zap.String(logKeyMethod, context.Request.Method),
					zap.String(logKeyPath, context.Request.URL.Path),
					zap.Any("panic", recovered),
					zap.String(logKeyPanicType, fmt.Sprintf("%T", recovered)),
				}
				fields = append(fields, serviceLog.RequestFields(context)...)
				loggerFromContext(context.Request.Context(), loggerProvider).Warn("http.connection_aborted", fields...)
				context.Abort()
				return
			}

			committed := context.Writer.Written()
			status := context.Writer.Status()
			if !committed {
				status = http.StatusInternalServerError
			}
			fields := []zap.Field{
				zap.String(logKeyMethod, context.Request.Method),
				zap.String(logKeyPath, context.Request.URL.Path),
				zap.Int(logKeyStatus, status),
				zap.Any("panic", recovered),
				zap.String(logKeyPanicType, fmt.Sprintf("%T", recovered)),
				zap.ByteString(logKeyStack, debug.Stack()),
			}
			fields = append(fields, serviceLog.RequestFields(context)...)
			loggerFromContext(context.Request.Context(), loggerProvider).Error("http.panic_recovered", fields...)
			context.Abort()
			if !committed {
				response.ReturnError(context, response.INTERNAL, "internal server error")
			}
		}()

		context.Next()
	}
}

func isAbortedConnection(recovered any) bool {
	err, ok := recovered.(error)
	return ok && (errors.Is(err, syscall.EPIPE) || errors.Is(err, syscall.ECONNRESET) || errors.Is(err, http.ErrAbortHandler))
}

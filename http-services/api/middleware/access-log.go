package middleware

import (
	"time"

	serviceLog "http-services/utils/log"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

const (
	logKeyMethod        = "method"
	logKeyPath          = "path"
	logKeyClientIP      = "client_ip"
	logKeyStatus        = "status"
	logKeyResponseBytes = "response_bytes"
	logKeyElapsed       = "elapsed"
	logKeyPanicType     = "panic_type"
	logKeyStack         = "stack"
)

// AccessLog records one safe request summary after downstream completion.
func AccessLog() gin.HandlerFunc {
	return AccessLogWithLogger(serviceLog.GetGinLogger)
}

// AccessLogWithLogger records access logs through a dynamic global logger provider.
func AccessLogWithLogger(loggerProvider LoggerProvider) gin.HandlerFunc {
	return func(context *gin.Context) {
		startedAt := time.Now()
		defer func() {
			if isProbePath(context.Request.URL.Path) {
				return
			}
			responseBytes := context.Writer.Size()
			if responseBytes < 0 {
				responseBytes = 0
			}
			loggerFromContext(context.Request.Context(), loggerProvider).Info(
				"http.request",
				zap.String(logKeyMethod, context.Request.Method),
				zap.String(logKeyPath, context.Request.URL.Path),
				zap.String(logKeyClientIP, context.ClientIP()),
				zap.Int(logKeyStatus, context.Writer.Status()),
				zap.Int(logKeyResponseBytes, responseBytes),
				zap.Duration(logKeyElapsed, time.Since(startedAt)),
			)
		}()

		context.Next()
	}
}

func isProbePath(path string) bool {
	return path == "/api/v1/open/health"
}

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

// AccessLog records one request summary through the global Gin zap logger.
func AccessLog() gin.HandlerFunc {
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
			logger := serviceLog.GetGinLogger()
			if traceID, ok := serviceLog.TraceID(context.Request.Context()); ok {
				logger = logger.With(zap.String("trace_id", traceID))
			}
			logger.Info(
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

package middleware

import (
	"strings"

	"http-services/api/response"
	"http-services/utils/contextkey"
	"http-services/utils/id"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

const (
	TraceIDHeaderKey  = contextkey.TraceIDHeader
	TraceIDContextKey = contextkey.TraceID
)

func TraceID() gin.HandlerFunc {
	return func(c *gin.Context) {
		traceID := c.GetHeader(TraceIDHeaderKey)
		parsedTraceID, parseErr := uuid.Parse(traceID)
		validTraceID := parseErr == nil && len(traceID) == 36 && strings.EqualFold(parsedTraceID.String(), traceID)
		if !validTraceID {
			generated, err := id.GenerateUUIDv7()
			if err != nil {
				zap.L().Error("generate trace ID failed", zap.Error(err))
				response.ReturnError(c, response.INTERNAL, "internal server error")
				return
			}
			traceID = generated.String()
		}

		c.Set(TraceIDContextKey, traceID)
		c.Header(TraceIDHeaderKey, traceID)

		contextLogger := zap.L().With(
			zap.String("trace_id", traceID),
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.String("client_ip", c.ClientIP()),
		)
		c.Set(contextkey.Logger, contextLogger)

		contextLogger.Debug("Request started")

		c.Next()

		contextLogger.Debug("Request completed",
			zap.Int("status_code", c.Writer.Status()),
		)
	}
}

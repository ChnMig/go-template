package middleware

import (
	"strings"

	"http-services/api/response"
	"http-services/utils/contextkey"
	"http-services/utils/id"
	serviceLog "http-services/utils/log"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// TraceIDHeader is the request and response correlation header.
const TraceIDHeader = serviceLog.TraceIDHeader

const (
	TraceIDHeaderKey  = TraceIDHeader
	TraceIDContextKey = contextkey.TraceID
)

func TraceID() gin.HandlerFunc {
	return func(c *gin.Context) {
		traceID := c.GetHeader(TraceIDHeader)
		parsedTraceID, parseErr := uuid.Parse(traceID)
		validTraceID := parseErr == nil && len(traceID) == 36 && strings.EqualFold(parsedTraceID.String(), traceID)
		if !validTraceID {
			generated, err := id.GenerateUUIDv7()
			if err != nil {
				serviceLog.WithRequest(c).Error("generate trace ID failed", zap.Error(err))
				response.ReturnError(c, response.INTERNAL, "internal server error")
				return
			}
			traceID = generated.String()
		}

		c.Set(TraceIDContextKey, traceID)
		c.Request = c.Request.WithContext(serviceLog.WithTraceID(c.Request.Context(), traceID))
		c.Header(TraceIDHeader, traceID)

		c.Set(contextkey.Logger, zap.L())
		serviceLog.CaptureRequestBody(c)
		contextLogger := serviceLog.FromContext(c)

		contextLogger.Debug("http.request.started")
		c.Next()
		contextLogger.Debug("http.request.completed", zap.Int("status", c.Writer.Status()))
	}
}

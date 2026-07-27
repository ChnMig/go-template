// Package middleware provides the service's ordered HTTP middleware chain.
package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"go.uber.org/zap"
	"http-services/api/response"
	"http-services/utils/contextkey"
	serviceLog "http-services/utils/log"
)

// TraceIDHeader is the request and response correlation header.
const TraceIDHeader = serviceLog.TraceIDHeader

const (
	TraceIDHeaderKey  = TraceIDHeader
	TraceIDContextKey = contextkey.TraceID
)

// TraceID validates or generates a trace identifier and adds it to request context.
func TraceID(factory IDFactory) gin.HandlerFunc {
	return TraceIDWithDependencies(factory, zap.L)
}

// TraceIDWithDependencies installs the trace ID and request-scoped logger.
func TraceIDWithDependencies(factory IDFactory, loggerProvider LoggerProvider) gin.HandlerFunc {
	return func(context *gin.Context) {
		traceID := context.GetHeader(TraceIDHeader)
		parsedTraceID, parseErr := uuid.Parse(traceID)
		validTraceID := parseErr == nil && len(traceID) == 36 && strings.EqualFold(parsedTraceID.String(), traceID)
		if !validTraceID {
			generated, generationErr := factory()
			if generationErr != nil {
				response.ReturnError(context, response.INTERNAL, "internal server error")
				return
			}
			traceID = generated.String()
		}

		context.Set(contextkey.TraceID, traceID)
		context.Request = context.Request.WithContext(serviceLog.WithTraceID(context.Request.Context(), traceID))
		context.Header(TraceIDHeader, traceID)
		requestLogger := loggerFromProvider(loggerProvider).With(
			zap.String("trace_id", traceID),
			zap.String(logKeyMethod, context.Request.Method),
			zap.String(logKeyPath, context.Request.URL.Path),
			zap.String(logKeyClientIP, context.ClientIP()),
		)
		context.Set(contextkey.Logger, requestLogger)
		requestLogger.Debug("http.request.started")
		context.Next()
		requestLogger.Debug("http.request.completed", zap.Int(logKeyStatus, context.Writer.Status()))
	}
}

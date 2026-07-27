package log

import (
	"context"

	"http-services/utils/contextkey"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

const traceIDKey = "trace_id"

// TraceIDHeader carries request correlation across HTTP service boundaries.
const TraceIDHeader = contextkey.TraceIDHeader

// WithTraceID returns a child context carrying the request trace identifier.
func WithTraceID(ctx context.Context, traceID string) context.Context {
	return contextkey.WithTraceID(ctx, traceID)
}

// TraceID returns the request trace identifier stored in ctx.
func TraceID(ctx context.Context) (string, bool) {
	return contextkey.TraceIDFromContext(ctx)
}

// FromStandardContext derives the global logger from a standard context.
func FromStandardContext(ctx context.Context) *zap.Logger {
	logger := GetLogger()
	traceID, ok := TraceID(ctx)
	if !ok {
		return logger
	}
	return logger.With(zap.String(traceIDKey, traceID))
}

// FromContext returns the request-scoped Gin logger when middleware installed one.
func FromContext(ctx *gin.Context) *zap.Logger {
	if ctx != nil {
		if value, exists := ctx.Get(contextkey.Logger); exists {
			if logger, ok := value.(*zap.Logger); ok && logger != nil {
				return logger
			}
		}
		if ctx.Request != nil {
			return FromStandardContext(ctx.Request.Context())
		}
	}
	return GetLogger()
}

// WithRequest returns the safe request-scoped logger installed by middleware.
// Query strings, forms, bound values, bodies, and credentials are never added.
func WithRequest(ctx *gin.Context) *zap.Logger {
	return FromContext(ctx)
}

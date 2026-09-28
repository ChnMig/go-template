package log

import (
	"context"

	"http-services/utils/contextkey"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

const traceIDKey = "trace_id"

// BoundParamsKey stores bound business parameters in gin.Context for request logging.
const BoundParamsKey = contextkey.BoundParams

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

// FromContext 从基础 logger 派生请求日志，统一附加一次请求元数据。
// 没有 Gin TraceID 时继续从标准 context 读取，保留下游传递能力。
func FromContext(ctx *gin.Context) *zap.Logger {
	var base *zap.Logger
	if ctx != nil {
		if value, exists := ctx.Get(contextkey.Logger); exists {
			base, _ = value.(*zap.Logger)
		}
	}
	if base == nil {
		base = GetLogger()
	}
	if ctx == nil {
		return base
	}
	fields := make([]zap.Field, 0, 4)
	traceID := ctx.GetString(contextkey.TraceID)
	if traceID == "" && ctx.Request != nil {
		traceID, _ = TraceID(ctx.Request.Context())
	}
	if traceID != "" {
		fields = append(fields, zap.String(traceIDKey, traceID))
	}
	if ctx.Request != nil {
		fields = append(fields, zap.String("method", ctx.Request.Method), zap.String("client_ip", ctx.ClientIP()))
		if ctx.Request.URL != nil {
			fields = append(fields, zap.String("path", ctx.Request.URL.Path))
		}
	}
	return base.With(fields...)
}

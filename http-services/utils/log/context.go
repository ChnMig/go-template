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

// RequestFields returns all request inputs captured by the HTTP middleware.
func RequestFields(ctx *gin.Context) []zap.Field {
	if ctx == nil || ctx.Request == nil {
		return nil
	}
	fields := make([]zap.Field, 0, 7)
	if ctx.Request.URL != nil && ctx.Request.URL.RawQuery != "" {
		fields = append(fields, zap.String("query", ctx.Request.URL.RawQuery))
	}
	if len(ctx.Request.Header) > 0 {
		fields = append(fields, zap.Any("headers", ctx.Request.Header))
	}
	if len(ctx.Request.PostForm) > 0 {
		fields = append(fields, zap.Any("form", ctx.Request.PostForm))
	}
	if ctx.Request.MultipartForm != nil && len(ctx.Request.MultipartForm.Value) > 0 {
		fields = append(fields, zap.Any("multipart_form", ctx.Request.MultipartForm.Value))
	}
	if len(ctx.Params) > 0 {
		params := make(map[string]string, len(ctx.Params))
		for _, param := range ctx.Params {
			params[param.Key] = param.Value
		}
		fields = append(fields, zap.Any("path_params", params))
	}
	if bound, exists := ctx.Get(contextkey.BoundParams); exists {
		fields = append(fields, zap.Any("params", bound))
	}
	if value, exists := ctx.Get(contextkey.RequestBody); exists {
		if capture, ok := value.(*contextkey.RequestBodyCapture); ok && len(capture.Bytes) > 0 {
			fields = append(fields, zap.ByteString("body", capture.Bytes))
		}
	}
	return fields
}

// WithRequest returns a request-scoped logger carrying the complete captured input.
func WithRequest(ctx *gin.Context) *zap.Logger {
	return FromContext(ctx).With(RequestFields(ctx)...)
}

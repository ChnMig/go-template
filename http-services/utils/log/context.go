package log

import (
	"context"
	"net/http"

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

// WithRequest adds all parsed request parameters to the request-scoped logger.
// It deliberately keeps full values for troubleshooting; callers decide when to use it.
func WithRequest(ctx *gin.Context) *zap.Logger {
	base := FromContext(ctx)
	if ctx == nil || ctx.Request == nil {
		return base
	}

	fields := []zap.Field{zap.String("method", ctx.Request.Method)}
	if ctx.Request.URL != nil {
		fields = append(fields, zap.String("path", ctx.Request.URL.Path))
		if rawQuery := ctx.Request.URL.RawQuery; rawQuery != "" {
			fields = append(fields, zap.String("query", rawQuery))
		}
	}
	if ctx.Request.Method == http.MethodPost ||
		ctx.Request.Method == http.MethodPut ||
		ctx.Request.Method == http.MethodPatch {
		if len(ctx.Request.PostForm) > 0 {
			fields = append(fields, zap.Any("form", ctx.Request.PostForm))
		}
		if ctx.Request.MultipartForm != nil && len(ctx.Request.MultipartForm.Value) > 0 {
			fields = append(fields, zap.Any("multipart_form", ctx.Request.MultipartForm.Value))
		}
	}
	if len(ctx.Params) > 0 {
		pathParams := make(map[string]string, len(ctx.Params))
		for _, param := range ctx.Params {
			pathParams[param.Key] = param.Value
		}
		fields = append(fields, zap.Any("path_params", pathParams))
	}
	if bound, exists := ctx.Get(BoundParamsKey); exists && bound != nil {
		fields = append(fields, zap.Any("params", bound))
	}
	return base.With(fields...)
}

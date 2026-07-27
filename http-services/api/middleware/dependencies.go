package middleware

import (
	"context"

	httplog "http-services/utils/log"

	"go.uber.org/zap"
)

// LoggerProvider 返回当前有效 logger，支持日志配置热切换。
type LoggerProvider func() *zap.Logger

func loggerFromProvider(provider LoggerProvider) *zap.Logger {
	if provider == nil {
		return zap.NewNop()
	}
	logger := provider()
	if logger == nil {
		return zap.NewNop()
	}
	return logger
}

func loggerFromContext(ctx context.Context, provider LoggerProvider) *zap.Logger {
	logger := loggerFromProvider(provider)
	if traceID, ok := httplog.TraceID(ctx); ok {
		return logger.With(zap.String("trace_id", traceID))
	}
	return logger
}

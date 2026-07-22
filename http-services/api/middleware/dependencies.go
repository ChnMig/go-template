package middleware

import "go.uber.org/zap"

// LoggerProvider 返回当前有效 logger，支持日志配置热切换。
type LoggerProvider func() *zap.Logger

// TraceIDFactory 生成模板规范的 32 位小写十六进制追踪 ID。
type TraceIDFactory func() string

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

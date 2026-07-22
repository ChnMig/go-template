package api

import (
	"errors"
	"fmt"
	"strings"

	"http-services/api/app"
	"http-services/api/middleware"
	"http-services/config"
	"http-services/utils/id"
	httplog "http-services/utils/log"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// ErrInvalidOptions 标识 Router 依赖缺失或配置无效。
var ErrInvalidOptions = errors.New("api: invalid router options")

// Registrar 在 /api 分组下挂载版本化业务路由。
type Registrar func(*gin.RouterGroup)

// LoggerProviders 提供支持运行期日志切换的三类 logger。
type LoggerProviders struct {
	Context middleware.LoggerProvider
	Access  middleware.LoggerProvider
	Error   middleware.LoggerProvider
}

// Options 包含 Router 的全部显式依赖和启动期 HTTP 配置快照。
type Options struct {
	Server         config.HTTPConfig
	Loggers        LoggerProviders
	TraceIDFactory middleware.TraceIDFactory
	RegisterRoutes Registrar
}

// DefaultOptions 组装模板生产默认依赖，同时保留 NewRouter 的可测试边界。
func DefaultOptions(server config.HTTPConfig) Options {
	return Options{
		Server: server,
		Loggers: LoggerProviders{
			Context: zap.L,
			Access:  httplog.GetGinLogger,
			Error:   httplog.GetGinErrorLogger,
		},
		TraceIDFactory: id.GenerateID,
		RegisterRoutes: app.RegisterRoutes,
	}
}

// NewRouter 构建有序中间件链并挂载调用方注入的业务路由。
func NewRouter(options Options) (*gin.Engine, error) {
	if err := validateOptions(options); err != nil {
		return nil, err
	}
	if err := options.Server.Validate(); err != nil {
		return nil, fmt.Errorf("%w: server config: %w", ErrInvalidOptions, err)
	}

	gin.DefaultWriter = httplog.NewZapWriterFunc(options.Loggers.Access, zapcore.InfoLevel)
	gin.DefaultErrorWriter = httplog.NewZapWriterFunc(options.Loggers.Error, zapcore.ErrorLevel)
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	if err := router.SetTrustedProxies(options.Server.TrustedProxies); err != nil {
		return nil, fmt.Errorf("%w: trusted proxies: %w", ErrInvalidOptions, err)
	}

	router.Use(
		middleware.TraceIDWithDependencies(options.TraceIDFactory, options.Loggers.Context),
		middleware.AccessLogWithLogger(options.Loggers.Access),
		middleware.RecoveryWithLogger(options.Loggers.Error),
	)
	if options.Server.EnableCORS {
		router.Use(middleware.CorsDomainHandler())
	}
	if options.Server.EnableRateLimit {
		router.Use(middleware.IPRateLimit(options.Server.GlobalRateLimit, options.Server.GlobalRateBurst))
	}
	router.Use(
		middleware.SecurityHeaders(),
		middleware.BodySizeLimit(options.Server.MaxBodySize),
	)
	if staticDir := strings.TrimSpace(options.Server.StaticDir); staticDir != "" {
		router.Static("/static", staticDir)
	}

	apiGroup := router.Group("/api")
	options.RegisterRoutes(apiGroup)
	return router, nil
}

// InitApi 保留旧调用方式；生产 bootstrap 使用 NewRouter 传播构建错误。
func InitApi() *gin.Engine {
	options := DefaultOptions(config.SnapshotHTTPConfig())
	router, err := NewRouter(options)
	if err == nil {
		return router
	}

	logger := middlewareLogger(options.Loggers.Error)
	logger.Error("Router 构建失败，兼容入口已禁用转发头", zap.Error(err))
	options.Server.TrustedProxies = nil
	router, fallbackErr := NewRouter(options)
	if fallbackErr == nil {
		return router
	}
	logger.Error("Router 安全回退失败", zap.Error(fallbackErr))
	return gin.New()
}

func validateOptions(options Options) error {
	switch {
	case options.Loggers.Context == nil:
		return fmt.Errorf("%w: context logger", ErrInvalidOptions)
	case options.Loggers.Context() == nil:
		return fmt.Errorf("%w: context logger returned nil", ErrInvalidOptions)
	case options.Loggers.Access == nil:
		return fmt.Errorf("%w: access logger", ErrInvalidOptions)
	case options.Loggers.Access() == nil:
		return fmt.Errorf("%w: access logger returned nil", ErrInvalidOptions)
	case options.Loggers.Error == nil:
		return fmt.Errorf("%w: error logger", ErrInvalidOptions)
	case options.Loggers.Error() == nil:
		return fmt.Errorf("%w: error logger returned nil", ErrInvalidOptions)
	case options.TraceIDFactory == nil:
		return fmt.Errorf("%w: trace ID factory", ErrInvalidOptions)
	case options.RegisterRoutes == nil:
		return fmt.Errorf("%w: route registrar", ErrInvalidOptions)
	default:
		return nil
	}
}

func middlewareLogger(provider middleware.LoggerProvider) *zap.Logger {
	if provider == nil {
		return zap.NewNop()
	}
	logger := provider()
	if logger == nil {
		return zap.NewNop()
	}
	return logger
}

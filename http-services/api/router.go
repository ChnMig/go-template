package api

import (
	"errors"
	"fmt"

	"http-services/api/app"
	"http-services/api/middleware"
	"http-services/config"
	httplog "http-services/utils/log"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// ErrInvalidOptions 标识 Router 依赖缺失或配置无效。
var ErrInvalidOptions = errors.New("api: invalid router options")

// Registrar 在 /api 分组下挂载版本化业务路由。
type Registrar func(*gin.RouterGroup)

// Options 包含 Router 的全部显式依赖和启动期 HTTP 配置快照。
type Options struct {
	Server         config.HTTPConfig
	RegisterRoutes Registrar
}

// DefaultOptions 组装模板生产默认依赖，同时保留 NewRouter 的可测试边界。
func DefaultOptions(server config.HTTPConfig) Options {
	return Options{
		Server:         server,
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

	gin.DefaultWriter = httplog.NewZapWriterFunc(httplog.GetGinLogger, zapcore.InfoLevel)
	gin.DefaultErrorWriter = httplog.NewZapWriterFunc(httplog.GetGinErrorLogger, zapcore.ErrorLevel)
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	if err := router.SetTrustedProxies(options.Server.TrustedProxies); err != nil {
		return nil, fmt.Errorf("%w: trusted proxies: %w", ErrInvalidOptions, err)
	}

	router.Use(
		middleware.TraceID(),
		middleware.AccessLog(),
		middleware.Recovery(),
	)
	if options.Server.EnableCORS {
		router.Use(middleware.CorsDomainHandler())
	}
	if options.Server.EnableRateLimit {
		router.Use(middleware.IPRateLimit(options.Server.GlobalRateLimit, options.Server.GlobalRateBurst))
	}
	router.Use(
		middleware.SecurityHeaders(),
		middleware.BodySizeLimit(config.ByteSize(options.Server.MaxBodySize)),
	)
	registerStatic(router, options.Server.StaticDir)

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

	zap.L().Error("Router 构建失败，兼容入口已禁用转发头", zap.Error(err))
	options.Server.TrustedProxies = nil
	router, fallbackErr := NewRouter(options)
	if fallbackErr == nil {
		return router
	}
	zap.L().Error("Router 安全回退失败", zap.Error(fallbackErr))
	return gin.New()
}

func validateOptions(options Options) error {
	if options.RegisterRoutes == nil {
		return fmt.Errorf("%w: route registrar", ErrInvalidOptions)
	}
	return nil
}

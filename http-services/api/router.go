package api

import (
	"http-services/api/app/example"
	"http-services/api/app/health"
	"http-services/api/middleware"
	"http-services/config"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// open
func openRouter(router *gin.RouterGroup) {
	// 示例：为开放接口添加 IP 限流（每秒10个请求，突发20个）
	exampleRouter := router.Group("/open/example", middleware.IPRateLimit(10, 20))
	{
		exampleRouter.GET("/pong", example.Pong)
		exampleRouter.POST("/token", example.CreateToken)
	}
}

// private
func privateRouter(router *gin.RouterGroup) {
	// 示例：为私有接口添加 Token 限流（每秒100个请求，突发200个）
	exampleRouter := router.Group("/private/example", middleware.TokenVerify, middleware.TokenRateLimit(100, 200))
	{
		exampleRouter.GET("/pong", example.Pong)
	}
}

// InitApi 初始化 API 路由
func InitApi() *gin.Engine {
	// gin.Default uses Use by default. Two global middlewares are added, Logger(), Recovery(), Logger is to print logs, Recovery is panic and returns 500
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	// https://pkg.go.dev/github.com/gin-gonic/gin#readme-don-t-trust-all-proxies
	router.SetTrustedProxies(nil)

	// 全局中间件（按推荐顺序：限流 → 安全 → 请求ID → 监控 → 跨域）
	// 1. 全局限流（如果启用）- 最先执行，防止恶意请求消耗资源
	if config.EnableRateLimit {
		router.Use(middleware.IPRateLimit(config.GlobalRateLimit, config.GlobalRateBurst))
	}

	// 2. 安全响应头 - 早期设置安全策略
	router.Use(middleware.SecurityHeaders())

	// 3. 请求 ID 追踪 - 用于日志关联
	router.Use(middleware.RequestID())

	// 4. Prometheus 监控 - 记录请求指标
	router.Use(middleware.Metrics())

	// 5. 请求体大小限制 - 使用配置值
	router.Use(middleware.BodySizeLimit(config.MaxBodySize))

	// 6. 跨域处理 - 在业务逻辑前处理
	router.Use(middleware.CorssDomainHandler())

	// 健康检查端点（不需要认证，不受限流影响）
	router.GET("/health", health.Health)
	router.GET("/ready", health.Ready)

	// Prometheus metrics 端点
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// static
	router.Static("/static", "./static")
	// api-v1
	// Using version control for iteration
	v1 := router.Group("/api/v1")
	{
		openRouter(v1)
		privateRouter(v1)
	}
	return router
}

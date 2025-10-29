package api

import (
	"http-services/api/app"
	"http-services/api/middleware"
	"http-services/config"

	"github.com/gin-gonic/gin"
)

// InitApi 初始化 API 路由
// 顶层仅负责：gin 初始化、全局中间件、挂载 /api 分组
// 具体业务路由由 app 层逐级（app -> v1 -> open/private -> module）注册
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

	// 4. 取消 Prometheus 监控中间件（不需要 metrics）

	// 5. 请求体大小限制 - 使用配置值
	router.Use(middleware.BodySizeLimit(config.MaxBodySize))

	// 6. 跨域处理 - 在业务逻辑前处理
	router.Use(middleware.CorssDomainHandler())

	// 健康检查端点已移动到 openRouter（/api/v1/open/health）

	// 移除 Prometheus metrics 端点（不需要 metrics）

	// static
	router.Static("/static", "./static")

	// /api 分组，业务路由由 app 层递归注册
	apiGroup := router.Group("/api")
	app.RegisterRoutes(apiGroup)
	return router
}

package api

import (
	healthopen "http-services/api/app/v1/open/health"
	"http-services/api/middleware"
	"http-services/config"

	"github.com/gin-gonic/gin"
)

// open 层级路由：仅负责定义 open 分组，并交由各 app 注册自身子路由
func openRouter(router *gin.RouterGroup) {
	// /api/v1/open
	open := router.Group("/open")
	// 各 app 负责在自身包内声明子路由（更清晰的分层）
	// 将健康检查放入 open 分组（由 health 模块自行注册）
	healthopen.RegisterOpenRoutes(open)
}

// private 层级路由：仅负责定义 private 分组，并交由各 app 注册自身子路由
func privateRouter(router *gin.RouterGroup) {
	// 预留：/api/v1/private 下当前无接口
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

	// 4. 取消 Prometheus 监控中间件（不需要 metrics）

	// 5. 请求体大小限制 - 使用配置值
	router.Use(middleware.BodySizeLimit(config.MaxBodySize))

	// 6. 跨域处理 - 在业务逻辑前处理
	router.Use(middleware.CorssDomainHandler())

	// 健康检查端点已移动到 openRouter（/api/v1/open/health）

	// 移除 Prometheus metrics 端点（不需要 metrics）

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

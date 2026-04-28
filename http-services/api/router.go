package api

import (
	"http-services/api/app"
	"http-services/api/middleware"
	"http-services/config"
	httplog "http-services/utils/log"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap/zapcore"
)

// InitApi 初始化 API 路由
// 顶层仅负责：gin 初始化、全局中间件、挂载 /api 分组
// 具体业务路由由 app 层逐级（app -> v1 -> open/private -> module）注册
func InitApi() *gin.Engine {
	// 将 gin 的默认日志输出重定向到 zap（Gin 独立日志文件），避免与业务日志混在同一个文件
	// 注意：main 中会在 InitApi 之前完成 zap 初始化
	ginLogWriter := httplog.NewZapWriterFunc(httplog.GetGinLogger, zapcore.InfoLevel)
	ginErrorWriter := httplog.NewZapWriterFunc(httplog.GetGinErrorLogger, zapcore.ErrorLevel)
	gin.DefaultWriter = ginLogWriter
	gin.DefaultErrorWriter = ginErrorWriter

	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	// https://pkg.go.dev/github.com/gin-gonic/gin#readme-don-t-trust-all-proxies
	router.SetTrustedProxies(nil)

	// 全局中间件：先注入 trace_id，再让 access log 包住 recovery。
	// handler panic 时 Recovery 先写统一响应，AccessLog 的 defer 再记录最终状态。
	router.Use(middleware.TraceID())
	router.Use(middleware.AccessLog())
	router.Use(middleware.Recovery())

	// 1. 全局限流（如果启用）
	if config.EnableRateLimit {
		router.Use(middleware.IPRateLimit(config.GlobalRateLimit, config.GlobalRateBurst))
	}

	// 2. 安全响应头
	router.Use(middleware.SecurityHeaders())

	// 4. 取消 Prometheus 监控中间件（不需要 metrics）

	// 5. 请求体大小限制 - 使用配置值
	router.Use(middleware.BodySizeLimit(config.MaxBodySize))

	// 6. 跨域处理 - 在业务逻辑前处理
	router.Use(middleware.CorsDomainHandler())

	// 健康检查端点已移动到 openRouter（/api/v1/open/health）

	// 移除 Prometheus metrics 端点（不需要 metrics）

	// static
	router.Static("/static", "./static")

	// /api 分组，业务路由由 app 层递归注册
	apiGroup := router.Group("/api")
	app.RegisterRoutes(apiGroup)
	return router
}

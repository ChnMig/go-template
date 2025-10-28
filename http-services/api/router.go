package api

import (
	"http-services/api/app/example"
	"http-services/api/middleware"

	"github.com/gin-gonic/gin"
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

// InitApi init gshop app
func InitApi() *gin.Engine {
	// gin.Default uses Use by default. Two global middlewares are added, Logger(), Recovery(), Logger is to print logs, Recovery is panic and returns 500
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	// https://pkg.go.dev/github.com/gin-gonic/gin#readme-don-t-trust-all-proxies
	router.SetTrustedProxies(nil)
	// Add consent cross-domain middleware
	router.Use(middleware.CorssDomainHandler())
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

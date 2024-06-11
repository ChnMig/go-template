package api

import (
	"http-server/api/app/example"
	"http-server/api/middleware"

	"github.com/gin-gonic/gin"
)

// open
func openRouter(router *gin.RouterGroup) {
	exampleRouter := router.Group("/open/example")
	{
		exampleRouter.GET("/pong", example.Pong)
		exampleRouter.POST("/login", example.TokenPong)
	}
}

// private
func privateRouter(router *gin.RouterGroup) {
	exampleRouter := router.Group("/private/example", middleware.TokenVerify)
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

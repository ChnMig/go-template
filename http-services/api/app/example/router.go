package example

import (
    "http-services/api/middleware"

    "github.com/gin-gonic/gin"
)

// RegisterOpenRoutes 注册 example 模块的开放接口子路由
// 路径前缀：/api/v1/open/example
// 说明：保持原有限流语义不变（每秒10个请求，突发20个）
func RegisterOpenRoutes(open *gin.RouterGroup) {
    if open == nil {
        return
    }
    r := open.Group("/example", middleware.IPRateLimit(10, 20))
    {
        r.GET("/pong", Pong)
    }
}

// RegisterPrivateRoutes 注册 example 模块的私有接口子路由
// 路径前缀：/api/v1/private/example
// 说明：保持原有认证与限流语义（Token 校验 + Token 限流：每秒100个请求、突发200个）
func RegisterPrivateRoutes(private *gin.RouterGroup) {
    if private == nil {
        return
    }
    r := private.Group("/example", middleware.TokenVerify, middleware.TokenRateLimit(100, 200))
    {
        r.GET("/pong", Pong)
    }
}

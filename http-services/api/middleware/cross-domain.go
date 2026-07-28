package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

const (
	corsAllowedMethods = "GET, POST, PUT, PATCH, DELETE, HEAD, OPTIONS"
	corsAllowedHeaders = "Authorization, Content-Type, X-Trace-ID"
)

// CorsDomainHandler 创建默认跨域处理中间件。
// 脚手架默认纯放开跨域，业务项目需要收紧时可在项目内自行替换。
func CorsDomainHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		if origin != "" {
			c.Header("Access-Control-Allow-Origin", "*")
			c.Header("Access-Control-Allow-Methods", corsAllowedMethods)
			c.Header("Access-Control-Allow-Headers", corsAllowedHeaders)
			c.Header("Access-Control-Expose-Headers", TraceIDHeaderKey)
			c.Header("Access-Control-Max-Age", "172800")
			c.Header("Access-Control-Allow-Credentials", "false")
		}

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

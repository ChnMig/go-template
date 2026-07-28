package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// CorsDomainHandler 创建默认跨域处理中间件。
// 脚手架默认纯放开跨域，业务项目需要收紧时可在项目内自行替换。
func CorsDomainHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Request.Method
		origin := c.Request.Header.Get("Origin")
		if origin != "" {
			c.Header("Access-Control-Allow-Origin", "*")
			c.Header("Access-Control-Allow-Methods", "*")
			c.Header("Access-Control-Allow-Headers", "*")
			c.Header("Access-Control-Expose-Headers", "*")
			c.Header("Access-Control-Max-Age", "172800")
			c.Header("Access-Control-Allow-Credentials", "false")
		}

		if method == http.MethodOptions {
			c.JSON(http.StatusOK, "Options Request!")
			return
		}

		c.Next()
	}
}

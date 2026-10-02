package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// CorssDomainHandler consent cross-domain middleware
func CorssDomainHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Request.Method               // method
		origin := c.Request.Header.Get("Origin") // header
		if origin != "" {
			c.Header("Access-Control-Allow-Origin", "*")  // This is to allow access to all domains
			c.Header("Access-Control-Allow-Methods", "*") // All cross-domain request methods supported by the server, in order to avoid multiple'pre-check' requests for browsing requests
			// header
			c.Header("Access-Control-Allow-Headers", "*")
			c.Header("Access-Control-Expose-Headers", "*")
			c.Header("Access-Control-Max-Age", "172800")
		}
		// Release all OPTIONS methods
		if method == http.MethodOptions {
			c.String(http.StatusOK, "Options Request!")
			c.Abort()
			return
		}
		// Processing request
		c.Next()
	}
}

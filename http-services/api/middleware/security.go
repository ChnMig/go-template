package middleware

import "github.com/gin-gonic/gin"

// SecurityHeaders sets the service's deterministic browser hardening headers.
func SecurityHeaders() gin.HandlerFunc {
	return func(context *gin.Context) {
		context.Header("X-Content-Type-Options", "nosniff")
		context.Header("X-Frame-Options", "DENY")
		context.Header("X-XSS-Protection", "1; mode=block")
		context.Next()
	}
}

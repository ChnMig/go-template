package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"

	"go-services/api/auth"
	"go-services/api/response"
)

// MultiTenantTokenVerify 多租户JWT认证中间件
// 这是一个新增的中间件，用于支持多租户场景
// 原有的 TokenVerify 中间件保持不变，确保向后兼容
func MultiTenantTokenVerify(c *gin.Context) {
	c.FormFile("file") // 防止文件未发送完成就返回错误, 导致前端504而不是正确响应

	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		response.ReturnError(c, response.UNAUTHENTICATED, "未携带 token")
		c.Abort()
		return
	}

	var tokenString string
	// 支持两种格式: "Bearer token" 或直接 "token"
	if strings.HasPrefix(authHeader, "Bearer ") {
		tokenString = strings.TrimPrefix(authHeader, "Bearer ")
	} else {
		tokenString = authHeader
	}

	// 解析多租户JWT
	claims, err := auth.JWTDecrypt(tokenString)
	if err != nil {
		response.ReturnError(c, response.UNAUTHENTICATED, "token 解析失败")
		c.Abort()
		return
	}

	// 将租户和用户信息存入上下文
	c.Set("tenant_id", claims.TenantID)
	c.Set("user_id", claims.UserID)
	c.Set("account", claims.Account)

	c.Next()
}

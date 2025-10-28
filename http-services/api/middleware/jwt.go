package middleware

import (
	"github.com/gin-gonic/gin"

	"http-services/api/response"
	"http-services/utils/authentication"
)

// TokenVerify 获取 token 并验证其有效性
func TokenVerify(c *gin.Context) {
	token := c.Request.Header.Get("token")
	if token == "" {
		response.ReturnError(c, response.UNAUTHENTICATED, "without token.")
		return
	}
	tokenID, err := authentication.JWTDecrypt(token)
	if err != nil {
		response.ReturnError(c, response.UNAUTHENTICATED, "token verify failed.")
		return
	}
	// 将 JWT 数据设置到 gin.Context 中
	c.Set("jwtData", tokenID)
	c.Next()
}

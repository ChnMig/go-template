package middleware

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"http-services/api/response"
	"http-services/utils/authentication"
	"http-services/utils/contextkey"
	"http-services/utils/log"
)

const AuthorizationHeader = "Authorization"

// TokenVerify 获取 token 并验证其有效性
func TokenVerify(c *gin.Context) {
	token := c.Request.Header.Get(AuthorizationHeader)
	if token == "" {
		response.ReturnError(c, response.UNAUTHENTICATED, "without token.")
		return
	}
	jwtData, err := authentication.JWTDecrypt(token)
	if err != nil {
		log.WithRequest(c).Warn("token verification failed", zap.Error(err))
		response.ReturnError(c, response.UNAUTHENTICATED, "token verify failed.")
		return
	}
	// 将 JWT 数据设置到 gin.Context 中
	c.Set(contextkey.JWTData, jwtData)
	c.Next()
}

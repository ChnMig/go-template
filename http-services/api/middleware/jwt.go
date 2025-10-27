package middleware

import (
	"github.com/gin-gonic/gin"

	"go-services/api/response"
	"go-services/util/authentication"
)

// TokenVerify Get the token and verify its validity
func TokenVerify(c *gin.Context) {
	c.FormFile("file") // Prevents an error from being returned before the file is sent, resulting in a front-end 504 instead of a correct response
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
	// set data to gin.Context
	c.Set("jwtData", tokenID)
	// Next
	c.Next()
}

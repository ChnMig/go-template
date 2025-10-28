package example

import (
	"http-services/api/middleware"
	"http-services/api/response"
	"http-services/util/authentication"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func CreateToken(c *gin.Context) {
	params := &struct {
		User     string `json:"user" form:"user" binding:"required"`
		Password string `json:"password" form:"password" binding:"required"`
	}{}
	if !middleware.CheckParam(params, c) {
		return
	}
	if params.User != "admin" || params.Password != "pwd" {
		response.ReturnError(c, response.NOT_FOUND, "User name or password error")
		return
	}
	token, err := authentication.JWTIssue(params.User)
	if err != nil {
		response.ReturnError(c, response.INTERNAL, err.Error())
		zap.L().Error("JWTIssue", zap.Error(err))
		return
	}
	response.ReturnOk(c, token)
}

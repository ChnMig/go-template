package example

import (
	"http-services/api/middleware"
	"http-services/api/response"
	"http-services/utils/authentication"

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

	// 使用 map 存储用户数据，灵活适配不同项目
	userData := map[string]interface{}{
		"user_id":  "12345",      // 用户ID
		"username": params.User,  // 用户名
		"role":     "admin",      // 角色
		// 可以根据项目需要添加更多字段
		// "permissions": []string{"read", "write"},
		// "dept": "engineering",
	}

	token, err := authentication.JWTIssue(userData)
	if err != nil {
		response.ReturnError(c, response.INTERNAL, err.Error())
		zap.L().Error("JWTIssue", zap.Error(err))
		return
	}
	response.ReturnOk(c, token)
}

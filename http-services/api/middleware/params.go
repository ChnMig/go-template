package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"

	"http-services/api/response"
	"http-services/utils/log"
)

// CheckParam 通用参数校验与绑定辅助函数。
// 成功绑定后，会将绑定结果挂载到 gin.Context 中，便于日志等场景统一获取。
func CheckParam(params interface{}, c *gin.Context) bool {
	if err := c.ShouldBindWith(params, binding.Default(c.Request.Method, c.ContentType())); err != nil {
		response.ReturnError(c, response.INVALID_ARGUMENT, err.Error())
		return false
	}
	// 将本次绑定的参数挂到 context 上，供日志等场景使用。
	// 注意：这里只是弱类型存储，不参与业务逻辑判断。
	c.Set(log.BoundParamsKey, params)
	return true
}

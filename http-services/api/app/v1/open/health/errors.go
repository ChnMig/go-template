package health

import (
	"errors"

	"github.com/gin-gonic/gin"
	"http-services/api/response"
	domain "http-services/domain/health"
)

// 模块级业务错误码定义（仅在 health 模块内部使用）
const (
	// CodeHealthServiceNotReady 健康检查：服务尚未就绪
	CodeHealthServiceNotReady = 10001
	// CodeHealthServiceUnhealthy 健康检查：服务当前不可用
	CodeHealthServiceUnhealthy = 10002
)

// ReturnDomainError 将领域层健康检查错误映射为统一的接口错误响应
// 示例：根据 ErrServiceNotReady / ErrServiceUnhealthy 返回不同的错误码与提示文案
func ReturnDomainError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, domain.ErrServiceNotReady):
		// 服务尚未就绪，基础码使用 FAILED_PRECONDITION，业务码使用模块内自定义 code
		data := response.FAILED_PRECONDITION
		data.Code = CodeHealthServiceNotReady
		response.ReturnError(c, data, "服务尚未就绪，请稍后重试")
	case errors.Is(err, domain.ErrServiceUnhealthy):
		// 服务当前不可用，基础码使用 UNAVAILABLE，业务码使用模块内自定义 code
		data := response.UNAVAILABLE
		data.Code = CodeHealthServiceUnhealthy
		response.ReturnError(c, data, "服务当前不可用，请稍后重试")
	default:
		// 未分类错误，统一按 INTERNAL 处理
		response.ReturnError(c, response.INTERNAL, "服务内部错误")
	}
}

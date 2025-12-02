package health

import (
	"github.com/gin-gonic/gin"
	"http-services/api/response"
	domain "http-services/domain/health"
	"http-services/utils/log"
)

// Status 合并后的健康检查接口
// 返回服务健康与就绪状态，以及运行时长
func Status(c *gin.Context) {
	l := log.FromContext(c)
	l.Info("Health check requested")

	status, err := domain.GetStatus()
	if err != nil {
		// 示例：将领域错误映射为统一的接口错误响应
		ReturnDomainError(c, err)
		return
	}

	dto := StatusDTO{
		Status:    status.Status,
		Ready:     status.Ready,
		Uptime:    status.Uptime.String(),
		Timestamp: status.Timestamp,
	}
	response.ReturnOk(c, dto)
}

package health

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"http-services/api/response"
	domain "http-services/domain/health"
	"http-services/utils/log"
)

// Status 合并后的健康检查接口
// 返回服务健康与就绪状态，以及运行时长
func Status(c *gin.Context) {
	// 默认使用带 trace_id 等上下文信息的 logger
	l := log.FromContext(c)
	l.Debug("健康检查开始")

	status, err := domain.GetStatus()
	if err != nil {
		// 仅在出错时记录请求参数，便于排查问题
		log.WithRequest(c).Error("健康检查失败", zap.Error(err))
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

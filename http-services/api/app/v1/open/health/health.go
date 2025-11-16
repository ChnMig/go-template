package health

import (
	"time"

	"github.com/gin-gonic/gin"
	"http-services/api/response"
	"http-services/utils/log"
)

var startTime = time.Now()

// Status 合并后的健康检查接口
// 返回服务健康与就绪状态，以及运行时长
func Status(c *gin.Context) {
	l := log.FromContext(c)
	l.Info("Health check requested")
	dto := StatusDTO{
		Status:    "ok",
		Ready:     true,
		Uptime:    time.Since(startTime).String(),
		Timestamp: time.Now().Unix(),
	}
	response.ReturnOk(c, dto)
}

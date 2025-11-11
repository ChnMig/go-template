package health

import (
	"http-services/api/response"
	"http-services/utils/log"
	"time"

	"github.com/gin-gonic/gin"
)

var startTime = time.Now()

// Status 合并后的健康检查接口
// 返回服务健康与就绪状态，以及运行时长
func Status(c *gin.Context) {
	l := log.FromContext(c)
	l.Info("Health check requested")
	response.ReturnOk(c, gin.H{
		"status":    "ok",
		"uptime":    time.Since(startTime).String(),
		"timestamp": time.Now().Unix(),
	})
}

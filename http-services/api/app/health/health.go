package health

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

var startTime = time.Now()

// Health 健康检查端点
// 用于 k8s liveness probe 或负载均衡器检查
func Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
	})
}

// Ready 就绪检查端点
// 用于 k8s readiness probe，检查服务是否准备好接收流量
func Ready(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "ready",
		"uptime": time.Since(startTime).String(),
	})
}

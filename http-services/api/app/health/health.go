package health

import (
    "net/http"
    "time"

    "github.com/gin-gonic/gin"
)

var startTime = time.Now()

// Status 合并后的健康检查接口
// 返回服务健康与就绪状态，以及运行时长
func Status(c *gin.Context) {
    c.JSON(http.StatusOK, gin.H{
        "status": "ok",
        "ready":  true,
        "uptime": time.Since(startTime).String(),
    })
}

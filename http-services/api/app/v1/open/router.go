package open

import (
	"github.com/gin-gonic/gin"
	health "http-services/api/app/v1/open/health"
)

// RegisterRoutes 统一在 /api/v1/open 下注册各模块公开路由
func RegisterRoutes(open *gin.RouterGroup) {
	if open == nil {
		return
	}
	// 健康检查
	health.RegisterOpenRoutes(open)
}

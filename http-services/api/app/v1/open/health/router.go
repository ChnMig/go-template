package health

import "github.com/gin-gonic/gin"

// RegisterOpenRoutes 注册 health 模块的开放路由
// 路径：/api/v1/open/health
func RegisterOpenRoutes(open *gin.RouterGroup) {
	if open == nil {
		return
	}
	open.GET("/health", Status)
}

// RegisterPrivateRoutes 预留（health 模块通常无私有接口）
func RegisterPrivateRoutes(_ *gin.RouterGroup) {}

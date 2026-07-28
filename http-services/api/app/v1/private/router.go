package private

import "github.com/gin-gonic/gin"

// RegisterRoutes 统一在 /api/v1/private 下注册各模块私有路由
func RegisterRoutes(private *gin.RouterGroup) {
	if private == nil {
		return
	}
	// 预留：按需在此注册各模块私有接口
}

package app

import (
	"github.com/gin-gonic/gin"
	v1 "http-services/api/app/v1"
)

// RegisterRoutes 负责在 /api 下挂载各版本的路由
// 当前仅提供 /v1，后续新增版本时在此统一编排
func RegisterRoutes(api *gin.RouterGroup) {
	if api == nil {
		return
	}
	v1Group := api.Group("/v1")
	v1.RegisterRoutes(v1Group)
}

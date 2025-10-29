package v1

import (
	"github.com/gin-gonic/gin"
	"http-services/api/app/v1/open"
	"http-services/api/app/v1/private"
)

// RegisterRoutes 负责在 /api/v1 下挂载各子分组（open/private等）
func RegisterRoutes(v1 *gin.RouterGroup) {
	if v1 == nil {
		return
	}

	// /api/v1/open
	openGroup := v1.Group("/open")
	open.RegisterRoutes(openGroup)

	// /api/v1/private
	privateGroup := v1.Group("/private")
	private.RegisterRoutes(privateGroup)
}

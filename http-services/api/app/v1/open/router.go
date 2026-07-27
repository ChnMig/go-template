package open

import (
	"http-services/api/app/v1/open/health"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(group *gin.RouterGroup, modules ...func(*gin.RouterGroup)) {
	if group == nil {
		return
	}
	openGroup := group.Group("/open")
	health.RegisterOpenRoutes(openGroup)
	for _, register := range modules {
		if register != nil {
			register(openGroup)
		}
	}
}

package v1

import (
	"http-services/api/app/v1/open"
	"http-services/api/app/v1/private"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(group *gin.RouterGroup, openModules ...func(*gin.RouterGroup)) {
	if group == nil {
		return
	}
	version := group.Group("/v1")
	open.RegisterRoutes(version, openModules...)
	private.RegisterRoutes(version)
}

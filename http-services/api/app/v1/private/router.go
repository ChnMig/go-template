package private

import "github.com/gin-gonic/gin"

func RegisterRoutes(group *gin.RouterGroup) {
	if group == nil {
		return
	}
	_ = group.Group("/private")
}

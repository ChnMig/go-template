// Package app mounts versioned routes beneath /api.
package app

import (
	v1 "http-services/api/app/v1"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes mounts the default route tree.
func RegisterRoutes(group *gin.RouterGroup) {
	NewRegistrar()(group)
}

// NewRegistrar adds injected open modules to the fixed route hierarchy.
func NewRegistrar(openModules ...func(*gin.RouterGroup)) func(*gin.RouterGroup) {
	return func(group *gin.RouterGroup) {
		if group == nil {
			return
		}
		v1.RegisterRoutes(group, openModules...)
	}
}

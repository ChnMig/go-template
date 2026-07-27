package api

import (
	"net/http"
	"path"
	"strings"

	"github.com/gin-gonic/gin"
)

func registerStatic(router *gin.Engine, directory string) {
	if strings.TrimSpace(directory) == "" {
		return
	}
	filesystem := http.Dir(directory)
	handler := func(context *gin.Context) {
		requested := strings.TrimPrefix(context.Param("filepath"), "/")
		context.FileFromFS(path.Clean("/"+requested), filesystem)
	}
	router.GET("/static/*filepath", handler)
	router.HEAD("/static/*filepath", handler)
}

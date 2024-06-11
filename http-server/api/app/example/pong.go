package example

import (
	"http-server/api/response"

	"github.com/gin-gonic/gin"
)

func Pong(c *gin.Context) {
	response.ReturnSuccess(c)
}

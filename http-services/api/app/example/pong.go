package example

import (
	"go-services/api/response"

	"github.com/gin-gonic/gin"
)

func Pong(c *gin.Context) {
	response.ReturnSuccess(c)
}

package response

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func ReturnErrorWithData(c *gin.Context, data responseData, result interface{}) {
	data.Timestamp = time.Now().Unix()
	data.Detail = result
	c.JSON(http.StatusOK, data)
	// Return directly
	c.Abort()
}

// ResponseOk
func ReturnOk(c *gin.Context, result interface{}) {
	data := OK
	data.Timestamp = time.Now().Unix()
	data.Detail = result
	c.JSON(http.StatusOK, data)
	// Return directly
	c.Abort()
}

// ResponseOkWithTotal
func ReturnOkWithTotal(c *gin.Context, total int, result interface{}) {
	data := OK
	data.Timestamp = time.Now().Unix()
	data.Detail = result
	data.Total = &total
	c.JSON(http.StatusOK, data)
	// Return directly
	c.Abort()
}

// ReturnOkWithCount 已废弃，请使用 ReturnOkWithTotal
// Deprecated: Use ReturnOkWithTotal instead
func ReturnOkWithCount(c *gin.Context, count int, result interface{}) {
	ReturnOkWithTotal(c, count, result)
}

// ResponseError
func ReturnError(c *gin.Context, data responseData, message string) {
	data.Timestamp = time.Now().Unix()
	if message != "" {
		data.Message = message
	}
	c.JSON(http.StatusOK, data)
	// Return directly
	c.Abort()
}

// ResponseSuccess
func ReturnSuccess(c *gin.Context) {
	data := OK
	data.Timestamp = time.Now().Unix()
	c.JSON(http.StatusOK, data)
	// Return directly
	c.Abort()
}

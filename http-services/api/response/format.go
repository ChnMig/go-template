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

// ResponseOkWithCount
func ReturnOkWithCount(c *gin.Context, count int, result interface{}) {
	data := OK
	data.Timestamp = time.Now().Unix()
	data.Detail = result
	data.Count = &count
	c.JSON(http.StatusOK, data)
	// Return directly
	c.Abort()
}

// ResponseError
func ReturnError(c *gin.Context, data responseData, description string) {
	data.Timestamp = time.Now().Unix()
	data.Description = func() string {
		if description == "" {
			return data.Description
		}
		return description
	}()
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

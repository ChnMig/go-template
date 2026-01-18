package response

import (
	"net/http"
	"time"

	"http-services/utils/log"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func getTraceID(c *gin.Context) string {
	if traceID, exists := c.Get("trace_id"); exists {
		if id, ok := traceID.(string); ok {
			return id
		}
	}
	return ""
}

func ReturnErrorWithData(c *gin.Context, data responseData, result interface{}) {
	l := log.WithRequest(c)
	data.Timestamp = time.Now().Unix()
	data.TraceID = getTraceID(c)
	data.Detail = result
	c.JSON(http.StatusOK, data)
	l.Error("Returning error response with data", zap.Any("response", data))
	// Return directly
	c.Abort()
}

// ResponseOk
func ReturnOk(c *gin.Context, result interface{}) {
	l := log.WithRequest(c)
	data := OK
	data.Timestamp = time.Now().Unix()
	data.TraceID = getTraceID(c)
	data.Detail = result
	c.JSON(http.StatusOK, data)
	l.Debug("Returning OK response", zap.Any("response", data))
	// Return directly
	c.Abort()
}

// ResponseOkWithTotal
func ReturnOkWithTotal(c *gin.Context, total int, result interface{}) {
	l := log.WithRequest(c)
	data := OK
	data.Timestamp = time.Now().Unix()
	data.TraceID = getTraceID(c)
	data.Detail = result
	data.Total = &total
	c.JSON(http.StatusOK, data)
	l.Debug("Returning OK response with total", zap.Any("response", data))
	// Return directly
	c.Abort()
}

// ResponseError
func ReturnError(c *gin.Context, data responseData, message string) {
	l := log.WithRequest(c)
	data.Timestamp = time.Now().Unix()
	data.TraceID = getTraceID(c)
	if message != "" {
		data.Message = message
	}
	c.JSON(http.StatusOK, data)
	l.Error("Returning error response", zap.Any("response", data))
	// Return directly
	c.Abort()
}

// ResponseSuccess
func ReturnSuccess(c *gin.Context) {
	l := log.WithRequest(c)
	data := OK
	data.Timestamp = time.Now().Unix()
	data.TraceID = getTraceID(c)
	c.JSON(http.StatusOK, data)
	l.Debug("Returning success response", zap.Any("response", data))
	// Return directly
	c.Abort()
}

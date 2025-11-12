package response

import (
	"http-services/utils/log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// getTraceID 从 context 中获取 trace_id
func getTraceID(c *gin.Context) string {
	if traceID, exists := c.Get("X-Request-ID"); exists {
		if id, ok := traceID.(string); ok {
			return id
		}
	}
	return ""
}

func ReturnErrorWithData(c *gin.Context, data responseData, result interface{}) {
	l := log.FromContext(c)
	data.Timestamp = time.Now().Unix()
	data.TraceID = getTraceID(c)
	data.Detail = result
	c.JSON(http.StatusOK, data)
	l.Info("Returning error response with data")
	// Return directly
	c.Abort()
}

// ResponseOk
func ReturnOk(c *gin.Context, result interface{}) {
	l := log.FromContext(c)
	data := OK
	data.Timestamp = time.Now().Unix()
	data.TraceID = getTraceID(c)
	data.Detail = result
	c.JSON(http.StatusOK, data)
	l.Info("Returning OK response")
	// Return directly
	c.Abort()
}

// ResponseOkWithTotal
func ReturnOkWithTotal(c *gin.Context, total int, result interface{}) {
	l := log.FromContext(c)
	data := OK
	data.Timestamp = time.Now().Unix()
	data.TraceID = getTraceID(c)
	data.Detail = result
	data.Total = &total
	c.JSON(http.StatusOK, data)
	l.Info("Returning OK response with total")
	// Return directly
	c.Abort()
}

// ResponseError
func ReturnError(c *gin.Context, data responseData, message string) {
	l := log.FromContext(c)
	data.Timestamp = time.Now().Unix()
	data.TraceID = getTraceID(c)
	if message != "" {
		data.Message = message
	}
	c.JSON(http.StatusOK, data)
	l.Info("Returning error response")
	// Return directly
	c.Abort()
}

// ResponseSuccess
func ReturnSuccess(c *gin.Context) {
	l := log.FromContext(c)
	data := OK
	data.Timestamp = time.Now().Unix()
	data.TraceID = getTraceID(c)
	c.JSON(http.StatusOK, data)
	l.Info("Returning success response")
	// Return directly
	c.Abort()
}

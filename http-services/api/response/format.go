package response

import (
	"net/http"
	"time"

	"http-services/utils/contextkey"
	"http-services/utils/log"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func ReturnErrorWithData(c *gin.Context, data responseData, result interface{}) {
	l := log.WithRequest(c)
	data.Timestamp = time.Now().Unix()
	data.TraceID = requestTraceID(c)
	data.Detail = result
	c.JSON(http.StatusOK, data)
	logErrorResponse(l, "Returning error response with data", data)
	// Return directly
	c.Abort()
}

// ResponseOk
func ReturnOk(c *gin.Context, result interface{}) {
	l := log.WithRequest(c)
	data := OK
	data.Timestamp = time.Now().Unix()
	data.TraceID = requestTraceID(c)
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
	data.TraceID = requestTraceID(c)
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
	data.TraceID = requestTraceID(c)
	if message != "" {
		data.Message = message
	}
	c.JSON(http.StatusOK, data)
	logErrorResponse(l, "Returning error response", data)
	// Return directly
	c.Abort()
}

func logErrorResponse(l *zap.Logger, message string, data responseData) {
	if l == nil {
		l = zap.L()
	}
	field := zap.Any("response", data)
	switch {
	case data.Code == CANCELLED.Code:
		l.Debug(message, field)
	case data.Code >= INTERNAL.Code:
		l.Error(message, field)
	default:
		l.Warn(message, field)
	}
}

// ResponseSuccess
func ReturnSuccess(c *gin.Context) {
	l := log.WithRequest(c)
	data := OK
	data.Timestamp = time.Now().Unix()
	data.TraceID = requestTraceID(c)
	c.JSON(http.StatusOK, data)
	l.Debug("Returning success response", zap.Any("response", data))
	// Return directly
	c.Abort()
}

func requestTraceID(c *gin.Context) string {
	if c == nil {
		return ""
	}
	if c.Request != nil {
		if traceID, ok := log.TraceID(c.Request.Context()); ok {
			return traceID
		}
	}
	if traceID, exists := c.Get(contextkey.TraceID); exists {
		value, _ := traceID.(string)
		return value
	}
	return ""
}

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
	logger := log.WithRequest(c)
	data.Timestamp = time.Now().Unix()
	data.TraceID = requestTraceID(c)
	data.Detail = result
	c.JSON(http.StatusOK, data)
	logErrorResponse(logger, "Returning error response with data", data)
	c.Abort()
}

// ResponseOk
func ReturnOk(c *gin.Context, result interface{}) {
	logger := log.WithRequest(c)
	data := OK
	data.Timestamp = time.Now().Unix()
	data.TraceID = requestTraceID(c)
	data.Detail = result
	c.JSON(http.StatusOK, data)
	logger.Debug("Returning OK response", zap.Any("response", data))
	c.Abort()
}

// ResponseOkWithTotal
func ReturnOkWithTotal(c *gin.Context, total int, result interface{}) {
	logger := log.WithRequest(c)
	data := OK
	data.Timestamp = time.Now().Unix()
	data.TraceID = requestTraceID(c)
	data.Detail = result
	data.Total = &total
	c.JSON(http.StatusOK, data)
	logger.Debug("Returning OK response with total", zap.Any("response", data))
	c.Abort()
}

// ResponseError
func ReturnError(c *gin.Context, data responseData, message string) {
	logger := log.WithRequest(c)
	data.Timestamp = time.Now().Unix()
	data.TraceID = requestTraceID(c)
	if message != "" {
		data.Message = message
	}
	c.JSON(http.StatusOK, data)
	logErrorResponse(logger, "Returning error response", data)
	c.Abort()
}

// ResponseSuccess
func ReturnSuccess(c *gin.Context) {
	logger := log.WithRequest(c)
	data := OK
	data.Timestamp = time.Now().Unix()
	data.TraceID = requestTraceID(c)
	c.JSON(http.StatusOK, data)
	logger.Debug("Returning success response", zap.Any("response", data))
	c.Abort()
}

func logErrorResponse(logger *zap.Logger, message string, data responseData) {
	if logger == nil {
		logger = zap.L()
	}
	field := zap.Any("response", data)
	switch {
	case data.Code == CANCELLED.Code:
		logger.Debug(message, field)
	case data.Code >= INTERNAL.Code:
		logger.Error(message, field)
	default:
		logger.Warn(message, field)
	}
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

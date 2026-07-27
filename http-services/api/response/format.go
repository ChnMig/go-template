package response

import (
	"encoding/json"
	"net/http"
	"time"

	"http-services/utils/contextkey"
	"http-services/utils/log"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func ReturnErrorWithData(c *gin.Context, data responseData, result interface{}) {
	write(c, data, result, true)
}

// ResponseOk
func ReturnOk(c *gin.Context, result interface{}) {
	write(c, OK, result, false)
}

// ResponseOkWithTotal
func ReturnOkWithTotal(c *gin.Context, total int, result interface{}) {
	data := OK
	data.Total = &total
	write(c, data, result, false)
}

// ResponseError
func ReturnError(c *gin.Context, data responseData, message string) {
	if message != "" {
		data.Message = message
	}
	write(c, data, nil, true)
}

// ResponseSuccess
func ReturnSuccess(c *gin.Context) {
	write(c, OK, nil, false)
}

func write(c *gin.Context, data responseData, detail interface{}, failed bool) {
	data.Timestamp = time.Now().Unix()
	data.TraceID = requestTraceID(c)
	data.Detail = detail
	encoded, err := json.Marshal(data)
	if err != nil {
		data = INTERNAL
		data.Timestamp = time.Now().Unix()
		data.TraceID = requestTraceID(c)
		data.Message = "internal server error"
		encoded, _ = json.Marshal(data)
		failed = true
	}
	c.Data(http.StatusOK, "application/json; charset=utf-8", encoded)
	c.Abort()
	logger := log.FromContext(c)
	fields := []zap.Field{zap.Int("code", data.Code), zap.String("status", data.Status)}
	if failed {
		logger.Error("http.response", fields...)
		return
	}
	logger.Debug("http.response", fields...)
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

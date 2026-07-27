package middleware

import (
	"errors"
	"io"
	"net/http"
	"sync"

	"http-services/api/response"
	"http-services/config"
	"http-services/utils/contextkey"
	serviceLog "http-services/utils/log"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type bodyLimitReadCloser struct {
	body    io.ReadCloser
	context *gin.Context
	logger  LoggerProvider
	capture *contextkey.RequestBodyCapture
	once    sync.Once
}

func (reader *bodyLimitReadCloser) Read(buffer []byte) (int, error) {
	bytesRead, err := reader.body.Read(buffer)
	if bytesRead > 0 {
		reader.capture.Bytes = append(reader.capture.Bytes, buffer[:bytesRead]...)
	}
	var maxBytesError *http.MaxBytesError
	if !errors.As(err, &maxBytesError) {
		return bytesRead, err
	}

	reader.once.Do(func() {
		reader.context.Abort()
		if !reader.context.Writer.Written() {
			response.ReturnError(reader.context, response.INVALID_ARGUMENT, "request body too large")
			return
		}
		loggerFromContext(reader.context.Request.Context(), reader.logger).Error(
			"http.body_too_large_after_commit",
			zap.String(logKeyMethod, reader.context.Request.Method),
			zap.String(logKeyPath, reader.context.Request.URL.Path),
			zap.Int(logKeyStatus, reader.context.Writer.Status()),
		)
	})
	return bytesRead, err
}

func (reader *bodyLimitReadCloser) Close() error {
	return reader.body.Close()
}

// BodySizeLimit rejects known oversize requests and aborts streamed overflow at the read boundary.
// A handler that receives a body-read error must return so the aborted chain can unwind.
func BodySizeLimit(maxSize config.ByteSize) gin.HandlerFunc {
	return BodySizeLimitWithLogger(maxSize, serviceLog.GetGinErrorLogger)
}

// BodySizeLimitWithLogger uses a dynamic global error logger provider.
func BodySizeLimitWithLogger(maxSize config.ByteSize, loggerProvider LoggerProvider) gin.HandlerFunc {
	return func(context *gin.Context) {
		capture := &contextkey.RequestBodyCapture{}
		context.Set(contextkey.RequestBody, capture)
		if context.Request.ContentLength > int64(maxSize) {
			response.ReturnError(context, response.INVALID_ARGUMENT, "request body too large")
			return
		}

		context.Request.Body = &bodyLimitReadCloser{
			body:    http.MaxBytesReader(context.Writer, context.Request.Body, int64(maxSize)),
			context: context,
			logger:  loggerProvider,
			capture: capture,
		}
		context.Next()
	}
}

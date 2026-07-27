package middleware_test

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"http-services/api/middleware"
	"http-services/config"
	"http-services/utils/contextkey"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func Test_BodySizeLimit_captures_consumed_request_body_for_error_logging(t *testing.T) {
	logger := newTestLogger(t, &bytes.Buffer{})
	router := gin.New()
	router.Use(middleware.BodySizeLimitWithLogger(config.ByteSize(64), logger))
	router.POST("/probe", func(context *gin.Context) {
		_, err := io.ReadAll(context.Request.Body)
		require.NoError(t, err)
		capture, ok := context.Get(contextkey.RequestBody)
		require.True(t, ok)
		require.Equal(t, `{"name":"diagnostic-value"}`, string(capture.(*contextkey.RequestBodyCapture).Bytes))
		context.Status(http.StatusNoContent)
	})
	request := httptest.NewRequest(http.MethodPost, "/probe", strings.NewReader(`{"name":"diagnostic-value"}`))
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	require.Equal(t, http.StatusNoContent, recorder.Code)
}

package response

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"http-services/utils/contextkey"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func Test_ReturnError_logs_request_response_and_underlying_error(t *testing.T) {
	var output bytes.Buffer
	context, _ := gin.CreateTestContext(httptest.NewRecorder())
	logger := zap.New(zapcore.NewCore(
		zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()), zapcore.AddSync(&output), zap.DebugLevel,
	))
	context.Request = httptest.NewRequest(http.MethodPost, "/probe?order_no=ORDER-1", nil)
	context.Request.Header.Set("Authorization", "Bearer diagnostic-token")
	context.Set(contextkey.Logger, logger)
	context.Set(contextkey.RequestBody, &contextkey.RequestBodyCapture{Bytes: []byte(`{"amount":100}`)})
	requestError := context.Error(errors.New("database unavailable"))
	require.NotNil(t, requestError)

	ReturnErrorWithData(context, INTERNAL, map[string]string{"state": "failed"})

	for _, expected := range []string{
		"http.response", "order_no=ORDER-1", "Bearer diagnostic-token", `\"amount\":100`,
		"database unavailable", `"state":"failed"`,
	} {
		require.Contains(t, output.String(), expected)
	}
}

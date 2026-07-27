package response_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"http-services/api/response"
	"http-services/utils/contextkey"
	serviceLog "http-services/utils/log"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func Test_ReturnError_logs_complete_params_and_response(t *testing.T) {
	// Given
	var output bytes.Buffer
	logger := zap.New(zapcore.NewCore(
		zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()), zapcore.AddSync(&output), zap.DebugLevel,
	))
	context, _ := gin.CreateTestContext(httptest.NewRecorder())
	context.Request = httptest.NewRequest(http.MethodPost, "/api/v1/open/bridge/ticket", nil)
	context.Set(contextkey.Logger, logger)
	context.Set(serviceLog.BoundParamsKey, &struct {
		Token      string `json:"token"`
		ExternalNo string `json:"external_no"`
	}{Token: "jwt-full-value", ExternalNo: "external-1"})

	// When
	response.ReturnError(context, response.INTERNAL, "internal server error")

	// Then
	var entry map[string]any
	require.NoError(t, json.Unmarshal(output.Bytes(), &entry))
	require.Equal(t, "error", entry["level"])
	require.Equal(t, "Returning error response", entry["msg"])
	require.Equal(t, map[string]any{"token": "jwt-full-value", "external_no": "external-1"}, entry["params"])
	loggedResponse, ok := entry["response"].(map[string]any)
	require.True(t, ok)
	require.Equal(t, float64(http.StatusInternalServerError), loggedResponse["code"])
	require.Equal(t, "INTERNAL", loggedResponse["status"])
	require.Equal(t, "internal server error", loggedResponse["message"])
}

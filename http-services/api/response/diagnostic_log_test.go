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
	if err := json.Unmarshal(output.Bytes(), &entry); err != nil {
		t.Fatalf("decode log entry: %v", err)
	}
	if entry["level"] != "error" {
		t.Fatalf("level = %v, want error", entry["level"])
	}
	if entry["msg"] != "Returning error response" {
		t.Fatalf("msg = %v, want Returning error response", entry["msg"])
	}
	params, ok := entry["params"].(map[string]any)
	if !ok || params["token"] != "jwt-full-value" || params["external_no"] != "external-1" {
		t.Fatalf("params = %#v, want complete bound params", entry["params"])
	}
	loggedResponse, ok := entry["response"].(map[string]any)
	if !ok {
		t.Fatalf("response = %#v, want object", entry["response"])
	}
	if loggedResponse["code"] != float64(http.StatusInternalServerError) ||
		loggedResponse["status"] != "INTERNAL" ||
		loggedResponse["message"] != "internal server error" {
		t.Fatalf("response = %#v, want complete INTERNAL envelope", loggedResponse)
	}
}

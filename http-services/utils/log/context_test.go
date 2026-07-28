package log

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"http-services/utils/contextkey"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func TestWithRequestKeepsCompleteParsedRequestValues(t *testing.T) {
	var output bytes.Buffer
	logger := zap.New(zapcore.NewCore(
		zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()), zapcore.AddSync(&output), zap.DebugLevel,
	))
	context, _ := gin.CreateTestContext(httptest.NewRecorder())
	context.Request = httptest.NewRequest(http.MethodPost, "/orders/OH-1?token=query-token", nil)
	context.Request.PostForm = url.Values{"sign": {"form-signature"}}
	context.Params = gin.Params{{Key: "order_no", Value: "OH-1"}}
	context.Set(contextkey.Logger, logger)
	params := &struct {
		Token      string `json:"token"`
		ExternalNo string `json:"external_no"`
	}{Token: "request-token", ExternalNo: "external-1"}
	context.Set(BoundParamsKey, params)

	WithRequest(context).Error("operation failed")

	var entry map[string]any
	if err := json.Unmarshal(output.Bytes(), &entry); err != nil {
		t.Fatalf("decode log entry: %v", err)
	}
	if entry["query"] != "token=query-token" {
		t.Fatalf("query = %#v, want complete query", entry["query"])
	}
	loggedParams, ok := entry["params"].(map[string]any)
	if !ok || loggedParams["token"] != "request-token" || loggedParams["external_no"] != "external-1" {
		t.Fatalf("params = %#v, want complete bound params", entry["params"])
	}
	loggedForm, ok := entry["form"].(map[string]any)
	if !ok || len(loggedForm) != 1 {
		t.Fatalf("form = %#v, want parsed form", entry["form"])
	}
	pathParams, ok := entry["path_params"].(map[string]any)
	if !ok || pathParams["order_no"] != "OH-1" {
		t.Fatalf("path_params = %#v, want complete path params", entry["path_params"])
	}
}

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
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func Test_WithRequest_matches_template_request_fields(t *testing.T) {
	// Given
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

	// When
	WithRequest(context).Error("operation failed")

	// Then
	var entry map[string]any
	require.NoError(t, json.Unmarshal(output.Bytes(), &entry))
	require.Equal(t, http.MethodPost, entry["method"])
	require.Equal(t, "/orders/OH-1", entry["path"])
	require.Equal(t, "token=query-token", entry["query"])
	require.Equal(t, map[string]any{"sign": []any{"form-signature"}}, entry["form"])
	require.Equal(t, map[string]any{"order_no": "OH-1"}, entry["path_params"])
	require.Equal(t, map[string]any{"token": "request-token", "external_no": "external-1"}, entry["params"])
	require.NotContains(t, entry, "headers")
	require.NotContains(t, entry, "body")
}

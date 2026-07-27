package log

import (
	"bytes"
	"errors"
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

func Test_WithRequest_records_complete_request_inputs(t *testing.T) {
	var output bytes.Buffer
	logger := zap.New(zapcore.NewCore(
		zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()), zapcore.AddSync(&output), zap.DebugLevel,
	))
	ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
	request := httptest.NewRequest(http.MethodPost, "/orders/42?token=query-value", nil)
	request.Header.Set("Authorization", "Bearer header-value")
	request.PostForm = url.Values{"form_key": {"form-value"}}
	ctx.Request = request
	ctx.Params = gin.Params{{Key: "order_id", Value: "42"}}
	ctx.Set(contextkey.Logger, logger)
	ctx.Set(contextkey.BoundParams, map[string]any{"bound_key": "bound-value"})
	ctx.Set(contextkey.RequestBody, &contextkey.RequestBodyCapture{Bytes: []byte(`{"body_key":"body-value"}`)})

	WithRequest(ctx).Error("request failed", zap.Error(errors.New("database unavailable")))

	for _, expected := range []string{
		"token=query-value", "Bearer header-value", "form-value", "order_id", "42",
		"bound-value", `\"body_key\":\"body-value\"`, "database unavailable",
	} {
		require.Contains(t, output.String(), expected)
	}
}

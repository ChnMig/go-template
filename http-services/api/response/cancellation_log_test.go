package response

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"http-services/utils/contextkey"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func TestErrorResponseLogsRequestCancellation(t *testing.T) {
	gin.SetMode(gin.TestMode)
	canceled, cancel := context.WithCancel(context.Background())
	cancel()
	expired, cancelDeadline := context.WithDeadline(context.Background(), time.Now().Add(-time.Second))
	defer cancelDeadline()
	cases := []struct {
		name  string
		ctx   context.Context
		data  responseData
		level string
	}{
		{name: "missing request", data: INTERNAL, level: "error"},
		{name: "active request", ctx: context.Background(), data: INTERNAL, level: "error"},
		{name: "canceled request", ctx: canceled, data: INTERNAL, level: "debug"},
		{name: "deadline exceeded", ctx: expired, data: INTERNAL, level: "error"},
		{name: "business error", ctx: context.Background(), data: INVALID_ARGUMENT, level: "warn"},
		{name: "canceled business error", ctx: canceled, data: INVALID_ARGUMENT, level: "debug"},
		{name: "expired business error", ctx: expired, data: INVALID_ARGUMENT, level: "warn"},
		{name: "canceled response", ctx: context.Background(), data: CANCELLED, level: "debug"},
	}
	responders := []struct {
		name   string
		call   func(*gin.Context, responseData)
		detail bool
	}{
		{name: "ReturnError", call: func(c *gin.Context, data responseData) { ReturnError(c, data, "response message") }},
		{name: "ReturnErrorWithData", detail: true, call: func(c *gin.Context, data responseData) {
			ReturnErrorWithData(c, data, map[string]string{"reason": "test"})
		}},
	}
	for _, responder := range responders {
		t.Run(responder.name, func(t *testing.T) {
			for _, tc := range cases {
				t.Run(tc.name, func(t *testing.T) {
					var output bytes.Buffer
					logger := zap.New(zapcore.NewCore(zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()), zapcore.AddSync(&output), zapcore.DebugLevel))
					recorder := httptest.NewRecorder()
					c, _ := gin.CreateTestContext(recorder)
					if tc.ctx != nil {
						c.Request = httptest.NewRequestWithContext(tc.ctx, http.MethodGet, "/test", nil)
					}
					c.Set(contextkey.Logger, logger)
					c.Set(contextkey.TraceID, "response-trace")
					data := tc.data
					data.Message = "response message"
					responder.call(c, data)
					var entry struct {
						Level string `json:"level"`
					}
					if err := json.Unmarshal(bytes.TrimSpace(output.Bytes()), &entry); err != nil {
						t.Fatalf("expected one response log: %v; logs: %s", err, output.String())
					}
					if entry.Level != tc.level {
						t.Errorf("expected %s log, got %s", tc.level, entry.Level)
					}
					if recorder.Code != http.StatusOK || !c.IsAborted() {
						t.Errorf("expected HTTP 200 and aborted context, got status=%d aborted=%t", recorder.Code, c.IsAborted())
					}
					var body responseData
					if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
						t.Fatalf("invalid response JSON: %v", err)
					}
					if body.Code != data.Code || body.Status != data.Status || body.Description != data.Description || body.Message != data.Message {
						t.Errorf("response error fields changed: %#v", body)
					}
					if body.TraceID != "response-trace" || body.Timestamp == 0 {
						t.Errorf("missing response metadata: %#v", body)
					}
					if responder.detail {
						detail, ok := body.Detail.(map[string]interface{})
						if !ok || detail["reason"] != "test" {
							t.Errorf("response detail changed: %#v", body.Detail)
						}
					} else if body.Detail != nil {
						t.Errorf("unexpected error detail: %#v", body.Detail)
					}
				})
			}
		})
	}
}

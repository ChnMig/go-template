package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"http-services/api/response"
	"http-services/utils/contextkey"
)

func TestBindingFailureLogsOriginalRequest(t *testing.T) {
	for _, tc := range []struct {
		name string
		bind func(*gin.Context)
	}{
		{"check helper", func(c *gin.Context) {
			CheckJSONParam(&struct {
				DockingType int `json:"docking_type"`
			}{}, c)
		}},
		{"bind helper after query", func(c *gin.Context) {
			if err := BindQueryParam(&struct {
				ID int `form:"id"`
			}{}, c); err != nil {
				t.Fatal(err)
			}
			if err := BindParam(&struct {
				DockingType int `json:"docking_type"`
			}{}, c); err == nil {
				t.Fatal("array should fail integer binding")
			}
			response.ReturnError(c, response.INVALID_ARGUMENT, "invalid type")
		}},
		{"Gin direct binding", func(c *gin.Context) {
			if err := c.ShouldBindJSON(&struct {
				DockingType int `json:"docking_type"`
			}{}); err == nil {
				t.Fatal("array should fail integer binding")
			}
			response.ReturnError(c, response.INVALID_ARGUMENT, "invalid type")
		}},
		{"validation failure", func(c *gin.Context) {
			CheckJSONParam(&struct {
				Name string `json:"name" binding:"required"`
			}{}, c)
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			logger, logs := requestParamsTestLogger()
			undo := zap.ReplaceGlobals(logger)
			t.Cleanup(undo)
			r := gin.New()
			r.Use(TraceID())
			r.PUT("/scope", tc.bind)
			body := `{"customer_id":9,"docking_type":[1,2],"password":"binding-secret"}`
			req := httptest.NewRequest(http.MethodPut, "/scope?id=9", strings.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set(TraceIDHeaderKey, "018f47a5-7b8c-7c11-8000-123456789abc")
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			entries := filterRequestLogs(t, logs, "msg", "Returning error response")
			if len(entries) != 1 {
				t.Fatalf("missing unified response log: %+v", logs.String())
			}
			fields := entries[0]
			if fields["trace_id"] != "018f47a5-7b8c-7c11-8000-123456789abc" || fields["method"] != "PUT" || fields["path"] != "/scope" || fields["params"] != nil {
				t.Fatalf("missing context or stale bound parameters: %+v", fields)
			}
			payload, ok := fields["request_body"].(map[string]any)
			if !ok || len(payload["docking_type"].([]any)) != 2 || payload["password"] != "binding-secret" {
				t.Fatalf("original request missing or changed: %+v", fields)
			}
			for _, entry := range filterRequestLogs(t, logs, "msg", "invalid request parameters") {
				if entry["level"] != "warn" || entry["trace_id"] != "018f47a5-7b8c-7c11-8000-123456789abc" || entry["request_body"] == nil {
					t.Fatalf("binding error lost request context: %+v", entry)
				}
			}
			if strings.Contains(w.Body.String(), "binding-secret") || strings.Contains(w.Body.String(), "request_body") {
				t.Fatal("request logging must not change the public error response")
			}
		})
	}
}

func TestBindingLogsRequestWithoutTraceMiddleware(t *testing.T) {
	logger, logs := requestParamsTestLogger()
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Set(contextkey.Logger, logger)
	c.Request = httptest.NewRequest(http.MethodPost, "/scope", strings.NewReader(`{"id":[]}`))
	c.Request.Header.Set("Content-Type", "application/json")
	if CheckJSONParam(&struct {
		ID int `json:"id"`
	}{}, c) {
		t.Fatal("invalid request accepted")
	}
	entry := filterRequestLogs(t, logs, "msg", "Returning error response")[0]
	if entry["request_body"] == nil || entry["path"] != "/scope" {
		t.Fatalf("binding helper did not capture request: %+v", entry)
	}
}

func requestParamsTestLogger() (*zap.Logger, *bytes.Buffer) {
	var output bytes.Buffer
	core := zapcore.NewCore(zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()), zapcore.AddSync(&output), zap.DebugLevel)
	return zap.New(core), &output
}

func filterRequestLogs(t *testing.T, output *bytes.Buffer, key, value string) []map[string]any {
	t.Helper()
	decoder := json.NewDecoder(bytes.NewReader(output.Bytes()))
	var entries []map[string]any
	for {
		var entry map[string]any
		if err := decoder.Decode(&entry); err == io.EOF {
			break
		} else if err != nil {
			t.Fatalf("invalid log: %v: %s", err, output.String())
		}
		if entry[key] == value {
			entries = append(entries, entry)
		}
	}
	return entries
}
func TestTokenVerifyInvalidTokenLogRetainsTrace(t *testing.T) {
	logger, logs := requestParamsTestLogger()
	undo := zap.ReplaceGlobals(logger)
	t.Cleanup(undo)
	r := gin.New()
	r.Use(TraceID(), TokenVerify)
	r.GET("/test", func(c *gin.Context) { t.Fatal("invalid token accepted") })
	req := httptest.NewRequest("GET", "/test?id=9", nil)
	req.Header.Set(AuthorizationHeader, "invalid.token.here")
	req.Header.Set(TraceIDHeaderKey, "018f47a5-7b8c-7c11-8000-123456789abc")
	r.ServeHTTP(httptest.NewRecorder(), req)
	entries := filterRequestLogs(t, logs, "msg", "token verification failed")
	if len(entries) != 1 || entries[0]["trace_id"] != "018f47a5-7b8c-7c11-8000-123456789abc" || entries[0]["query"] != "id=9" {
		t.Fatalf("JWT diagnostic lost request context: %s", logs.String())
	}
	if entries := filterRequestLogs(t, logs, "msg", "operation failed"); len(entries) != 0 {
		t.Fatal("JWT parser emitted a duplicate diagnostic without request context")
	}
}

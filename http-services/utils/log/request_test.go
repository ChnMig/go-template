package log

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"http-services/utils/contextkey"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func requestLogContext(t *testing.T, method, target, body string) (*gin.Context, *bytes.Buffer) {
	t.Helper()
	oldMode := gin.Mode()
	t.Cleanup(func() { gin.SetMode(oldMode) })
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(method, target, strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	var output bytes.Buffer
	core := zapcore.NewCore(zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()), zapcore.AddSync(&output), zap.DebugLevel)
	c.Set(contextkey.Logger, zap.New(core))
	c.Set(contextkey.TraceID, "request-log-test")
	return c, &output
}

func requestLogEntry(t *testing.T, output *bytes.Buffer) map[string]any {
	t.Helper()
	var entry map[string]any
	if err := json.Unmarshal(output.Bytes(), &entry); err != nil {
		t.Fatalf("decode log: %v: %s", err, output.String())
	}
	return entry
}

func TestWithRequestPreservesOriginalJSONAndFieldValues(t *testing.T) {
	body := `{"customer_id":9007199254740993,"scopes":[{"docking_type":[1,2]}],"Password":"body-secret","nested":{"access_token":"token-secret"}}`
	c, output := requestLogContext(t, http.MethodPut, "/scope?id=9&token=query-secret", body)
	c.Request.Header.Set("Authorization", "Bearer header-secret")
	CaptureRequestBody(c)
	got, err := io.ReadAll(c.Request.Body)
	if err != nil || string(got) != body {
		t.Fatalf("capture changed payload: %s, %v", got, err)
	}
	WithRequest(c).Warn("binding failed")
	entry := requestLogEntry(t, output)
	request := entry["request_body"].(map[string]any)
	if request["Password"] != "body-secret" || request["nested"].(map[string]any)["access_token"] != "token-secret" {
		t.Fatalf("request values were changed: %s", output.String())
	}
	scopes := request["scopes"].([]any)
	if values := scopes[0].(map[string]any)["docking_type"].([]any); len(values) != 2 || values[0] != float64(1) || values[1] != float64(2) {
		t.Fatalf("lost original invalid field type: %s", output.String())
	}
	if !strings.Contains(output.String(), `"customer_id":9007199254740993`) {
		t.Fatalf("large integer lost precision: %s", output.String())
	}
	for _, key := range []string{"trace_id", "method", "path", "client_ip"} {
		if entry[key] == nil || bytes.Count(output.Bytes(), []byte(`"`+key+`":`)) != 1 {
			t.Fatalf("missing or duplicate %s: %s", key, output.String())
		}
	}
	for _, secret := range []string{"body-secret", "token-secret", "query-secret"} {
		if !strings.Contains(output.String(), secret) {
			t.Fatalf("request log omitted %s", secret)
		}
	}
}

func TestWithRequestPreservesBoundAndDeleteFormWithoutMutation(t *testing.T) {
	c, output := requestLogContext(t, http.MethodDelete, "/scope/9", "")
	bound := &struct {
		ID       int    `json:"id"`
		Password string `json:"password"`
	}{9, "bound-secret"}
	c.Set(BoundParamsKey, bound)
	c.Request.PostForm = url.Values{"id": {"9"}, "client_secret": {"form-secret"}}
	c.Params = gin.Params{{Key: "id", Value: "9"}, {Key: "token", Value: "path-secret"}}
	WithRequest(c).Warn("failed")
	entry := requestLogEntry(t, output)
	if entry["params"].(map[string]any)["id"] != float64(9) || entry["form"] == nil || entry["path_params"] == nil {
		t.Fatalf("request fields missing: %s", output.String())
	}
	for _, value := range []string{"bound-secret", "form-secret", "path-secret"} {
		if !strings.Contains(output.String(), value) {
			t.Fatalf("request value missing: %s", output.String())
		}
	}
	if bound.Password != "bound-secret" || c.Request.PostForm.Get("client_secret") != "form-secret" || c.Param("token") != "path-secret" {
		t.Fatal("logging mutated business parameters")
	}
}

func TestRequestBodyCaptureIsPassiveAndPreservesReadErrorsAndClose(t *testing.T) {
	c, output := requestLogContext(t, http.MethodPost, "/scope", "")
	wantErr := errors.New("read interrupted")
	source := &requestLogTestReader{err: wantErr}
	c.Request.Body = source
	c.Request.ContentLength = -1
	CaptureRequestBody(c)
	wrapper := c.Request.Body
	CaptureRequestBody(c)
	WithRequest(c).Warn("auth rejected before reading")
	if source.reads != 0 || source.closed || c.Request.Body != wrapper {
		t.Fatal("logging consumed, closed or wrapped the request twice")
	}
	if entry := requestLogEntry(t, output); entry["request_body_status"] != "unread" {
		t.Fatalf("unread body was not identified: %v", entry)
	}
	data, err := io.ReadAll(c.Request.Body)
	if !errors.Is(err, wantErr) || string(data) != `{"id":9}` {
		t.Fatalf("body read semantics changed: %s, %v", data, err)
	}
	if err := c.Request.Body.Close(); err != nil || !source.closed {
		t.Fatal("Close was not forwarded")
	}
}

type requestLogTestReader struct {
	err    error
	reads  int
	closed bool
}

func (r *requestLogTestReader) Read(p []byte) (int, error) {
	r.reads++
	return copy(p, `{"id":9}`), r.err
}

func (r *requestLogTestReader) Close() error {
	r.closed = true
	return nil
}

func TestRequestBodyCaptureBoundsAndMalformedPayloads(t *testing.T) {
	for _, tc := range []struct{ name, body, status string }{
		{"oversized", `{"password":"` + strings.Repeat("secret", requestLogMaxBytes) + `"}`, "truncated"},
		{"malformed", `{"password":"malformed-secret`, "invalid_json"},
		{"binary", "\x00\x01binary-secret", "unsupported_content_type"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c, output := requestLogContext(t, http.MethodPost, "/scope", tc.body)
			if tc.name == "binary" {
				c.Request.Header.Set("Content-Type", "application/octet-stream")
			}
			CaptureRequestBody(c)
			data, err := io.ReadAll(c.Request.Body)
			if err != nil || string(data) != tc.body {
				t.Fatal("logging truncated or rejected the actual request")
			}
			capture, _ := c.Get(requestBodyKey)
			if len(capture.(*requestBodyCapture).data) > requestLogMaxBytes {
				t.Fatal("request log memory is not bounded")
			}
			WithRequest(c).Warn("failed")
			entry := requestLogEntry(t, output)
			if entry["request_body_status"] != tc.status {
				t.Fatalf("missing body diagnostic: %s", output.String())
			}
			if tc.name != "binary" && entry["request_body"] != tc.body[:min(len(tc.body), requestLogMaxBytes)] {
				t.Fatal("malformed or truncated text was not preserved")
			}
			if tc.name == "oversized" && entry["request_body_truncated"] != true {
				t.Fatal("missing truncation marker")
			}
		})
	}
}

func TestRequestBodyCapturePreservesHTTPBodyLimit(t *testing.T) {
	c, output := requestLogContext(t, http.MethodPost, "/scope", `{"id":123456789}`)
	CaptureRequestBody(c)
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 8)
	_, err := io.ReadAll(c.Request.Body)
	var limitErr *http.MaxBytesError
	if !errors.As(err, &limitErr) {
		t.Fatalf("body limit bypassed: %v", err)
	}
	WithRequest(c).Warn("body too large")
	if requestLogEntry(t, output)["request_body_incomplete"] != true {
		t.Fatalf("partial body must be identified: %s", output.String())
	}
}

func TestRequestBodyCaptureSurvivesReplayAndFormFailure(t *testing.T) {
	body := "customer_id=9&docking_type=bad&password=form-secret"
	c, output := requestLogContext(t, http.MethodPost, "/scope", body)
	c.Request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	CaptureRequestBody(c)
	data, err := io.ReadAll(c.Request.Body)
	if err != nil {
		t.Fatal(err)
	}
	c.Request.Body = io.NopCloser(bytes.NewReader(data))
	CaptureRequestBody(c)
	var params struct {
		DockingType int `form:"docking_type"`
	}
	if err := c.ShouldBind(&params); err == nil {
		t.Fatal("invalid form accepted")
	}
	WithRequest(c).Warn("invalid form")
	entry := requestLogEntry(t, output)
	if entry["request_body_bytes"] != float64(len(body)) || entry["request_body"] == nil || entry["form"] == nil {
		t.Fatalf("replay lost or double-counted the request: %s", output.String())
	}
	if !strings.Contains(output.String(), "form-secret") {
		t.Fatal("form value omitted")
	}
}

func TestFromContextNilIsSafe(t *testing.T) {
	if FromContext(nil) == nil || WithRequest(nil) == nil {
		t.Fatal("nil context should fall back to global logger")
	}
}

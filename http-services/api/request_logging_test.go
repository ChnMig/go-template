package api

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"http-services/api/middleware"
	"http-services/api/response"
	"http-services/config"
	httplog "http-services/utils/log"
)

func TestInitAPIRateLimitRetainsRequestTrace(t *testing.T) {
	httplog.GetGinLogger()
	oldWriter, oldErrorWriter, oldMode := gin.DefaultWriter, gin.DefaultErrorWriter, gin.Mode()
	oldMaxBodySize := config.MaxBodySize
	t.Cleanup(func() {
		gin.DefaultWriter, gin.DefaultErrorWriter = oldWriter, oldErrorWriter
		gin.SetMode(oldMode)
		config.MaxBodySize = oldMaxBodySize
	})
	config.MaxBodySize = 1 << 20
	oldEnabled, oldRate, oldBurst := config.EnableRateLimit, config.GlobalRateLimit, config.GlobalRateBurst
	t.Cleanup(func() {
		config.EnableRateLimit, config.GlobalRateLimit, config.GlobalRateBurst = oldEnabled, oldRate, oldBurst
		middleware.CleanupAllLimiters()
	})
	config.EnableRateLimit, config.GlobalRateLimit, config.GlobalRateBurst = true, 0, 1
	var logs bytes.Buffer
	core := zapcore.NewCore(zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()), zapcore.AddSync(&logs), zap.DebugLevel)
	undo := zap.ReplaceGlobals(zap.New(core))
	t.Cleanup(undo)
	r := InitApi()
	r.GET("/request-log-test", func(c *gin.Context) { c.Status(http.StatusNoContent) })
	for i := 0; i < 2; i++ {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/request-log-test?id=9", nil)
		req.RemoteAddr = "192.0.2.123:4567"
		req.Header.Set(middleware.TraceIDHeaderKey, "018f47a5-7b8c-7c11-8000-123456789abc")
		r.ServeHTTP(w, req)
		if w.Header().Get(middleware.TraceIDHeaderKey) != "018f47a5-7b8c-7c11-8000-123456789abc" {
			t.Fatal("early response lost trace header")
		}
		if i == 1 {
			var result struct {
				Code    int    `json:"code"`
				TraceID string `json:"trace_id"`
			}
			if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil || result.Code != response.RESOURCE_EXHAUSTED.Code || result.TraceID != "018f47a5-7b8c-7c11-8000-123456789abc" {
				t.Fatalf("rate limit response lost trace: %s, %v", w.Body.String(), err)
			}
		}
	}
	decoder := json.NewDecoder(bytes.NewReader(logs.Bytes()))
	found := false
	for {
		var entry map[string]any
		if err := decoder.Decode(&entry); err == io.EOF {
			break
		} else if err != nil {
			t.Fatal(err)
		}
		if entry["msg"] == "Returning error response" {
			found = true
			if entry["trace_id"] != "018f47a5-7b8c-7c11-8000-123456789abc" || entry["path"] != "/request-log-test" || entry["query"] != "id=9" {
				t.Fatalf("rate limit log lost request context: %+v", entry)
			}
		}
	}
	if !found {
		t.Fatal("missing rate limit diagnostic")
	}
}

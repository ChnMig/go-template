package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"http-services/api/response"
	"http-services/utils/contextkey"

	"github.com/gin-gonic/gin"
)

func TestRecoveryReturnsUnifiedInternalResponse(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(TraceID(), AccessLog(), Recovery())
	router.GET("/panic", func(c *gin.Context) {
		panic("boom")
	})

	req := httptest.NewRequest(http.MethodGet, "/panic", nil)
	req.Header.Set(contextkey.TraceIDHeader, "trace-123")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("HTTP status = %d, want %d", w.Code, http.StatusOK)
	}

	var body struct {
		Code    int    `json:"code"`
		Status  string `json:"status"`
		Message string `json:"message"`
		TraceID string `json:"trace_id"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("响应不是合法 JSON: %v", err)
	}
	if body.Code != response.INTERNAL.Code || body.Status != response.INTERNAL.Status {
		t.Fatalf("响应状态 = %d/%s, want %d/%s", body.Code, body.Status, response.INTERNAL.Code, response.INTERNAL.Status)
	}
	if body.TraceID != "trace-123" {
		t.Fatalf("trace_id = %q, want trace-123", body.TraceID)
	}
	if body.Message != "服务内部错误" {
		t.Fatalf("message = %q, want 服务内部错误", body.Message)
	}
}

func TestRecoveryWritesResponseBeforeOuterAccessLogDefer(t *testing.T) {
	gin.SetMode(gin.TestMode)

	responseSizeInOuterDefer := -1
	router := gin.New()
	router.Use(TraceID())
	router.Use(func(c *gin.Context) {
		defer func() {
			responseSizeInOuterDefer = c.Writer.Size()
		}()
		c.Next()
	})
	router.Use(Recovery())
	router.GET("/panic", func(c *gin.Context) {
		panic("boom")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/panic", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("HTTP status = %d, want %d", w.Code, http.StatusOK)
	}
	if responseSizeInOuterDefer <= 0 {
		t.Fatalf("outer access-log-style defer saw response size %d, want a written recovery response", responseSizeInOuterDefer)
	}
}

func TestAccessLogDoesNotBlockRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(TraceID(), AccessLog())
	router.GET("/ok", func(c *gin.Context) {
		response.ReturnSuccess(c)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/ok?foo=bar", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("HTTP status = %d, want %d", w.Code, http.StatusOK)
	}
	if w.Header().Get(contextkey.TraceIDHeader) == "" {
		t.Fatalf("未写入 %s 响应头", contextkey.TraceIDHeader)
	}
}

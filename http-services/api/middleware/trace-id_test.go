package middleware

import (
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestTraceIDMiddleware_GenerateWhenMissing(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(TraceID())
	router.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	traceID := w.Header().Get(TraceIDHeaderKey)
	if traceID == "" {
		t.Fatal("expected X-Trace-ID header to be set")
	}

	matched, err := regexp.MatchString(`^[a-f0-9]{32}$`, traceID)
	if err != nil {
		t.Fatalf("Regex error: %v", err)
	}
	if !matched {
		t.Fatalf("expected X-Trace-ID to be 32 char lower hex, got: %s", traceID)
	}
}

func TestTraceIDMiddleware_KeepWhenProvided(t *testing.T) {
	gin.SetMode(gin.TestMode)

	const expectedTraceID = "provided-trace-id"

	router := gin.New()
	router.Use(TraceID())
	router.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(TraceIDHeaderKey, expectedTraceID)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	traceID := w.Header().Get(TraceIDHeaderKey)
	if traceID != expectedTraceID {
		t.Fatalf("expected X-Trace-ID %q, got %q", expectedTraceID, traceID)
	}
}

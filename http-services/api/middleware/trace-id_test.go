package middleware

import (
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	"http-services/utils/contextkey"

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

func TestTraceIDMiddleware_KeepValidProvidedValue(t *testing.T) {
	gin.SetMode(gin.TestMode)

	const expectedTraceID = "0123456789abcdef0123456789abcdef"

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

func TestTraceIDMiddleware_ReplacesInvalidProvidedValue(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(TraceID())
	router.GET("/", func(c *gin.Context) {
		traceID, ok := contextkey.TraceIDFromContext(c.Request.Context())
		if !ok {
			t.Fatal("标准 context 中缺少 trace ID")
		}
		if traceID != c.GetString(TraceIDContextKey) {
			t.Fatalf("标准 context trace ID = %q, Gin context = %q", traceID, c.GetString(TraceIDContextKey))
		}
		c.String(http.StatusOK, "ok")
	})

	for _, invalid := range []string{"provided-trace-id", "0123456789ABCDEF0123456789ABCDEF", "0123456789abcdef0123456789abcdeg"} {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set(TraceIDHeaderKey, invalid)
		router.ServeHTTP(w, req)

		got := w.Header().Get(TraceIDHeaderKey)
		if got == invalid || !regexp.MustCompile(`^[a-f0-9]{32}$`).MatchString(got) {
			t.Fatalf("非法 trace ID %q 未被替换，got %q", invalid, got)
		}
	}
}

func TestTraceIDWithDependenciesFallsBackWhenFactoryReturnsInvalidValue(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(TraceIDWithDependencies(func() string { return "invalid" }, nil))
	router.GET("/", func(c *gin.Context) { c.Status(http.StatusNoContent) })
	w := httptest.NewRecorder()
	router.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/", nil))

	traceID := w.Header().Get(TraceIDHeaderKey)
	if !regexp.MustCompile(`^[a-f0-9]{32}$`).MatchString(traceID) {
		t.Fatalf("fallback trace ID = %q", traceID)
	}
}

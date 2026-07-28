package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func TestTraceIDMiddleware_GenerateUUIDv7WhenMissing(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(TraceID())
	router.GET("/", func(c *gin.Context) { c.Status(http.StatusNoContent) })
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))

	if recorder.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusNoContent)
	}
	assertUUIDv7(t, recorder.Header().Get(TraceIDHeaderKey))
}

func TestTraceIDMiddleware_KeepCanonicalUUIDWhenProvided(t *testing.T) {
	gin.SetMode(gin.TestMode)
	const expectedTraceID = "018f47a5-7b8c-7c11-8000-123456789abc"
	router := gin.New()
	router.Use(TraceID())
	router.GET("/", func(c *gin.Context) { c.Status(http.StatusNoContent) })
	request := httptest.NewRequest(http.MethodGet, "/", nil)
	request.Header.Set(TraceIDHeaderKey, expectedTraceID)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if traceID := recorder.Header().Get(TraceIDHeaderKey); traceID != expectedTraceID {
		t.Fatalf("X-Trace-ID = %q, want %q", traceID, expectedTraceID)
	}
}

func TestTraceIDMiddleware_ReplaceMalformedHeaderWithUUIDv7(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(TraceID())
	router.GET("/", func(c *gin.Context) { c.Status(http.StatusNoContent) })
	request := httptest.NewRequest(http.MethodGet, "/", nil)
	request.Header.Set(TraceIDHeaderKey, "provided-trace-id")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	assertUUIDv7(t, recorder.Header().Get(TraceIDHeaderKey))
}

func assertUUIDv7(t *testing.T, value string) {
	t.Helper()
	if len(value) != 36 {
		t.Fatalf("UUID length = %d, want 36: %q", len(value), value)
	}
	parsed, err := uuid.Parse(value)
	if err != nil {
		t.Fatalf("parse UUID: %v", err)
	}
	if parsed.Version() != uuid.Version(7) {
		t.Fatalf("UUID version = %v, want 7", parsed.Version())
	}
	if parsed.Variant() != uuid.RFC4122 {
		t.Fatalf("UUID variant = %v, want RFC4122", parsed.Variant())
	}
}

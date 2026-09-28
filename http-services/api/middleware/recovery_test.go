package middleware

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"syscall"
	"testing"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func TestRecoveryLogsRequestAndPreservesResponseSemantics(t *testing.T) {
	for _, tc := range []struct {
		name    string
		panic   any
		status  int
		message string
	}{
		{"panic", "test panic", http.StatusInternalServerError, "request panic recovered"},
		{"broken pipe", fmt.Errorf("write failed: %w", syscall.EPIPE), http.StatusOK, "request connection aborted"},
		{"connection reset", syscall.ECONNRESET, http.StatusOK, "request connection aborted"},
		{"abort handler", http.ErrAbortHandler, http.StatusOK, "request connection aborted"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			logger, logs := requestParamsTestLogger()
			undo := zap.ReplaceGlobals(logger)
			t.Cleanup(undo)
			r := gin.New()
			r.Use(Recovery(), TraceID())
			r.POST("/panic", func(c *gin.Context) {
				if _, err := io.ReadAll(c.Request.Body); err != nil {
					t.Fatal(err)
				}
				panic(tc.panic)
			})
			req := httptest.NewRequest(http.MethodPost, "/panic", strings.NewReader(`{"id":9,"password":"panic-secret"}`))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer header-secret")
			req.Header.Set("Cookie", "session=cookie-secret")
			req.Header.Set(TraceIDHeaderKey, "018f47a5-7b8c-7c11-8000-123456789abc")
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			if w.Code != tc.status || w.Body.Len() != 0 {
				t.Fatalf("changed recovery response: %d %s", w.Code, w.Body.String())
			}
			entries := filterRequestLogs(t, logs, "msg", tc.message)
			if len(entries) != 1 || entries[0]["trace_id"] != "018f47a5-7b8c-7c11-8000-123456789abc" || entries[0]["request_body"] == nil {
				t.Fatalf("panic lost request context: %s", logs.String())
			}
			if tc.status == http.StatusInternalServerError && (entries[0]["stack"] == nil || entries[0]["level"] != "error") {
				t.Fatal("panic lost stack or severity")
			}
			if tc.status != http.StatusInternalServerError && (entries[0]["stack"] != nil || entries[0]["level"] != "debug") {
				t.Fatal("disconnected client should not create a server panic report")
			}
			if !strings.Contains(logs.String(), "panic-secret") {
				t.Fatalf("recovery omitted request data: %s", logs.String())
			}
		})
	}
}

package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestCorsDomainHandlerAbortsPreflight(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handled := false
	router := gin.New()
	router.Use(CorsDomainHandler())
	router.OPTIONS("/resource", func(c *gin.Context) {
		handled = true
		c.Status(http.StatusTeapot)
	})

	req := httptest.NewRequest(http.MethodOptions, "/resource", nil)
	req.Header.Set("Origin", "https://example.com")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if handled {
		t.Fatal("preflight 请求进入了业务 handler")
	}
	if w.Code != http.StatusNoContent || w.Body.Len() != 0 {
		t.Fatalf("preflight response = %d %q, want 204 empty body", w.Code, w.Body.String())
	}
	if got := w.Header().Get("Access-Control-Allow-Headers"); got != "Authorization, Content-Type, X-Trace-ID" {
		t.Fatalf("Access-Control-Allow-Headers = %q", got)
	}
	if got := w.Header().Get("Access-Control-Allow-Methods"); got == "*" || got == "" {
		t.Fatalf("Access-Control-Allow-Methods = %q, want explicit methods", got)
	}
}

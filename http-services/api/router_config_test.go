package api

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"http-services/config"

	"github.com/gin-gonic/gin"
)

func TestRouterUsesConfiguredCORSProxiesAndStaticDirectory(t *testing.T) {
	previousMaxBodySize := config.MaxBodySize
	previousCORS := config.EnableCORS
	previousProxies := config.TrustedProxies
	previousStaticDir := config.StaticDir
	t.Cleanup(func() {
		config.MaxBodySize = previousMaxBodySize
		config.EnableCORS = previousCORS
		config.TrustedProxies = previousProxies
		config.StaticDir = previousStaticDir
	})

	staticDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(staticDir, "probe.txt"), []byte("static-ok"), 0o600); err != nil {
		t.Fatal(err)
	}
	config.MaxBodySize = 1 << 20
	config.EnableCORS = true
	config.TrustedProxies = []string{"127.0.0.1", "::1"}
	config.StaticDir = staticDir
	router := InitApi()
	router.GET("/client-ip", func(c *gin.Context) { c.String(http.StatusOK, c.ClientIP()) })

	preflight := httptest.NewRequest(http.MethodOptions, "/api/v1/open/health", nil)
	preflight.Header.Set("Origin", "https://client.example")
	preflightRecorder := httptest.NewRecorder()
	router.ServeHTTP(preflightRecorder, preflight)
	if preflightRecorder.Code != http.StatusNoContent {
		t.Fatalf("preflight status = %d", preflightRecorder.Code)
	}
	if got := preflightRecorder.Header().Get("Access-Control-Allow-Headers"); got != corsAllowedHeadersForTest {
		t.Fatalf("allow headers = %q", got)
	}

	staticRecorder := httptest.NewRecorder()
	router.ServeHTTP(staticRecorder, httptest.NewRequest(http.MethodGet, "/static/probe.txt", nil))
	if staticRecorder.Code != http.StatusOK || staticRecorder.Body.String() != "static-ok" {
		t.Fatalf("static response = %d %q", staticRecorder.Code, staticRecorder.Body.String())
	}

	proxyRequest := httptest.NewRequest(http.MethodGet, "/client-ip", nil)
	proxyRequest.RemoteAddr = "127.0.0.1:12345"
	proxyRequest.Header.Set("X-Forwarded-For", "203.0.113.10")
	proxyRecorder := httptest.NewRecorder()
	router.ServeHTTP(proxyRecorder, proxyRequest)
	if proxyRecorder.Body.String() != "203.0.113.10" {
		t.Fatalf("client IP = %q", proxyRecorder.Body.String())
	}
}

const corsAllowedHeadersForTest = "Authorization, Content-Type, X-Trace-ID"

func TestRouterCanDisableCORSAndStaticFiles(t *testing.T) {
	previousMaxBodySize := config.MaxBodySize
	previousCORS := config.EnableCORS
	previousProxies := config.TrustedProxies
	previousStaticDir := config.StaticDir
	t.Cleanup(func() {
		config.MaxBodySize = previousMaxBodySize
		config.EnableCORS = previousCORS
		config.TrustedProxies = previousProxies
		config.StaticDir = previousStaticDir
	})
	config.MaxBodySize = 1 << 20
	config.EnableCORS = false
	config.TrustedProxies = []string{"127.0.0.1", "::1"}
	config.StaticDir = ""
	router := InitApi()

	request := httptest.NewRequest(http.MethodOptions, "/api/v1/open/health", nil)
	request.Header.Set("Origin", "https://client.example")
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)
	if recorder.Code == http.StatusNoContent || recorder.Header().Get("Access-Control-Allow-Origin") != "" {
		t.Fatalf("disabled CORS response = %d headers %#v", recorder.Code, recorder.Header())
	}

	staticRecorder := httptest.NewRecorder()
	router.ServeHTTP(staticRecorder, httptest.NewRequest(http.MethodGet, "/static/probe.txt", nil))
	if staticRecorder.Code != http.StatusNotFound {
		t.Fatalf("disabled static status = %d", staticRecorder.Code)
	}
}

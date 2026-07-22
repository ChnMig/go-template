package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"

	"http-services/config"

	"github.com/gin-gonic/gin"
)

// openHealthResponse 用于解析通过路由访问健康检查接口的统一响应
type openHealthResponse struct {
	Code   int                    `json:"code"`
	Status string                 `json:"status"`
	Detail map[string]interface{} `json:"detail"`
}

// 测试开放路由是否按分层注册成功（health）
func TestOpenHealth(t *testing.T) {
	gin.SetMode(gin.TestMode)
	// 避免未加载配置导致请求体限制为0
	config.MaxBodySize = 10 << 20 // 10MB

	r := InitApi()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/api/v1/open/health", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var body openHealthResponse
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("invalid json: %v", err)
	}

	if body.Code != 200 {
		t.Fatalf("unexpected code: %v", body.Code)
	}

	if body.Status != "OK" {
		t.Fatalf("unexpected wrapper status: %v", body.Status)
	}

	status, ok := body.Detail["status"].(string)
	if !ok || status != "ok" {
		t.Fatalf("unexpected detail.status: %v", body.Detail["status"])
	}
}

func TestInitApiMiddlewareOrder(t *testing.T) {
	router := InitApi()
	if len(router.Handlers) < 3 {
		t.Fatalf("global middleware count = %d, want at least 3", len(router.Handlers))
	}

	want := []string{
		".TraceIDWithDependencies.func",
		".AccessLogWithLogger.func",
		".RecoveryWithLogger.func",
	}
	for i, namePart := range want {
		got := runtime.FuncForPC(reflect.ValueOf(router.Handlers[i]).Pointer()).Name()
		if !strings.Contains(got, namePart) {
			t.Fatalf("middleware[%d] = %s, want name containing %s", i, got, namePart)
		}
	}
}

func TestInitApiUsesConfiguredStaticDirectory(t *testing.T) {
	gin.SetMode(gin.TestMode)
	restoreRouterConfig(t)

	staticDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(staticDir, "hello.txt"), []byte("hello"), 0o600); err != nil {
		t.Fatalf("write static fixture: %v", err)
	}
	config.StaticDir = staticDir
	config.TrustedProxies = []string{"127.0.0.1", "::1"}
	config.MaxBodySize = 10 << 20

	router := InitApi()
	w := httptest.NewRecorder()
	router.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/static/hello.txt", nil))
	if w.Code != http.StatusOK || w.Body.String() != "hello" {
		t.Fatalf("static response = %d %q", w.Code, w.Body.String())
	}
}

func TestInitApiUsesConfiguredTrustedProxies(t *testing.T) {
	gin.SetMode(gin.TestMode)
	restoreRouterConfig(t)

	config.StaticDir = ""
	config.TrustedProxies = []string{"10.0.0.0/8"}
	config.MaxBodySize = 10 << 20
	router := InitApi()
	router.GET("/client-ip", func(c *gin.Context) { c.String(http.StatusOK, c.ClientIP()) })

	req := httptest.NewRequest(http.MethodGet, "/client-ip", nil)
	req.RemoteAddr = "10.1.2.3:1234"
	req.Header.Set("X-Forwarded-For", "203.0.113.9")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Body.String() != "203.0.113.9" {
		t.Fatalf("client IP = %q, want forwarded IP", w.Body.String())
	}
}

func TestInitApiEnablesCORSFromConfig(t *testing.T) {
	gin.SetMode(gin.TestMode)
	restoreRouterConfig(t)

	config.StaticDir = ""
	config.TrustedProxies = nil
	config.EnableCORS = true
	config.MaxBodySize = 10 << 20
	router := InitApi()

	req := httptest.NewRequest(http.MethodOptions, "/api/v1/open/health", nil)
	req.Header.Set("Origin", "https://example.com")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusNoContent {
		t.Fatalf("CORS preflight status = %d, want 204", w.Code)
	}
}

func restoreRouterConfig(t *testing.T) {
	t.Helper()
	oldStaticDir := config.StaticDir
	oldTrustedProxies := append([]string(nil), config.TrustedProxies...)
	oldEnableCORS := config.EnableCORS
	oldMaxBodySize := config.MaxBodySize
	t.Cleanup(func() {
		config.StaticDir = oldStaticDir
		config.TrustedProxies = oldTrustedProxies
		config.EnableCORS = oldEnableCORS
		config.MaxBodySize = oldMaxBodySize
	})
}

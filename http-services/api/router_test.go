package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"http-services/config"
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

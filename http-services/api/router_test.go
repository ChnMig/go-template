package api

import (
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"

    "github.com/gin-gonic/gin"
    "http-services/config"
    "http-services/utils/authentication"
)

// 测试开放路由是否按分层注册成功
func TestOpenExamplePong(t *testing.T) {
    gin.SetMode(gin.TestMode)
    // 避免未加载配置导致请求体限制为0
    // 配置中间件依赖此值
    // 10MB
    config.MaxBodySize = 10 << 20

    r := InitApi()

    w := httptest.NewRecorder()
    req, _ := http.NewRequest(http.MethodGet, "/api/v1/open/example/pong", nil)
    r.ServeHTTP(w, req)

    if w.Code != http.StatusOK {
        t.Fatalf("expected 200, got %d", w.Code)
    }

    var body map[string]any
    if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
        t.Fatalf("invalid json: %v", err)
    }
    if code, ok := body["code"].(float64); !ok || int(code) != 200 {
        t.Fatalf("unexpected response code: %v", body["code"])
    }
}

// 测试私有路由经 Token 验证后访问成功（确保子路由注册生效）
func TestPrivateExamplePongWithToken(t *testing.T) {
    gin.SetMode(gin.TestMode)
    // 避免未加载配置导致请求体限制为0
    config.MaxBodySize = 10 << 20

    r := InitApi()

    // 生成测试用 token（无需登录接口）
    token, err := authentication.JWTIssue(map[string]interface{}{"user_id": "12345", "username": "admin"})
    if err != nil || token == "" {
        t.Fatalf("failed to issue token: %v", err)
    }

    // 使用 token 访问私有接口
    w2 := httptest.NewRecorder()
    req2, _ := http.NewRequest(http.MethodGet, "/api/v1/private/example/pong", nil)
    req2.Header.Set("token", token)
    r.ServeHTTP(w2, req2)
    if w2.Code != http.StatusOK {
        t.Fatalf("expected 200 with token, got %d", w2.Code)
    }
    var body2 map[string]any
    if err := json.Unmarshal(w2.Body.Bytes(), &body2); err != nil {
        t.Fatalf("invalid json: %v", err)
    }
    if code, ok := body2["code"].(float64); !ok || int(code) != 200 {
        t.Fatalf("unexpected response code: %v", body2["code"])
    }
}

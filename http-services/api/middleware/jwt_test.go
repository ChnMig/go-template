package middleware

import (
	"net/http/httptest"
	"testing"
	"time"

	"http-services/config"
	"http-services/utils/authentication"

	"github.com/gin-gonic/gin"
)

func TestTokenVerify_ValidToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// 设置测试配置
	config.JWTKey = "test-secret-key-for-testing-at-least-32-chars"
	config.JWTExpiration = 1 * time.Hour

	// 创建有效的 token
	userData := map[string]interface{}{"user_id": "user123"}
	token, err := authentication.JWTIssue(userData)
	if err != nil {
		t.Fatalf("Failed to issue token: %v", err)
	}

	// 创建测试路由
	router := gin.New()
	router.Use(TokenVerify)
	router.GET("/test", func(c *gin.Context) {
		jwtData, exists := c.Get("jwtData")
		if !exists {
			t.Error("jwtData not found in context")
		}
		c.JSON(200, gin.H{"message": "ok", "data": jwtData})
	})

	// 发送带有效 token 的请求
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("token", token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("Expected 200, got %d", w.Code)
	}
}

func TestTokenVerify_MissingToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// 创建测试路由
	router := gin.New()
	router.Use(TokenVerify)
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "ok"})
	})

	// 发送不带 token 的请求
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// response.ReturnError 返回 HTTP 200，错误信息在 body 中
	if w.Code != 200 {
		t.Errorf("Expected 200, got %d", w.Code)
	}
	// 验证响应体包含 401 错误码
	if !contains(w.Body.String(), "401") || !contains(w.Body.String(), "UNAUTHENTICATED") {
		t.Errorf("Expected UNAUTHENTICATED error in body, got: %s", w.Body.String())
	}
}

func TestTokenVerify_InvalidToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// 创建测试路由
	router := gin.New()
	router.Use(TokenVerify)
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "ok"})
	})

	// 发送带无效 token 的请求
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("token", "invalid.token.here")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// response.ReturnError 返回 HTTP 200，错误信息在 body 中
	if w.Code != 200 {
		t.Errorf("Expected 200, got %d", w.Code)
	}
	// 验证响应体包含 401 错误码
	if !contains(w.Body.String(), "401") || !contains(w.Body.String(), "UNAUTHENTICATED") {
		t.Errorf("Expected UNAUTHENTICATED error in body, got: %s", w.Body.String())
	}
}

func TestTokenVerify_DifferentKey(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// 使用一个密钥创建 token
	config.JWTKey = "test-secret-key-for-testing-at-least-32-chars"
	config.JWTExpiration = 1 * time.Hour
	userData := map[string]interface{}{"user_id": "user123"}
	token, err := authentication.JWTIssue(userData)
	if err != nil {
		t.Fatalf("Failed to issue token: %v", err)
	}

	// 更换密钥后验证（模拟密钥轮换场景）
	config.JWTKey = "different-secret-key-for-testing-32chars"

	// 创建测试路由
	router := gin.New()
	router.Use(TokenVerify)
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "ok"})
	})

	// 发送用旧密钥生成的 token
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("token", token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// response.ReturnError 返回 HTTP 200，错误信息在 body 中
	if w.Code != 200 {
		t.Errorf("Expected 200, got %d", w.Code)
	}
	// 验证响应体包含 401 错误码（密钥不匹配导致验证失败）
	if !contains(w.Body.String(), "401") || !contains(w.Body.String(), "UNAUTHENTICATED") {
		t.Errorf("Expected UNAUTHENTICATED error in body, got: %s", w.Body.String())
	}

	// 恢复密钥
	config.JWTKey = "test-secret-key-for-testing-at-least-32-chars"
}

// contains 检查字符串是否包含子串
func contains(s, substr string) bool {
	return len(s) >= len(substr) && findSubstring(s, substr)
}

func findSubstring(s, substr string) bool {
	if len(substr) == 0 {
		return true
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

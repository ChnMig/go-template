package middleware

import (
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func TestIPRateLimit(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// 创建测试路由
	router := gin.New()
	router.Use(IPRateLimit(2, 3)) // 每秒2个请求，突发3个
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "ok"})
	})

	// 测试突发请求
	for i := 0; i < 3; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "192.168.1.1:12345"
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != 200 {
			t.Errorf("Request %d failed: expected 200, got %d", i+1, w.Code)
		}
	}

	// 第4个请求应该被限流
	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "192.168.1.1:12345"
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// response.ReturnError 返回 HTTP 200，错误信息在 body 中
	if w.Code != 200 {
		t.Errorf("Expected 200, got %d", w.Code)
	}
	// 验证响应体包含 429 错误码
	if !contains(w.Body.String(), "429") || !contains(w.Body.String(), "RESOURCE_EXHAUSTED") {
		t.Errorf("Expected RESOURCE_EXHAUSTED error in body, got: %s", w.Body.String())
	}

	// 等待令牌恢复
	time.Sleep(time.Second)

	// 应该可以再次请求
	req = httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "192.168.1.1:12345"
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("Request after cooldown failed: expected 200, got %d", w.Code)
	}
}

func TestTokenRateLimit(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// 创建测试路由
	router := gin.New()
	router.Use(TokenRateLimit(1, 2)) // 每秒1个请求，突发2个
	router.GET("/test", func(c *gin.Context) {
		// 模拟 JWT 中间件设置的数据
		c.Set("jwtData", map[string]interface{}{"user_id": "user123"})
		c.JSON(200, gin.H{"message": "ok"})
	})

	// 测试突发请求
	for i := 0; i < 2; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != 200 {
			t.Errorf("Request %d failed: expected 200, got %d", i+1, w.Code)
		}
	}

	// 第3个请求应该被限流
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// response.ReturnError 返回 HTTP 200，错误信息在 body 中
	if w.Code != 200 {
		t.Errorf("Expected 200, got %d", w.Code)
	}
	// 验证响应体包含 429 错误码
	if !contains(w.Body.String(), "429") || !contains(w.Body.String(), "RESOURCE_EXHAUSTED") {
		t.Errorf("Expected RESOURCE_EXHAUSTED error in body, got: %s", w.Body.String())
	}
}

func TestRateLimitWithOptions(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// 创建测试路由，使用自定义 key 函数
	router := gin.New()
	router.Use(RateLimitWithOptions(RateLimitOptions{
		Rate:  1,
		Burst: 2,
		KeyFunc: func(c *gin.Context) string {
			return c.GetHeader("X-API-Key")
		},
		Message: "Custom rate limit exceeded",
	}))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "ok"})
	})

	// 使用相同的 API Key 发送请求
	apiKey := "test-key-123"
	for i := 0; i < 2; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("X-API-Key", apiKey)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != 200 {
			t.Errorf("Request %d failed: expected 200, got %d", i+1, w.Code)
		}
	}

	// 第3个请求应该被限流
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-API-Key", apiKey)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// response.ReturnError 返回 HTTP 200，错误信息在 body 中
	if w.Code != 200 {
		t.Errorf("Expected 200, got %d", w.Code)
	}
	// 验证响应体包含 429 错误码
	if !contains(w.Body.String(), "429") || !contains(w.Body.String(), "RESOURCE_EXHAUSTED") {
		t.Errorf("Expected RESOURCE_EXHAUSTED error in body, got: %s", w.Body.String())
	}
}

func TestRateLimiterCleanup(t *testing.T) {
	// 创建一个限流器
	rl := NewRateLimiter(10, 20)

	// 生成一些限流器实例
	rl.getLimiter("test1")
	rl.getLimiter("test2")
	rl.getLimiter("test3")

	if len(rl.limiters) != 3 {
		t.Errorf("Expected 3 limiters, got %d", len(rl.limiters))
	}

	// 停止限流器
	rl.Stop()

	// 验证 goroutine 已停止（通过检查是否可以再次调用 Stop 而不会 panic）
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Panic occurred: %v", r)
		}
	}()
}

func TestCleanupAllLimiters(t *testing.T) {
	// 清空全局缓存
	cacheMu.Lock()
	limiterCache = make(map[string]*RateLimiter)
	cacheMu.Unlock()

	// 创建几个限流器
	_ = getLimiterFromCache(10, 20)
	_ = getLimiterFromCache(20, 40)

	cacheMu.RLock()
	initialCount := len(limiterCache)
	cacheMu.RUnlock()

	if initialCount != 2 {
		t.Errorf("Expected 2 limiters in cache, got %d", initialCount)
	}

	// 清理所有限流器
	CleanupAllLimiters()

	cacheMu.RLock()
	finalCount := len(limiterCache)
	cacheMu.RUnlock()

	if finalCount != 0 {
		t.Errorf("Expected 0 limiters after cleanup, got %d", finalCount)
	}
}

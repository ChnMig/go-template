package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"http-services/utils/contextkey"

	"github.com/gin-gonic/gin"
)

func TestIPRateLimitUsesIndependentClientBuckets(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handled := 0
	router := gin.New()
	router.Use(IPRateLimit(1, 1))
	router.GET("/test", func(c *gin.Context) {
		handled++
		c.Status(http.StatusNoContent)
	})

	serveRateLimitRequest(router, "192.0.2.1:1000", "", "")
	limited := serveRateLimitRequest(router, "192.0.2.1:1001", "", "")
	serveRateLimitRequest(router, "192.0.2.2:1000", "", "")

	if handled != 2 {
		t.Fatalf("业务 handler 执行 %d 次，want 2", handled)
	}
	assertRateLimited(t, limited)
}

func TestTokenRateLimitUsesJWTIdentity(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handled := 0
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set(contextkey.JWTData, map[string]interface{}{"user_id": c.GetHeader("X-User-ID")})
		c.Next()
	})
	router.Use(TokenRateLimit(1, 1))
	router.GET("/test", func(c *gin.Context) {
		handled++
		c.Status(http.StatusNoContent)
	})

	serveRateLimitRequest(router, "192.0.2.1:1000", "user-a", "")
	serveRateLimitRequest(router, "192.0.2.1:1000", "user-b", "")
	limited := serveRateLimitRequest(router, "192.0.2.1:1000", "user-a", "")

	if handled != 2 {
		t.Fatalf("业务 handler 执行 %d 次，want 2", handled)
	}
	assertRateLimited(t, limited)
}

func TestRateLimitWithOptionsUsesCustomKeyAndMessage(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handled := 0
	router := gin.New()
	router.Use(RateLimitWithOptions(RateLimitOptions{
		Rate:  1,
		Burst: 1,
		KeyFunc: func(c *gin.Context) string {
			return c.GetHeader("X-API-Key")
		},
		Message: "Custom rate limit exceeded",
	}))
	router.GET("/test", func(c *gin.Context) {
		handled++
		c.Status(http.StatusNoContent)
	})

	serveRateLimitRequest(router, "192.0.2.1:1000", "", "key-a")
	serveRateLimitRequest(router, "192.0.2.1:1000", "", "key-b")
	limited := serveRateLimitRequest(router, "192.0.2.1:1000", "", "key-a")

	if handled != 2 {
		t.Fatalf("业务 handler 执行 %d 次，want 2", handled)
	}
	assertRateLimited(t, limited)
	if !contains(limited.Body.String(), "Custom rate limit exceeded") {
		t.Fatalf("限流响应未包含自定义文案: %s", limited.Body.String())
	}
}

func TestRateLimiterLazilyRemovesExpiredEntries(t *testing.T) {
	limiter := NewRateLimiter(10, 20)
	limiter.ttl = time.Minute
	limiter.maxEntries = 10

	startedAt := time.Unix(1_000, 0)
	limiter.allowAt("old-a", startedAt)
	limiter.allowAt("old-b", startedAt)
	limiter.allowAt("new", startedAt.Add(2*time.Minute))

	if _, exists := limiter.limiters["old-a"]; exists {
		t.Fatal("过期条目 old-a 未被惰性清理")
	}
	if _, exists := limiter.limiters["old-b"]; exists {
		t.Fatal("过期条目 old-b 未被惰性清理")
	}
	if len(limiter.limiters) != 1 {
		t.Fatalf("条目数 = %d, want 1", len(limiter.limiters))
	}
}

func TestRateLimiterEvictsOldestAtCapacity(t *testing.T) {
	limiter := NewRateLimiter(10, 20)
	limiter.ttl = time.Hour
	limiter.maxEntries = 2

	startedAt := time.Unix(2_000, 0)
	limiter.allowAt("oldest", startedAt)
	limiter.allowAt("second", startedAt.Add(time.Second))
	limiter.allowAt("third", startedAt.Add(2*time.Second))

	if len(limiter.limiters) != 2 {
		t.Fatalf("条目数 = %d, want 2", len(limiter.limiters))
	}
	if _, exists := limiter.limiters["oldest"]; exists {
		t.Fatal("达到容量上限后未淘汰最旧条目")
	}
}

func serveRateLimitRequest(router http.Handler, remoteAddr, userID, apiKey string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.RemoteAddr = remoteAddr
	if userID != "" {
		req.Header.Set("X-User-ID", userID)
	}
	if apiKey != "" {
		req.Header.Set("X-API-Key", apiKey)
	}
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w
}

func assertRateLimited(t *testing.T, response *httptest.ResponseRecorder) {
	t.Helper()
	if !contains(response.Body.String(), "429") || !contains(response.Body.String(), "RESOURCE_EXHAUSTED") {
		t.Fatalf("want RESOURCE_EXHAUSTED response, got %s", response.Body.String())
	}
}

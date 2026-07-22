package middleware

import (
	"sync"
	"time"

	"http-services/api/response"
	"http-services/utils/contextkey"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

const (
	MaxLimiters = 10000
	DefaultTTL  = 5 * time.Minute
)

type limiterEntry struct {
	limiter    *rate.Limiter
	lastAccess time.Time
}

// RateLimiter 管理一组有容量上限、按需清理的令牌桶。
type RateLimiter struct {
	limiters    map[string]*limiterEntry
	lastCleanup time.Time
	mu          sync.Mutex
	rate        int
	burst       int
	ttl         time.Duration
	maxEntries  int
}

// NewRateLimiter 创建不启动后台 goroutine 的限流器。
func NewRateLimiter(requestsPerSecond, burst int) *RateLimiter {
	return &RateLimiter{
		limiters:   make(map[string]*limiterEntry),
		rate:       requestsPerSecond,
		burst:      burst,
		ttl:        DefaultTTL,
		maxEntries: MaxLimiters,
	}
}

func (limiter *RateLimiter) allow(key string) bool {
	return limiter.allowAt(key, time.Now())
}

func (limiter *RateLimiter) allowAt(key string, now time.Time) bool {
	bucket := limiter.getLimiterAt(key, now)
	return bucket.AllowN(now, 1)
}

func (limiter *RateLimiter) getLimiterAt(key string, now time.Time) *rate.Limiter {
	limiter.mu.Lock()
	defer limiter.mu.Unlock()

	if limiter.lastCleanup.IsZero() {
		limiter.lastCleanup = now
	} else if now.Sub(limiter.lastCleanup) >= limiter.ttl {
		limiter.cleanupLocked(now)
		limiter.lastCleanup = now
	}

	if entry, exists := limiter.limiters[key]; exists {
		entry.lastAccess = now
		return entry.limiter
	}

	if len(limiter.limiters) >= limiter.maxEntries {
		limiter.cleanupLocked(now)
	}
	if len(limiter.limiters) >= limiter.maxEntries {
		limiter.removeOldestLocked()
	}

	bucket := rate.NewLimiter(rate.Limit(limiter.rate), limiter.burst)
	limiter.limiters[key] = &limiterEntry{limiter: bucket, lastAccess: now}
	return bucket
}

func (limiter *RateLimiter) cleanupLocked(now time.Time) {
	for key, entry := range limiter.limiters {
		if now.Sub(entry.lastAccess) > limiter.ttl {
			delete(limiter.limiters, key)
		}
	}
}

func (limiter *RateLimiter) removeOldestLocked() {
	var oldestKey string
	var oldestTime time.Time
	for key, entry := range limiter.limiters {
		if oldestKey == "" || entry.lastAccess.Before(oldestTime) {
			oldestKey = key
			oldestTime = entry.lastAccess
		}
	}
	if oldestKey != "" {
		delete(limiter.limiters, oldestKey)
	}
}

// Stats 是单个限流器的当前统计快照。
type Stats struct {
	TotalLimiters int
	Rate          int
	Burst         int
	TTL           time.Duration
}

// GetStats 返回并发安全的统计快照。
func (limiter *RateLimiter) GetStats() Stats {
	limiter.mu.Lock()
	defer limiter.mu.Unlock()
	return Stats{
		TotalLimiters: len(limiter.limiters),
		Rate:          limiter.rate,
		Burst:         limiter.burst,
		TTL:           limiter.ttl,
	}
}

// RateLimitOptions 定义自定义 key 限流参数。
type RateLimitOptions struct {
	Rate    int
	Burst   int
	KeyFunc func(*gin.Context) string
	Message string
}

// IPRateLimit 按可信客户端 IP 限流。
func IPRateLimit(requestsPerSecond, burst int) gin.HandlerFunc {
	limiter := NewRateLimiter(requestsPerSecond, burst)
	return rateLimitHandler(limiter, func(c *gin.Context) string { return c.ClientIP() }, "IP rate limit exceeded")
}

// TokenRateLimit 优先按 JWT 用户标识限流，缺失时回退到客户端 IP。
func TokenRateLimit(requestsPerSecond, burst int) gin.HandlerFunc {
	limiter := NewRateLimiter(requestsPerSecond, burst)
	return rateLimitHandler(limiter, getTokenKey, "Rate limit exceeded")
}

// RateLimitWithOptions 使用调用方提供的 key 和错误文案限流。
func RateLimitWithOptions(options RateLimitOptions) gin.HandlerFunc {
	if options.KeyFunc == nil {
		options.KeyFunc = func(c *gin.Context) string { return c.ClientIP() }
	}
	if options.Message == "" {
		options.Message = "Rate limit exceeded"
	}
	return rateLimitHandler(NewRateLimiter(options.Rate, options.Burst), options.KeyFunc, options.Message)
}

func rateLimitHandler(limiter *RateLimiter, keyFunc func(*gin.Context) string, message string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !limiter.allow(keyFunc(c)) {
			response.ReturnError(c, response.RESOURCE_EXHAUSTED, message)
			return
		}
		c.Next()
	}
}

func getTokenKey(c *gin.Context) string {
	jwtData, exists := c.Get(contextkey.JWTData)
	if !exists {
		return c.ClientIP()
	}

	var key string
	switch value := jwtData.(type) {
	case string:
		key = value
	case map[string]interface{}:
		if id, ok := value["id"].(string); ok {
			key = id
		} else if userID, ok := value["user_id"].(string); ok {
			key = userID
		}
	}
	if key == "" {
		return c.ClientIP()
	}
	return key
}

func StrictRateLimit() gin.HandlerFunc {
	return IPRateLimit(5, 10)
}

func ModerateRateLimit() gin.HandlerFunc {
	return IPRateLimit(50, 100)
}

func RelaxedRateLimit() gin.HandlerFunc {
	return IPRateLimit(100, 200)
}

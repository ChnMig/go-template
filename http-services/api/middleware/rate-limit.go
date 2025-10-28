package middleware

import (
	"fmt"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"

	"http-services/api/response"
)

// limiterEntry 限流器条目，包含限流器和最后访问时间
type limiterEntry struct {
	limiter    *rate.Limiter
	lastAccess time.Time
}

// RateLimiter 限流管理器
type RateLimiter struct {
	mu       sync.RWMutex
	limiters map[string]*limiterEntry
	rate     int           // 每秒请求数
	burst    int           // 突发请求数
	ttl      time.Duration // 限流器过期时间
	ticker   *time.Ticker  // 清理定时器
	stopChan chan struct{} // 停止信号
}

// NewRateLimiter 创建新的限流管理器
func NewRateLimiter(r, b int) *RateLimiter {
	rl := &RateLimiter{
		limiters: make(map[string]*limiterEntry),
		rate:     r,
		burst:    b,
		ttl:      10 * time.Minute, // 10分钟未访问则清理
		stopChan: make(chan struct{}),
	}
	// 启动自动清理
	rl.startCleanup(5 * time.Minute)
	return rl
}

// getLimiter 获取或创建限流器
func (rl *RateLimiter) getLimiter(key string) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	entry, exists := rl.limiters[key]
	if !exists {
		limiter := rate.NewLimiter(rate.Limit(rl.rate), rl.burst)
		rl.limiters[key] = &limiterEntry{
			limiter:    limiter,
			lastAccess: time.Now(),
		}
		return limiter
	}

	// 更新最后访问时间
	entry.lastAccess = time.Now()
	return entry.limiter
}

// allow 检查是否允许请求
func (rl *RateLimiter) allow(key string) bool {
	limiter := rl.getLimiter(key)
	return limiter.Allow()
}

// cleanup 清理过期的限流器
func (rl *RateLimiter) cleanup() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	for key, entry := range rl.limiters {
		if now.Sub(entry.lastAccess) > rl.ttl {
			delete(rl.limiters, key)
		}
	}
}

// startCleanup 启动定期清理任务
func (rl *RateLimiter) startCleanup(interval time.Duration) {
	rl.ticker = time.NewTicker(interval)
	go func() {
		for {
			select {
			case <-rl.ticker.C:
				rl.cleanup()
			case <-rl.stopChan:
				rl.ticker.Stop()
				return
			}
		}
	}()
}

// Stop 停止限流器的清理任务
func (rl *RateLimiter) Stop() {
	close(rl.stopChan)
}

// limiters 全局限流器缓存，key 为 "rate-burst" 组合
var (
	limiterCache = make(map[string]*RateLimiter)
	cacheMu      sync.RWMutex
)

// CleanupAllLimiters 清理所有限流器资源（应用关闭时调用）
func CleanupAllLimiters() {
	cacheMu.Lock()
	defer cacheMu.Unlock()

	for _, limiter := range limiterCache {
		limiter.Stop()
	}
	// 清空缓存
	limiterCache = make(map[string]*RateLimiter)
}

// getLimiterFromCache 从缓存中获取或创建限流器
func getLimiterFromCache(r, b int) *RateLimiter {
	key := fmt.Sprintf("%d-%d", r, b)

	// 先尝试读取
	cacheMu.RLock()
	if limiter, exists := limiterCache[key]; exists {
		cacheMu.RUnlock()
		return limiter
	}
	cacheMu.RUnlock()

	// 需要创建新的限流器
	cacheMu.Lock()
	defer cacheMu.Unlock()

	// 双重检查，防止并发创建
	if limiter, exists := limiterCache[key]; exists {
		return limiter
	}

	limiter := NewRateLimiter(r, b)
	limiterCache[key] = limiter
	return limiter
}

// RateLimitOptions 限流配置选项
type RateLimitOptions struct {
	Rate    int                       // 每秒请求数
	Burst   int                       // 突发请求数
	KeyFunc func(*gin.Context) string // 自定义获取限流 key 的函数
	Message string                    // 自定义错误消息
}

// IPRateLimit IP 限流中间件（可指定速率）
// 参数：
//   - r: 每秒请求数（rate），例如 10 表示每秒最多 10 个请求
//   - b: 突发请求数（burst），例如 20 表示桶最大容量为 20
//
// 示例：
//
//	middleware.IPRateLimit(10, 20)  // 每秒10个请求，突发20个
func IPRateLimit(r, b int) gin.HandlerFunc {
	limiter := getLimiterFromCache(r, b)

	return func(c *gin.Context) {
		// 获取客户端 IP
		ip := c.ClientIP()

		// 检查是否允许请求
		if !limiter.allow(ip) {
			response.ReturnError(c, response.RESOURCE_EXHAUSTED, "IP rate limit exceeded")
			return
		}

		c.Next()
	}
}

// TokenRateLimit Token 限流中间件（可指定速率）
// 参数：
//   - r: 每秒请求数（rate）
//   - b: 突发请求数（burst）
//
// 说明：
//   - 优先使用 JWT 中的用户标识作为限流 key
//   - 如果没有 JWT 数据，回退到使用 IP
//   - 需要在 TokenVerify 中间件之后使用
//
// 示例：
//
//	middleware.TokenRateLimit(100, 200)  // 每秒100个请求，突发200个
func TokenRateLimit(r, b int) gin.HandlerFunc {
	limiter := getLimiterFromCache(r, b)

	return func(c *gin.Context) {
		// 从 context 中获取 JWT 数据（需要在 TokenVerify 中间件之后使用）
		key := getTokenKey(c)

		// 检查是否允许请求
		if !limiter.allow(key) {
			response.ReturnError(c, response.RESOURCE_EXHAUSTED, "Rate limit exceeded")
			return
		}

		c.Next()
	}
}

// RateLimitWithOptions 自定义限流中间件（高级用法）
// 参数：
//   - opts: 限流配置选项
//
// 示例：
//
//	middleware.RateLimitWithOptions(middleware.RateLimitOptions{
//	    Rate: 50,
//	    Burst: 100,
//	    KeyFunc: func(c *gin.Context) string {
//	        // 自定义 key 生成逻辑，例如按 API Key 限流
//	        return c.GetHeader("X-API-Key")
//	    },
//	    Message: "API rate limit exceeded",
//	})
func RateLimitWithOptions(opts RateLimitOptions) gin.HandlerFunc {
	limiter := getLimiterFromCache(opts.Rate, opts.Burst)

	// 设置默认 KeyFunc
	if opts.KeyFunc == nil {
		opts.KeyFunc = func(c *gin.Context) string {
			return c.ClientIP()
		}
	}

	// 设置默认错误消息
	if opts.Message == "" {
		opts.Message = "Rate limit exceeded"
	}

	return func(c *gin.Context) {
		key := opts.KeyFunc(c)

		// 检查是否允许请求
		if !limiter.allow(key) {
			response.ReturnError(c, response.RESOURCE_EXHAUSTED, opts.Message)
			return
		}

		c.Next()
	}
}

// getTokenKey 从 context 中提取 token key
func getTokenKey(c *gin.Context) string {
	// 从 context 中获取 JWT 数据（需要在 TokenVerify 中间件之后使用）
	jwtData, exists := c.Get("jwtData")
	if !exists {
		// 如果没有 JWT 数据，使用 IP 作为 key
		return c.ClientIP()
	}

	// 使用 JWT 数据中的用户标识作为 key
	key := ""
	switch v := jwtData.(type) {
	case string:
		key = v
	case map[string]interface{}:
		if id, ok := v["id"].(string); ok {
			key = id
		} else if userID, ok := v["user_id"].(string); ok {
			key = userID
		}
	}

	if key == "" {
		// 如果无法获取用户标识，使用 IP 作为 key
		key = c.ClientIP()
	}

	return key
}

// 预定义的常用限流配置

// StrictRateLimit 严格限流（适用于写操作、敏感接口）
// 每秒 5 个请求，突发 10 个
func StrictRateLimit() gin.HandlerFunc {
	return IPRateLimit(5, 10)
}

// ModerateRateLimit 中等限流（适用于一般接口）
// 每秒 50 个请求，突发 100 个
func ModerateRateLimit() gin.HandlerFunc {
	return IPRateLimit(50, 100)
}

// RelaxedRateLimit 宽松限流（适用于读操作）
// 每秒 100 个请求，突发 200 个
func RelaxedRateLimit() gin.HandlerFunc {
	return IPRateLimit(100, 200)
}

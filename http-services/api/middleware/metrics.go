package middleware

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// HTTP 请求总数
	httpRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)

	// HTTP 请求延迟
	httpRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Duration of HTTP requests in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path", "status"},
	)

	// HTTP 请求大小
	httpRequestSize = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_size_bytes",
			Help:    "Size of HTTP requests in bytes",
			Buckets: prometheus.ExponentialBuckets(100, 10, 8),
		},
		[]string{"method", "path"},
	)

	// HTTP 响应大小
	httpResponseSize = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_response_size_bytes",
			Help:    "Size of HTTP responses in bytes",
			Buckets: prometheus.ExponentialBuckets(100, 10, 8),
		},
		[]string{"method", "path"},
	)

	// 当前正在处理的请求数
	httpRequestsInFlight = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "http_requests_in_flight",
			Help: "Current number of HTTP requests being processed",
		},
	)
)

// Metrics Prometheus 监控中间件
func Metrics() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 记录请求开始时间
		start := time.Now()

		// 增加正在处理的请求数
		httpRequestsInFlight.Inc()
		defer httpRequestsInFlight.Dec()

		// 获取路径（优先使用 FullPath，如果为空则使用 URL.Path）
		path := c.FullPath()
		if path == "" {
			path = c.Request.URL.Path
		}

		// 记录请求大小
		if c.Request.ContentLength > 0 {
			httpRequestSize.WithLabelValues(
				c.Request.Method,
				path,
			).Observe(float64(c.Request.ContentLength))
		}

		// 处理请求
		c.Next()

		// 记录请求延迟
		duration := time.Since(start).Seconds()
		status := strconv.Itoa(c.Writer.Status())

		httpRequestDuration.WithLabelValues(
			c.Request.Method,
			path,
			status,
		).Observe(duration)

		// 记录请求总数
		httpRequestsTotal.WithLabelValues(
			c.Request.Method,
			path,
			status,
		).Inc()

		// 记录响应大小
		httpResponseSize.WithLabelValues(
			c.Request.Method,
			path,
		).Observe(float64(c.Writer.Size()))
	}
}

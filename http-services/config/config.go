package config

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"http-services/utils/pathtool"
)

// HTTPConfig 是 Router 构建时使用的启动期不可变配置快照。
type HTTPConfig struct {
	StaticDir       string
	TrustedProxies  []string
	MaxBodySize     int64
	GlobalRateLimit int
	GlobalRateBurst int
	EnableCORS      bool
	EnableRateLimit bool
}

// LogConfig 是允许热更新的日志配置快照。
type LogConfig struct {
	MaxSize  int
	MaxAge   int
	Level    string
	GinLevel string
}

// Here are some basic configurations
// These configurations are usually generic
var (
	// listen
	ListenPort = 8080 // api listen port
	// run model
	RunModelKey      = "model"
	RunModel         = ""
	RunModelDevValue = "dev"
	RunModelRelease  = "release"
	// path
	SelfName = filepath.Base(os.Args[0])      // own file name
	AbsPath  = pathtool.GetCurrentDirectory() // current directory
	// log
	LogDir      = filepath.Join(pathtool.GetCurrentDirectory(), "log")   // log directory
	LogPath     = filepath.Join(LogDir, fmt.Sprintf("%s.log", SelfName)) // self log path
	LogModelDev = "dev"                                                  // dev model
)

var (
	logConfigMu      sync.RWMutex
	runtimeLogConfig = LogConfig{MaxSize: 50, MaxAge: 30, Level: "info"}
)

// SnapshotHTTPConfig 复制当前启动配置，避免 Router 持有可变切片。
func SnapshotHTTPConfig() HTTPConfig {
	return HTTPConfig{
		StaticDir:       StaticDir,
		TrustedProxies:  append([]string(nil), TrustedProxies...),
		MaxBodySize:     MaxBodySize,
		GlobalRateLimit: GlobalRateLimit,
		GlobalRateBurst: GlobalRateBurst,
		EnableCORS:      EnableCORS,
		EnableRateLimit: EnableRateLimit,
	}
}

// CurrentLogConfig 返回并发安全的日志配置快照。
func CurrentLogConfig() LogConfig {
	logConfigMu.RLock()
	defer logConfigMu.RUnlock()
	return runtimeLogConfig
}

// UpdateLogConfig 原子替换日志配置，供 watcher 和受控调用方使用。
func UpdateLogConfig(next LogConfig) {
	logConfigMu.Lock()
	runtimeLogConfig = next
	logConfigMu.Unlock()
}

// 从配置文件加载的配置变量
var (
	// JWT
	JWTKey        string
	JWTExpiration time.Duration

	// Server
	MaxBodySize     int64         // 请求体大小限制（字节）
	ShutdownTimeout time.Duration // 优雅关闭超时时间
	ReadTimeout     time.Duration // 读取超时
	WriteTimeout    time.Duration // 写入超时
	IdleTimeout     time.Duration // 空闲超时
	MaxHeaderBytes  int           // 最大请求头大小
	EnableRateLimit bool          // 是否启用全局限流
	GlobalRateLimit int           // 全局限流速率（每秒请求数）
	GlobalRateBurst int           // 全局限流突发容量
	StaticDir       string        // static 公开资源目录；空值表示关闭
	TrustedProxies  []string      // 可信反向代理 IP 或 CIDR
	EnableCORS      bool          // 是否启用全局 CORS
	PidFile         string        // pid 文件路径（支持相对路径，相对 AbsPath）

	// Log
	LogMaxSize  int
	LogMaxAge   int
	LogLevel    string
	GinLogLevel string

	// Database
	MysqlDSN string // MySQL 数据库连接字符串

	// Redis
	RedisHost      string // Redis 连接地址
	RedisPassword  string // Redis 密码
	RedisKeyPrefix string
)

// 分页配置
var (
	DefaultPageSize = 20 // 默认分页大小
	DefaultPage     = 1  // 默认页码
	CancelPageSize  = -1 // 取消分页大小
	CancelPage      = -1 // 取消页码
)

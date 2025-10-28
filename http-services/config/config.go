package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	pathtool "http-services/utils/path-tool"
)

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
	LogDir        = filepath.Join(pathtool.GetCurrentDirectory(), "log")   // log directory
	LogPath       = filepath.Join(LogDir, fmt.Sprintf("%s.log", SelfName)) // self log path
	LogMaxSize    = 50                                                     // M
	LogMaxBackups = 3                                                      // backups
	LogMaxAge     = 30                                                     // days
	LogModelDev   = "dev"                                                  // dev model
)

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
)

// 分页配置
var (
	DefaultPageSize = 20 // 默认分页大小
	DefaultPage     = 1  // 默认页码
	CancelPageSize  = -1 // 取消分页大小
	CancelPage      = -1 // 取消页码
)

func init() {
	pathtool.CreateDir(LogDir)
	// 配置校验逻辑已移至 main.go，确保 zap logger 初始化后再校验
}

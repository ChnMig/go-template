package config

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

var (
	v *viper.Viper // Viper 实例
)

// LoadConfig 使用 Viper 加载配置
func LoadConfig() error {
	v = viper.New()

	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(currentDirectory())    // 工作目录
	v.AddConfigPath("/etc/http-services/") // 系统目录

	// 支持环境变量（自动转换：HTTP_SERVICES_SERVER_PORT）
	v.SetEnvPrefix("HTTP_SERVICES")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// 设置默认值
	setDefaults()

	// 读取配置文件
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// 配置文件不存在，使用默认值
			zap.L().Warn("Config file not found, using defaults", zap.String("path", currentDirectory()))
		} else {
			// 配置文件存在但读取失败
			return fmt.Errorf("failed to read config file: %w", err)
		}
	} else {
		zap.L().Info("Config file loaded", zap.String("file", v.ConfigFileUsed()))
	}

	// 应用配置到全局变量
	return applyConfig()
}

// setDefaults 设置默认配置值
func setDefaults() {
	// Server 默认配置
	v.SetDefault("server.host", "0.0.0.0")
	v.SetDefault("server.port", 8080)
	v.SetDefault("server.max_body_size", "10MB")
	v.SetDefault("server.max_header_bytes", 1<<20) // 1MB
	v.SetDefault("server.shutdown_timeout", "10s")
	v.SetDefault("server.read_timeout", "30s")
	v.SetDefault("server.write_timeout", "30s")
	v.SetDefault("server.idle_timeout", "120s")
	v.SetDefault("server.enable_rate_limit", false)
	v.SetDefault("server.global_rate_limit", 100)
	v.SetDefault("server.global_rate_burst", 200)
	v.SetDefault("server.pid_file", "http-services.pid")
	v.SetDefault("server.static_dir", "./static")
	v.SetDefault("server.trusted_proxies", []string{"127.0.0.1", "::1"})
	v.SetDefault("server.enable_cors", true)

	// JWT 默认配置
	v.SetDefault("jwt.expiration", "12h")

	// Log 默认配置
	v.SetDefault("log.max_size", 50) // 50MB
	v.SetDefault("log.max_age", 30)  // 保留 30 天
	v.SetDefault("log.level", "info")
	v.SetDefault("log.gin_level", "")

	// Database 默认配置
	v.SetDefault("database.mysql_dsn", "")

	// Redis 默认配置
	v.SetDefault("redis.host", "127.0.0.1:6379")
	v.SetDefault("redis.password", "")
	v.SetDefault("redis.key_prefix", "")
}

// applyConfig 将 Viper 配置应用到全局变量
func applyConfig() error {
	// Server 配置
	ListenHost = strings.TrimSpace(v.GetString("server.host"))
	ListenPort = v.GetInt("server.port")

	// 解析大小字符串
	maxBodySizeStr := v.GetString("server.max_body_size")
	size, err := parseSize(maxBodySizeStr)
	if err != nil {
		return fmt.Errorf("invalid max_body_size: %w", err)
	}
	MaxBodySize = size

	MaxHeaderBytes = v.GetInt("server.max_header_bytes")

	// 解析超时时间
	ShutdownTimeout = v.GetDuration("server.shutdown_timeout")
	ReadTimeout = v.GetDuration("server.read_timeout")
	WriteTimeout = v.GetDuration("server.write_timeout")
	IdleTimeout = v.GetDuration("server.idle_timeout")

	// 限流配置
	EnableRateLimit = v.GetBool("server.enable_rate_limit")
	GlobalRateLimit = v.GetInt("server.global_rate_limit")
	GlobalRateBurst = v.GetInt("server.global_rate_burst")

	// pid 文件（相对路径基于程序所在目录）
	PidFile = strings.TrimSpace(v.GetString("server.pid_file"))
	if PidFile != "" && !filepath.IsAbs(PidFile) {
		PidFile = filepath.Join(AbsPath, PidFile)
	}
	StaticDir = strings.TrimSpace(v.GetString("server.static_dir"))
	TrustedProxies = getStringSlice("server.trusted_proxies")
	EnableCORS = v.GetBool("server.enable_cors")

	// JWT 配置
	JWTKey = v.GetString("jwt.key")
	JWTExpiration = v.GetDuration("jwt.expiration")

	// Log 配置
	LogMaxSize = v.GetInt("log.max_size")
	LogMaxAge = v.GetInt("log.max_age")
	LogLevel = v.GetString("log.level")
	GinLogLevel = v.GetString("log.gin_level")

	// Database 配置
	MysqlDSN = v.GetString("database.mysql_dsn")

	// Redis 配置
	RedisHost = strings.TrimSpace(v.GetString("redis.host"))
	RedisPassword = v.GetString("redis.password")
	RedisKeyPrefix = strings.TrimSpace(v.GetString("redis.key_prefix"))

	return nil
}

func getStringSlice(key string) []string {
	values := v.GetStringSlice(key)
	if len(values) == 1 && strings.Contains(values[0], ",") {
		values = strings.Split(values[0], ",")
	}
	for index := range values {
		values[index] = strings.TrimSpace(values[index])
	}
	return values
}

// WatchConfig 监听配置文件变化并自动重新加载（热重载）
func WatchConfig(onChange func()) {
	if v == nil || v.ConfigFileUsed() == "" {
		return
	}
	v.WatchConfig()
	v.OnConfigChange(func(e fsnotify.Event) {
		zap.L().Info("Config file changed, reloading...",
			zap.String("file", e.Name),
			zap.String("op", e.Op.String()),
		)

		// 重新应用配置
		if err := applyConfig(); err != nil {
			zap.L().Error("Failed to reload config", zap.Error(err))
			return
		}

		// 执行回调
		if onChange != nil {
			onChange()
		}

		zap.L().Info("Config reloaded successfully")
	})
}

// GetViper 返回 Viper 实例（用于高级用法）
func GetViper() *viper.Viper {
	return v
}

// parseSize 解析大小字符串（支持 KB, MB, GB）
func parseSize(sizeStr string) (int64, error) {
	var size int64
	var unit string
	_, err := fmt.Sscanf(sizeStr, "%d%s", &size, &unit)
	if err != nil {
		return 0, err
	}

	switch strings.ToUpper(unit) {
	case "B", "":
		return size, nil
	case "KB", "K":
		return size * 1024, nil
	case "MB", "M":
		return size * 1024 * 1024, nil
	case "GB", "G":
		return size * 1024 * 1024 * 1024, nil
	default:
		return 0, fmt.Errorf("unknown size unit: %s", unit)
	}
}

package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/goccy/go-yaml"
)

// YamlConfig 表示 YAML 配置文件的结构
type YamlConfig struct {
	Server struct {
		Port            int    `yaml:"port"`
		MaxBodySize     string `yaml:"max_body_size"`     // 例如: "10MB"
		ShutdownTimeout string `yaml:"shutdown_timeout"`  // 例如: "10s"
		ReadTimeout     string `yaml:"read_timeout"`      // 例如: "30s"
		WriteTimeout    string `yaml:"write_timeout"`     // 例如: "30s"
		IdleTimeout     string `yaml:"idle_timeout"`      // 例如: "120s"
		MaxHeaderBytes  int    `yaml:"max_header_bytes"`  // 例如: 1048576 (1MB)
		EnableRateLimit bool   `yaml:"enable_rate_limit"` // 是否启用全局限流
		GlobalRateLimit int    `yaml:"global_rate_limit"` // 全局限流速率
		GlobalRateBurst int    `yaml:"global_rate_burst"` // 全局限流突发
	} `yaml:"server"`
	JWT struct {
		Key        string `yaml:"key"`
		Expiration string `yaml:"expiration"`
	} `yaml:"jwt"`
}

// LoadConfig 从 config.yaml 文件加载配置
func LoadConfig() error {
	configPath := filepath.Join(AbsPath, "config.yaml")
	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read config file: %v", err)
	}

	var config YamlConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("failed to parse config file: %v", err)
	}

	// 应用配置值
	if config.Server.Port != 0 {
		ListenPort = config.Server.Port
	}

	// 默认值设置
	MaxBodySize = 10 << 20 // 默认 10MB
	ShutdownTimeout = 10 * time.Second
	ReadTimeout = 30 * time.Second
	WriteTimeout = 30 * time.Second
	IdleTimeout = 120 * time.Second
	MaxHeaderBytes = 1 << 20 // 默认 1MB
	EnableRateLimit = false
	GlobalRateLimit = 100
	GlobalRateBurst = 200

	// 解析服务器配置
	if config.Server.MaxBodySize != "" {
		size, err := parseSize(config.Server.MaxBodySize)
		if err != nil {
			return fmt.Errorf("invalid max_body_size format: %v", err)
		}
		MaxBodySize = size
	}

	if config.Server.ShutdownTimeout != "" {
		timeout, err := time.ParseDuration(config.Server.ShutdownTimeout)
		if err != nil {
			return fmt.Errorf("invalid shutdown_timeout format: %v", err)
		}
		ShutdownTimeout = timeout
	}

	if config.Server.ReadTimeout != "" {
		timeout, err := time.ParseDuration(config.Server.ReadTimeout)
		if err != nil {
			return fmt.Errorf("invalid read_timeout format: %v", err)
		}
		ReadTimeout = timeout
	}

	if config.Server.WriteTimeout != "" {
		timeout, err := time.ParseDuration(config.Server.WriteTimeout)
		if err != nil {
			return fmt.Errorf("invalid write_timeout format: %v", err)
		}
		WriteTimeout = timeout
	}

	if config.Server.IdleTimeout != "" {
		timeout, err := time.ParseDuration(config.Server.IdleTimeout)
		if err != nil {
			return fmt.Errorf("invalid idle_timeout format: %v", err)
		}
		IdleTimeout = timeout
	}

	if config.Server.MaxHeaderBytes != 0 {
		MaxHeaderBytes = config.Server.MaxHeaderBytes
	}

	// 限流配置
	EnableRateLimit = config.Server.EnableRateLimit
	if config.Server.GlobalRateLimit != 0 {
		GlobalRateLimit = config.Server.GlobalRateLimit
	}
	if config.Server.GlobalRateBurst != 0 {
		GlobalRateBurst = config.Server.GlobalRateBurst
	}

	// JWT 配置
	if config.JWT.Key != "" {
		JWTKey = config.JWT.Key
	}

	if config.JWT.Expiration != "" {
		expiration, err := time.ParseDuration(config.JWT.Expiration)
		if err != nil {
			return fmt.Errorf("invalid JWT expiration format: %v", err)
		}
		JWTExpiration = expiration
	}

	return nil
}

// parseSize 解析大小字符串（支持 KB, MB, GB）
func parseSize(sizeStr string) (int64, error) {
	var size int64
	var unit string
	_, err := fmt.Sscanf(sizeStr, "%d%s", &size, &unit)
	if err != nil {
		return 0, err
	}

	switch unit {
	case "B", "b", "":
		return size, nil
	case "KB", "kb", "K", "k":
		return size * 1024, nil
	case "MB", "mb", "M", "m":
		return size * 1024 * 1024, nil
	case "GB", "gb", "G", "g":
		return size * 1024 * 1024 * 1024, nil
	default:
		return 0, fmt.Errorf("unknown size unit: %s", unit)
	}
}

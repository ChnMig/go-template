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
		Port int `yaml:"port"`
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

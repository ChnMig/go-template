package config

import (
	"fmt"
	"net"
	"strings"
)

func (config Config) Validate() error {
	if strings.TrimSpace(config.Server.Host) == "" {
		return fmt.Errorf("server.host must not be empty")
	}
	if config.Server.Port < 1 || config.Server.Port > 65535 {
		return fmt.Errorf("server.port must be between 1 and 65535")
	}
	if strings.TrimSpace(config.Server.PIDFile) == "" {
		return fmt.Errorf("server.pid_file must not be empty")
	}
	if config.Server.ReadTimeout <= 0 || config.Server.WriteTimeout <= 0 ||
		config.Server.IdleTimeout <= 0 || config.Server.ShutdownTimeout <= 0 {
		return fmt.Errorf("server timeouts must be positive")
	}
	if config.Server.MaxHeaderBytes < 1024 || config.Server.MaxBodySize < 1024 {
		return fmt.Errorf("server header and body limits are too small")
	}
	if err := config.Server.HTTPConfig().Validate(); err != nil {
		return err
	}
	if err := config.Log.validate(); err != nil {
		return err
	}
	if config.Redis.Host != "" {
		if _, _, err := net.SplitHostPort(config.Redis.Host); err != nil {
			return fmt.Errorf("redis.host must be empty or a valid host:port: %w", err)
		}
	}
	if config.Redis.KeyPrefix != "" && !strings.HasSuffix(config.Redis.KeyPrefix, ":") {
		return fmt.Errorf("redis.key_prefix must be empty or end with ':'")
	}
	return ValidateConfig(config.JWT.Key, int64(config.JWT.Expiration))
}

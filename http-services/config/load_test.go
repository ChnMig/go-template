package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestParseSize(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    int64
		wantErr bool
	}{
		{"bytes", "100B", 100, false},
		{"kilobytes", "10KB", 10 * 1024, false},
		{"megabytes", "5MB", 5 * 1024 * 1024, false},
		{"gigabytes", "2GB", 2 * 1024 * 1024 * 1024, false},
		{"lowercase kb", "10kb", 10 * 1024, false},
		{"short form k", "10K", 10 * 1024, false},
		{"short form m", "5M", 5 * 1024 * 1024, false},
		{"short form g", "2G", 2 * 1024 * 1024 * 1024, false},
		{"invalid format", "invalid", 0, true},
		{"unknown unit", "10XB", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseSize(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseSize() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("parseSize() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSetDefaults(t *testing.T) {
	// 创建新的 viper 实例用于测试
	if err := LoadConfig(); err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}

	tests := []struct {
		name string
		key  string
		want any
	}{
		{"server port", "server.port", 8080},
		{"max body size", "server.max_body_size", "10MB"},
		{"pid file", "server.pid_file", "http-services.pid"},
		{"static dir", "server.static_dir", "static"},
		{"enable cors", "server.enable_cors", true},
		{"jwt expiration", "jwt.expiration", "12h"},
		{"log max size", "log.max_size", 50},
		{"enable rate limit", "server.enable_rate_limit", false},
		{"database mysql dsn", "database.mysql_dsn", ""},
		{"redis host", "redis.host", ""},
		{"redis password", "redis.password", ""},
		{"redis key prefix", "redis.key_prefix", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := v.Get(tt.key)
			if got != tt.want {
				t.Errorf("default %s = %v, want %v", tt.key, got, tt.want)
			}
		})
	}
	if got := v.GetStringSlice("server.trusted_proxies"); len(got) != 2 || got[0] != "127.0.0.1" || got[1] != "::1" {
		t.Fatalf("default server.trusted_proxies = %#v", got)
	}
}

func TestApplyConfig(t *testing.T) {
	// 初始化配置
	err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}

	// 检查全局变量是否正确设置
	if ListenPort != 8080 {
		t.Errorf("ListenPort = %d, want 8080", ListenPort)
	}

	if MaxBodySize != 10*1024*1024 {
		t.Errorf("MaxBodySize = %d, want %d", MaxBodySize, 10*1024*1024)
	}

	if JWTExpiration != 12*time.Hour {
		t.Errorf("JWTExpiration = %v, want %v", JWTExpiration, 12*time.Hour)
	}

	if filepath.Base(PidFile) != "http-services.pid" {
		t.Errorf("PidFile = %s, want base http-services.pid", PidFile)
	}

	if CurrentLogConfig().MaxSize != 50 {
		t.Errorf("log max size = %d, want 50", CurrentLogConfig().MaxSize)
	}
	if StaticDir != "static" {
		t.Errorf("StaticDir = %q, want static", StaticDir)
	}
	if len(TrustedProxies) != 2 || TrustedProxies[0] != "127.0.0.1" || TrustedProxies[1] != "::1" {
		t.Errorf("TrustedProxies = %#v", TrustedProxies)
	}
	if !EnableCORS {
		t.Error("EnableCORS = false, want true")
	}

	if MysqlDSN != "" {
		t.Errorf("MysqlDSN = %q, want empty string", MysqlDSN)
	}

	if RedisHost != "" {
		t.Errorf("RedisHost = %q, want empty string", RedisHost)
	}
}

func TestLoadConfigWithEnv(t *testing.T) {
	// 设置环境变量
	t.Setenv("HTTP_SERVICES_SERVER_PORT", "9090")
	t.Setenv("HTTP_SERVICES_JWT_EXPIRATION", "24h")
	t.Setenv("HTTP_SERVICES_DATABASE_MYSQL_DSN", "user:pass@tcp(127.0.0.1:3306)/app?charset=utf8mb4&parseTime=True&loc=Local")
	t.Setenv("HTTP_SERVICES_REDIS_HOST", "127.0.0.1:6380")
	t.Setenv("HTTP_SERVICES_SERVER_STATIC_DIR", "public")
	t.Setenv("HTTP_SERVICES_SERVER_TRUSTED_PROXIES", "10.0.0.0/8,192.0.2.10")
	t.Setenv("HTTP_SERVICES_SERVER_ENABLE_CORS", "false")
	pidPath := filepath.Join(t.TempDir(), "http-services.pid")
	t.Setenv("HTTP_SERVICES_SERVER_PID_FILE", pidPath)

	// 重新加载配置
	err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}

	// 验证环境变量覆盖
	if ListenPort != 9090 {
		t.Errorf("ListenPort = %d, want 9090 (from env)", ListenPort)
	}

	if JWTExpiration != 24*time.Hour {
		t.Errorf("JWTExpiration = %v, want 24h (from env)", JWTExpiration)
	}

	if PidFile != pidPath {
		t.Errorf("PidFile = %s, want %s (from env)", PidFile, pidPath)
	}

	if MysqlDSN != "user:pass@tcp(127.0.0.1:3306)/app?charset=utf8mb4&parseTime=True&loc=Local" {
		t.Errorf("MysqlDSN = %q, want env value", MysqlDSN)
	}

	if RedisHost != "127.0.0.1:6380" {
		t.Errorf("RedisHost = %q, want 127.0.0.1:6380 (from env)", RedisHost)
	}
	if StaticDir != "public" {
		t.Errorf("StaticDir = %q, want public", StaticDir)
	}
	if len(TrustedProxies) != 2 || TrustedProxies[0] != "10.0.0.0/8" || TrustedProxies[1] != "192.0.2.10" {
		t.Errorf("TrustedProxies = %#v, want env values", TrustedProxies)
	}
	if EnableCORS {
		t.Error("EnableCORS = true, want false from env")
	}
}

func TestLoadConfig_RedisKeyPrefix(t *testing.T) {
	tests := []struct {
		name     string
		yaml     string
		envValue string
		setEnv   bool
		want     string
	}{
		{name: "defaults to empty", want: ""},
		{name: "trims yaml value", yaml: "redis:\n  key_prefix: \" service:env: \"\n", want: "service:env:"},
		{name: "treats whitespace-only yaml value as empty", yaml: "redis:\n  key_prefix: \"   \"\n", want: ""},
		{name: "reads trimmed environment value", envValue: " service:env: ", setEnv: true, want: "service:env:"},
		{name: "environment overrides distinct yaml value", yaml: "redis:\n  key_prefix: \"yaml:\"\n", envValue: " env: ", setEnv: true, want: "env:"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			originalAbsPath := AbsPath
			originalPrefix := RedisKeyPrefix
			originalViper := v
			t.Cleanup(func() {
				AbsPath = originalAbsPath
				RedisKeyPrefix = originalPrefix
				v = originalViper
			})

			configDir := t.TempDir()
			AbsPath = configDir
			if tt.setEnv {
				t.Setenv("HTTP_SERVICES_REDIS_KEY_PREFIX", tt.envValue)
			} else {
				t.Setenv("HTTP_SERVICES_REDIS_KEY_PREFIX", "")
			}
			if tt.yaml != "" {
				if err := os.WriteFile(filepath.Join(configDir, "config.yaml"), []byte(tt.yaml), 0o600); err != nil {
					t.Fatalf("write config file: %v", err)
				}
			}

			if err := LoadConfig(); err != nil {
				t.Fatalf("LoadConfig() error = %v", err)
			}
			if RedisKeyPrefix != tt.want {
				t.Errorf("RedisKeyPrefix = %q, want %q", RedisKeyPrefix, tt.want)
			}
		})
	}
}

func TestLoadConfigRejectsInvalidTrustedProxy(t *testing.T) {
	originalAbsPath := AbsPath
	originalViper := v
	t.Cleanup(func() {
		AbsPath = originalAbsPath
		v = originalViper
	})

	configDir := t.TempDir()
	AbsPath = configDir
	content := []byte("server:\n  trusted_proxies:\n    - not-a-proxy\n")
	if err := os.WriteFile(filepath.Join(configDir, "config.yaml"), content, 0o600); err != nil {
		t.Fatalf("write config file: %v", err)
	}

	if err := LoadConfig(); err == nil {
		t.Fatal("LoadConfig() error = nil, want invalid trusted proxy error")
	}
}

func TestLoadConfigRejectsUnsafeHTTPInfrastructureConfig(t *testing.T) {
	tests := []struct {
		name string
		yaml string
	}{
		{name: "working directory as static root", yaml: "server:\n  static_dir: .\n"},
		{name: "filesystem root as static root", yaml: "server:\n  static_dir: /\n"},
		{name: "non-positive enabled rate", yaml: "server:\n  enable_rate_limit: true\n  global_rate_limit: 0\n"},
		{name: "non-positive enabled burst", yaml: "server:\n  enable_rate_limit: true\n  global_rate_burst: 0\n"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			originalAbsPath := AbsPath
			originalViper := v
			t.Cleanup(func() {
				AbsPath = originalAbsPath
				v = originalViper
			})

			configDir := t.TempDir()
			AbsPath = configDir
			if err := os.WriteFile(filepath.Join(configDir, "config.yaml"), []byte(test.yaml), 0o600); err != nil {
				t.Fatalf("write config file: %v", err)
			}
			if err := LoadConfig(); err == nil {
				t.Fatal("LoadConfig() error = nil, want unsafe HTTP infrastructure config error")
			}
		})
	}
}

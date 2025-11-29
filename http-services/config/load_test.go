package config

import (
	"os"
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
	LoadConfig() // 初始化 v

	tests := []struct {
		name string
		key  string
		want interface{}
	}{
		{"server port", "server.port", 8080},
		{"max body size", "server.max_body_size", "10MB"},
		{"jwt expiration", "jwt.expiration", "12h"},
		{"log max size", "log.max_size", 50},
		{"log max backups", "log.max_backups", 3},
		{"enable rate limit", "server.enable_rate_limit", false},
		{"enable acme", "server.enable_acme", false},
		{"acme domain", "server.acme_domain", ""},
		{"acme cache dir", "server.acme_cache_dir", "acme-cert-cache"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := v.Get(tt.key)
			if got != tt.want {
				t.Errorf("default %s = %v, want %v", tt.key, got, tt.want)
			}
		})
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

	if LogMaxSize != 50 {
		t.Errorf("LogMaxSize = %d, want 50", LogMaxSize)
	}

	if EnableACME {
		t.Errorf("EnableACME = %v, want false", EnableACME)
	}

	if ACMEDomain != "" {
		t.Errorf("ACMEDomain = %s, want empty", ACMEDomain)
	}

	if ACMECacheDir == "" {
		t.Errorf("ACMECacheDir is empty, want non-empty default value")
	}
}

func TestLoadConfigWithEnv(t *testing.T) {
	// 设置环境变量
	os.Setenv("HTTP_SERVICES_SERVER_PORT", "9090")
	os.Setenv("HTTP_SERVICES_JWT_EXPIRATION", "24h")
	os.Setenv("HTTP_SERVICES_SERVER_ENABLE_ACME", "true")
	os.Setenv("HTTP_SERVICES_SERVER_ACME_DOMAIN", "api.example.com")
	os.Setenv("HTTP_SERVICES_SERVER_ACME_CACHE_DIR", "/tmp/acme-cache")
	defer func() {
		os.Unsetenv("HTTP_SERVICES_SERVER_PORT")
		os.Unsetenv("HTTP_SERVICES_JWT_EXPIRATION")
		os.Unsetenv("HTTP_SERVICES_SERVER_ENABLE_ACME")
		os.Unsetenv("HTTP_SERVICES_SERVER_ACME_DOMAIN")
		os.Unsetenv("HTTP_SERVICES_SERVER_ACME_CACHE_DIR")
	}()

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

	if !EnableACME {
		t.Errorf("EnableACME = %v, want true (from env)", EnableACME)
	}

	if ACMEDomain != "api.example.com" {
		t.Errorf("ACMEDomain = %s, want api.example.com (from env)", ACMEDomain)
	}

	if ACMECacheDir != "/tmp/acme-cache" {
		t.Errorf("ACMECacheDir = %s, want /tmp/acme-cache (from env)", ACMECacheDir)
	}
}

func TestGetViper(t *testing.T) {
	LoadConfig()
	viper := GetViper()
	if viper == nil {
		t.Error("GetViper() returned nil")
	}
	if viper != v {
		t.Error("GetViper() did not return the expected viper instance")
	}
}

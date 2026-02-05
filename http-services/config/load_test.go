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
	LoadConfig() // 初始化 v

	tests := []struct {
		name string
		key  string
		want interface{}
	}{
		{"server port", "server.port", 8080},
		{"max body size", "server.max_body_size", "10MB"},
		{"pid file", "server.pid_file", "http-services.pid"},
		{"jwt expiration", "jwt.expiration", "12h"},
		{"log max size", "log.max_size", 50},
		{"enable rate limit", "server.enable_rate_limit", false},
		{"enable acme", "server.enable_acme", false},
		{"acme domain", "server.acme_domain", ""},
		{"acme cache dir", "server.acme_cache_dir", "acme-cert-cache"},
		{"enable tls", "server.enable_tls", false},
		{"tls cert file", "server.tls_cert_file", ""},
		{"tls key file", "server.tls_key_file", ""},
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

	if filepath.Base(PidFile) != "http-services.pid" {
		t.Errorf("PidFile = %s, want base http-services.pid", PidFile)
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

	if EnableTLS {
		t.Errorf("EnableTLS = %v, want false", EnableTLS)
	}

	if TLSCertFile != "" {
		t.Errorf("TLSCertFile = %s, want empty", TLSCertFile)
	}

	if TLSKeyFile != "" {
		t.Errorf("TLSKeyFile = %s, want empty", TLSKeyFile)
	}
}

func TestLoadConfigWithEnv(t *testing.T) {
	// 设置环境变量
	os.Setenv("HTTP_SERVICES_SERVER_PORT", "9090")
	os.Setenv("HTTP_SERVICES_JWT_EXPIRATION", "24h")
	pidPath := filepath.Join(t.TempDir(), "http-services.pid")
	os.Setenv("HTTP_SERVICES_SERVER_PID_FILE", pidPath)
	os.Setenv("HTTP_SERVICES_SERVER_ENABLE_ACME", "true")
	os.Setenv("HTTP_SERVICES_SERVER_ACME_DOMAIN", "api.example.com")
	os.Setenv("HTTP_SERVICES_SERVER_ACME_CACHE_DIR", "/tmp/acme-cache")
	os.Setenv("HTTP_SERVICES_SERVER_ENABLE_TLS", "true")
	os.Setenv("HTTP_SERVICES_SERVER_TLS_CERT_FILE", "/tmp/server.crt")
	os.Setenv("HTTP_SERVICES_SERVER_TLS_KEY_FILE", "/tmp/server.key")
	defer func() {
		os.Unsetenv("HTTP_SERVICES_SERVER_PORT")
		os.Unsetenv("HTTP_SERVICES_JWT_EXPIRATION")
		os.Unsetenv("HTTP_SERVICES_SERVER_PID_FILE")
		os.Unsetenv("HTTP_SERVICES_SERVER_ENABLE_ACME")
		os.Unsetenv("HTTP_SERVICES_SERVER_ACME_DOMAIN")
		os.Unsetenv("HTTP_SERVICES_SERVER_ACME_CACHE_DIR")
		os.Unsetenv("HTTP_SERVICES_SERVER_ENABLE_TLS")
		os.Unsetenv("HTTP_SERVICES_SERVER_TLS_CERT_FILE")
		os.Unsetenv("HTTP_SERVICES_SERVER_TLS_KEY_FILE")
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

	if PidFile != pidPath {
		t.Errorf("PidFile = %s, want %s (from env)", PidFile, pidPath)
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

	if !EnableTLS {
		t.Errorf("EnableTLS = %v, want true (from env)", EnableTLS)
	}

	if TLSCertFile != "/tmp/server.crt" {
		t.Errorf("TLSCertFile = %s, want /tmp/server.crt (from env)", TLSCertFile)
	}

	if TLSKeyFile != "/tmp/server.key" {
		t.Errorf("TLSKeyFile = %s, want /tmp/server.key (from env)", TLSKeyFile)
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

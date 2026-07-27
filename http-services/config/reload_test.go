package config

import (
	"sync"
	"testing"

	"github.com/spf13/viper"
)

func TestApplyReloadableLogConfigKeepsStartupConfigurationFrozen(t *testing.T) {
	originalViper := v
	originalPort := ListenPort
	originalJWTKey := JWTKey
	originalRuntimeLog := CurrentLogConfig()
	t.Cleanup(func() {
		v = originalViper
		ListenPort = originalPort
		JWTKey = originalJWTKey
		UpdateLogConfig(originalRuntimeLog)
	})

	v = viper.New()
	setDefaults()
	ListenPort = 8080
	JWTKey = "startup-jwt-key"
	UpdateLogConfig(LogConfig{MaxSize: 50, MaxAge: 30, Level: "info", GinLevel: "info"})
	v.Set("server.port", 9090)
	v.Set("jwt.key", "reloaded-jwt-key")
	v.Set("log.max_size", 100)
	v.Set("log.max_age", 7)
	v.Set("log.level", "debug")
	v.Set("log.gin_level", "warn")

	applyReloadableLogConfig()

	if ListenPort != 8080 || JWTKey != "startup-jwt-key" {
		t.Fatalf("startup config changed: port=%d jwt=%q", ListenPort, JWTKey)
	}
	got := CurrentLogConfig()
	want := (LogConfig{MaxSize: 100, MaxAge: 7, Level: "debug", GinLevel: "warn"})
	if got != want {
		t.Fatalf("runtime log config = %#v, want %#v", got, want)
	}
}

func TestLogConfigSnapshotSupportsConcurrentReloadAndRead(t *testing.T) {
	original := CurrentLogConfig()
	t.Cleanup(func() { UpdateLogConfig(original) })

	var waitGroup sync.WaitGroup
	waitGroup.Add(2)
	go func() {
		defer waitGroup.Done()
		for index := range 100 {
			UpdateLogConfig(LogConfig{MaxSize: LogFileSizeMB(index + 1), MaxAge: 30, Level: "info"})
		}
	}()
	go func() {
		defer waitGroup.Done()
		for range 100 {
			_ = CurrentLogConfig()
		}
	}()
	waitGroup.Wait()
}

func TestSnapshotHTTPConfigCopiesTrustedProxies(t *testing.T) {
	original := append([]string(nil), TrustedProxies...)
	t.Cleanup(func() { TrustedProxies = original })
	TrustedProxies = []string{"127.0.0.1", "::1"}

	snapshot := SnapshotHTTPConfig()
	TrustedProxies[0] = "192.0.2.1"

	if snapshot.TrustedProxies[0] != "127.0.0.1" {
		t.Fatalf("snapshot trusted proxies changed: %#v", snapshot.TrustedProxies)
	}
}

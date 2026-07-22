package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
)

func TestStopWatchConfigWaitsForWatcherAndAllowsRestart(t *testing.T) {
	originalViper := v
	t.Cleanup(func() {
		StopWatchConfig()
		v = originalViper
	})

	configPath := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(configPath, []byte("log:\n  level: info\n"), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}
	v = viper.New()
	v.SetConfigFile(configPath)
	if err := v.ReadInConfig(); err != nil {
		t.Fatalf("read config: %v", err)
	}

	for iteration := range 2 {
		if err := WatchConfig(nil); err != nil {
			t.Fatalf("start watcher %d: %v", iteration, err)
		}
		watchConfigMu.Lock()
		running := activeConfigWatcher
		watchConfigMu.Unlock()
		if running == nil {
			t.Fatalf("watcher %d was not registered", iteration)
		}

		StopWatchConfig()
		select {
		case <-running.stopped:
		default:
			t.Fatalf("watcher %d was not joined", iteration)
		}
	}
}

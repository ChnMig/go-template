package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_LogWatcher_reloads_only_valid_log_snapshot(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	require.NoError(t, os.WriteFile(path, []byte("log:\n  level: info\n  gin_level: warn\n  max_size: 50\n  max_age: 30\n"), 0o600))
	var received LogConfig
	watcher := &LogWatcher{configFile: path, onChange: func(next LogConfig) error {
		received = next
		return nil
	}}
	require.NoError(t, os.WriteFile(path, []byte("server:\n  port: 1\nlog:\n  level: debug\n  gin_level: error\n  max_size: 100\n  max_age: 60\n"), 0o600))

	err := watcher.reload()

	require.NoError(t, err)
	require.Equal(t, LogLevelDebug, received.Level)
	require.Equal(t, LogLevelError, received.GinLevel)
	require.Equal(t, LogFileSizeMB(100), received.MaxSize)
	require.Equal(t, LogRetentionDays(60), received.MaxAge)
}

func Test_LogWatcher_rejects_invalid_log_reload(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	require.NoError(t, os.WriteFile(path, []byte("log:\n  level: verbose\n  max_size: 50\n  max_age: 30\n"), 0o600))
	called := false
	watcher := &LogWatcher{configFile: path, onChange: func(LogConfig) error {
		called = true
		return nil
	}}

	err := watcher.reload()

	require.Error(t, err)
	require.False(t, called)
}

func Test_LogWatcher_close_joins_and_allows_independent_restart(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	require.NoError(t, os.WriteFile(path, []byte("log:\n  level: info\n"), 0o600))
	initial := LogConfig{Level: "info", MaxSize: 50, MaxAge: 30, sourcePath: path}

	for range 2 {
		watcher, err := WatchLogConfig(initial, nil)
		require.NoError(t, err)
		require.NotNil(t, watcher.watcher)

		require.NoError(t, watcher.Close())
		select {
		case <-watcher.stopped:
		default:
			t.Fatal("watcher was not joined")
		}
	}
}

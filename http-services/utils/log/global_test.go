package log

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing"

	"http-services/config"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func Test_SetLogger_development_replaces_the_global_zap_logger(t *testing.T) {
	// Given
	var output bytes.Buffer
	cfg := testLogConfig()

	// When
	err := SetLogger(cfg, true, &output)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, Close()) })
	zap.L().Info("business event")
	GetGinLogger().Info("Gin event")

	// Then
	require.Same(t, GetLogger(), zap.L())
	require.Contains(t, output.String(), "business event")
	require.Contains(t, output.String(), "Gin event")
}

func Test_SetLogger_release_writes_program_named_business_and_Gin_files(t *testing.T) {
	// Given
	directory := t.TempDir()
	t.Chdir(directory)
	originalArgZero := os.Args[0]
	os.Args[0] = filepath.Join(directory, "http-services-qa")
	t.Cleanup(func() { os.Args[0] = originalArgZero })

	// When
	err := SetLogger(testLogConfig(), false, io.Discard)
	require.NoError(t, err)
	zap.L().Info("business event")
	GetGinLogger().Info("Gin event")
	require.NoError(t, Close())

	// Then
	business, err := os.ReadFile(filepath.Join(directory, "log", "http-services-qa.log"))
	require.NoError(t, err)
	ginOutput, err := os.ReadFile(filepath.Join(directory, "log", "http-services-qa.gin.log"))
	require.NoError(t, err)
	require.Contains(t, string(business), "business event")
	require.NotContains(t, string(business), "Gin event")
	require.Contains(t, string(ginOutput), "Gin event")
	require.NotContains(t, string(ginOutput), "business event")
}

func Test_SetLogger_uses_independent_Gin_level_and_business_fallback(t *testing.T) {
	for _, testCase := range []struct {
		name       string
		business   config.LogLevel
		gin        config.LogLevel
		ginInfo    bool
		businessOK bool
	}{
		{name: "independent Gin level", business: config.LogLevelWarn, gin: config.LogLevelInfo, ginInfo: true},
		{name: "empty Gin level follows business", business: config.LogLevelError},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			// Given
			cfg := testLogConfig()
			cfg.Level = testCase.business
			cfg.GinLevel = testCase.gin

			// When
			err := SetLogger(cfg, true, io.Discard)
			require.NoError(t, err)
			t.Cleanup(func() { require.NoError(t, Close()) })

			// Then
			require.Equal(t, testCase.businessOK, GetLogger().Check(zap.InfoLevel, "business info") != nil)
			require.Equal(t, testCase.ginInfo, GetGinLogger().Check(zap.InfoLevel, "Gin info") != nil)
		})
	}
}

func Test_FromContext_adds_trace_ID_to_the_global_business_logger(t *testing.T) {
	// Given
	var output bytes.Buffer
	require.NoError(t, SetLogger(testLogConfig(), true, &output))
	t.Cleanup(func() { require.NoError(t, Close()) })
	ctx := WithTraceID(t.Context(), "trace-123")

	// When
	FromStandardContext(ctx).Info("correlated event")

	// Then
	require.Contains(t, output.String(), "trace_id")
	require.Contains(t, output.String(), "trace-123")
}

func Test_Close_joins_the_release_monitor_and_releases_global_loggers(t *testing.T) {
	// Given
	t.Chdir(t.TempDir())
	require.NoError(t, SetLogger(testLogConfig(), false, io.Discard))
	StartMonitor()
	monitorLifecycleMu.Lock()
	running := activeMonitor
	monitorLifecycleMu.Unlock()
	require.NotNil(t, running)

	// When
	err := Close()

	// Then
	require.NoError(t, err)
	monitorLifecycleMu.Lock()
	require.Nil(t, activeMonitor)
	monitorLifecycleMu.Unlock()
	loggerMu.RLock()
	defer loggerMu.RUnlock()
	require.Nil(t, logger)
	require.Nil(t, ginLogger)
	require.Nil(t, loggerCore)
	require.Nil(t, ginLoggerCore)
}

func testLogConfig() config.LogConfig {
	return config.LogConfig{
		MaxSize: config.LogFileSizeMB(1), MaxAge: config.LogRetentionDays(1),
		Level: config.LogLevelInfo, GinLevel: config.LogLevelInfo,
	}
}

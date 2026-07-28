package log

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"http-services/config"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func TestSetLogger_UsesIndependentGinLogLevel(t *testing.T) {
	tempDir := t.TempDir()
	oldRunModel := config.RunModel
	oldLogDir := config.LogDir
	oldLogPath := config.LogPath
	oldSelfName := config.SelfName
	oldLogMaxSize := config.LogMaxSize
	oldLogMaxAge := config.LogMaxAge
	oldLogLevel := config.LogLevel
	oldGinLogLevel := config.GinLogLevel
	t.Cleanup(func() {
		config.RunModel = oldRunModel
		config.LogDir = oldLogDir
		config.LogPath = oldLogPath
		config.SelfName = oldSelfName
		config.LogMaxSize = oldLogMaxSize
		config.LogMaxAge = oldLogMaxAge
		config.LogLevel = oldLogLevel
		config.GinLogLevel = oldGinLogLevel
		SetLogger()
	})

	config.RunModel = config.RunModelRelease
	config.LogDir = tempDir
	config.SelfName = "http-services-test"
	config.LogPath = filepath.Join(tempDir, "app.log")
	config.LogMaxSize = 1
	config.LogMaxAge = 1
	config.LogLevel = "warn"
	config.GinLogLevel = "info"

	SetLogger()

	if ce := GetLogger().Check(zap.InfoLevel, "business info"); ce != nil {
		t.Fatal("expected business logger to suppress info level")
	}
	if ce := GetLogger().Check(zap.WarnLevel, "business warn"); ce == nil {
		t.Fatal("expected business logger to allow warn level")
	}
	if ce := GetGinLogger().Check(zap.InfoLevel, "gin info"); ce == nil {
		t.Fatal("expected gin logger to allow info level")
	}
	if ce := GetGinErrorLogger().Check(zapcore.ErrorLevel, "gin error"); ce == nil {
		t.Fatal("expected gin error logger to allow error level")
	}
}

func TestSetLogger_EmptyGinLogLevelFallsBackToBusinessLevel(t *testing.T) {
	tempDir := t.TempDir()
	oldRunModel := config.RunModel
	oldLogDir := config.LogDir
	oldLogPath := config.LogPath
	oldSelfName := config.SelfName
	oldLogMaxSize := config.LogMaxSize
	oldLogMaxAge := config.LogMaxAge
	oldLogLevel := config.LogLevel
	oldGinLogLevel := config.GinLogLevel
	t.Cleanup(func() {
		config.RunModel = oldRunModel
		config.LogDir = oldLogDir
		config.LogPath = oldLogPath
		config.SelfName = oldSelfName
		config.LogMaxSize = oldLogMaxSize
		config.LogMaxAge = oldLogMaxAge
		config.LogLevel = oldLogLevel
		config.GinLogLevel = oldGinLogLevel
		SetLogger()
	})

	config.RunModel = config.RunModelRelease
	config.LogDir = tempDir
	config.SelfName = "http-services-test"
	config.LogPath = filepath.Join(tempDir, "app.log")
	config.LogMaxSize = 1
	config.LogMaxAge = 1
	config.LogLevel = "error"
	config.GinLogLevel = ""

	SetLogger()

	if ce := GetGinLogger().Check(zap.InfoLevel, "gin info"); ce != nil {
		t.Fatal("expected gin logger to follow business level when gin level is empty")
	}
	if ce := GetGinLogger().Check(zap.ErrorLevel, "gin error"); ce == nil {
		t.Fatal("expected gin logger to allow error level")
	}
}

func TestSetLogger_ReleaseWritesSeparateBusinessAndGinFiles(t *testing.T) {
	tempDir := t.TempDir()
	oldRunModel := config.RunModel
	oldLogDir := config.LogDir
	oldLogPath := config.LogPath
	oldSelfName := config.SelfName
	oldLogMaxSize := config.LogMaxSize
	oldLogMaxAge := config.LogMaxAge
	oldLogLevel := config.LogLevel
	oldGinLogLevel := config.GinLogLevel
	t.Cleanup(func() {
		config.RunModel = oldRunModel
		config.LogDir = oldLogDir
		config.LogPath = oldLogPath
		config.SelfName = oldSelfName
		config.LogMaxSize = oldLogMaxSize
		config.LogMaxAge = oldLogMaxAge
		config.LogLevel = oldLogLevel
		config.GinLogLevel = oldGinLogLevel
		SetLogger()
	})

	config.RunModel = config.RunModelRelease
	config.LogDir = tempDir
	config.SelfName = "http-services-test"
	config.LogPath = filepath.Join(tempDir, "http-services-test.log")
	config.LogMaxSize = 1
	config.LogMaxAge = 1
	config.LogLevel = "info"
	config.GinLogLevel = "info"
	SetLogger()

	zap.L().Info("business event")
	GetGinLogger().Info("gin event")
	if err := zap.L().Sync(); err != nil {
		t.Fatalf("sync business logger: %v", err)
	}
	if err := GetGinLogger().Sync(); err != nil {
		t.Fatalf("sync Gin logger: %v", err)
	}

	business, err := os.ReadFile(config.LogPath)
	if err != nil {
		t.Fatalf("read business log: %v", err)
	}
	ginOutput, err := os.ReadFile(filepath.Join(tempDir, "http-services-test.gin.log"))
	if err != nil {
		t.Fatalf("read Gin log: %v", err)
	}
	if !strings.Contains(string(business), "business event") || strings.Contains(string(business), "gin event") {
		t.Fatalf("business log content = %q", business)
	}
	if !strings.Contains(string(ginOutput), "gin event") || strings.Contains(string(ginOutput), "business event") {
		t.Fatalf("Gin log content = %q", ginOutput)
	}
}

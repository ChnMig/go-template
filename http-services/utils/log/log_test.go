package log

import (
	"path/filepath"
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

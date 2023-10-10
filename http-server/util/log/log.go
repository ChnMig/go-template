package log

import (
	"os"
	"time"

	"http-server/config"

	"github.com/fsnotify/fsnotify"
	"github.com/natefinch/lumberjack"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var logger *zap.Logger

// config
const (
	envKey         = "log_model"
	modelDevValue  = "dev"
	fileMaxSize    = 1  // M
	fileMaxBackups = 10 // backups
	fileMaxAge     = 30 // days
)

// Creating Dev logger
// DEV mode outputs logs to the terminal and is more readable
func createDevLogger(fileName string) *zap.Logger {
	encoder := zap.NewDevelopmentEncoderConfig()
	core := zapcore.NewTee(
		zapcore.NewSamplerWithOptions(
			zapcore.NewCore(zapcore.NewConsoleEncoder(encoder), os.Stdout, zap.DebugLevel), time.Second, 4, 1),
	)
	return zap.New(core, zap.AddCaller())
}

// Creating product logger
// The product pattern outputs logs to a file and is architecturally structured, in json format.
func createProductLogger(fileName string) *zap.Logger {
	fileEncoder := zap.NewProductionEncoderConfig()
	fileEncoder.EncodeTime = zapcore.ISO8601TimeEncoder
	fileWriter := zapcore.AddSync(&lumberjack.Logger{
		Filename:   fileName,
		MaxSize:    fileMaxSize,
		MaxBackups: fileMaxBackups,
		MaxAge:     fileMaxAge,
	})
	core := zapcore.NewTee(
		zapcore.NewSamplerWithOptions(
			zapcore.NewCore(zapcore.NewJSONEncoder(fileEncoder), fileWriter, zap.InfoLevel), time.Second, 4, 1),
	)
	return zap.New(core, zap.AddCaller())
}

// Reset logger to prevent zap persistence problems after files are deleted
func ResetLogger() {
	model := os.Getenv(envKey)
	if model == modelDevValue {
		logger = createDevLogger(config.LogPath)
	} else {
		logger = createProductLogger(config.LogPath)
	}
	zap.ReplaceGlobals(logger)
}

// Listen to log files
func MonitorFile() {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		zap.L().Error("File listening error", zap.Error(err))
		return
	}
	defer watcher.Close()
	err = watcher.Add(config.LogPath)
	if err != nil {
		zap.L().Error("File listening error", zap.Error(err))
	}
	for {
		select {
		case event := <-watcher.Events:
			if event.Has(fsnotify.Remove) {
				zap.L().Warn("The log file was deleted")
				ResetLogger()
			}
			if event.Has(fsnotify.Rename) {
				zap.L().Warn("Log files are renamed and new files are monitored")
				ResetLogger()
			}
		case err := <-watcher.Errors:
			zap.L().Error("File listening error", zap.Error(err))
		}
	}
}

func GetLogger() *zap.Logger {
	return logger
}

func init() {
	// Get log mode
	model := os.Getenv(envKey)
	if model == modelDevValue {
		logger = createDevLogger(config.LogPath)
	} else {
		logger = createProductLogger(config.LogPath)
	}
	zap.ReplaceGlobals(logger)
	go MonitorFile()
}

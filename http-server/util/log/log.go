package log

import (
	"os"
	"time"

	"http-server/config"

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

func GetLogger() *zap.Logger {
	return logger
}

func init() {
	// Get log mode
	model := os.Getenv(envKey)
	if model == modelDevValue {
		logger = createDevLogger(config.SelfLogPath)
	} else {
		logger = createProductLogger(config.SelfLogPath)
	}
	zap.ReplaceGlobals(logger)
}

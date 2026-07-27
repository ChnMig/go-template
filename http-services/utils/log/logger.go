// Package log owns the process-global business and Gin zap loggers.
package log

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"http-services/config"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	configureMu sync.Mutex
	loggerMu    sync.RWMutex

	logger          *zap.Logger
	loggerCore      *managedCore
	ginLogger       *zap.Logger
	ginErrorLogger  *zap.Logger
	ginLoggerCore   *managedCore
	currentSettings *settings
)

type settings struct {
	output      io.Writer
	logDir      string
	logPath     string
	ginLogPath  string
	cfg         config.LogConfig
	development bool
}

// SetLogger installs process-global business and Gin loggers using the example-project layout.
func SetLogger(cfg config.LogConfig, development bool, output io.Writer) error {
	next, err := newSettings(cfg, development, output)
	if err != nil {
		return err
	}
	return applySettings(next)
}

func newSettings(cfg config.LogConfig, development bool, output io.Writer) (settings, error) {
	if output == nil {
		output = os.Stdout
	}
	directory, err := os.Getwd()
	if err != nil {
		return settings{}, fmt.Errorf("resolve log working directory: %w", err)
	}
	programName := filepath.Base(os.Args[0])
	if strings.TrimSpace(programName) == "" || programName == "." {
		return settings{}, errors.New("resolve log program name: executable name is empty")
	}
	logDir := filepath.Join(directory, "log")
	return settings{
		cfg: cfg, development: development, output: output, logDir: logDir,
		logPath:    filepath.Join(logDir, programName+".log"),
		ginLogPath: filepath.Join(logDir, programName+".gin.log"),
	}, nil
}

func applySettings(next settings) error {
	configureMu.Lock()
	defer configureMu.Unlock()

	if !next.development {
		if err := os.MkdirAll(next.logDir, 0o750); err != nil {
			return fmt.Errorf("create log directory %s: %w", next.logDir, err)
		}
	}
	businessLevel := parseLogLevel(next.cfg.Level)
	ginLevel := businessLevel
	if strings.TrimSpace(string(next.cfg.GinLevel)) != "" {
		ginLevel = parseLogLevel(next.cfg.GinLevel)
	}
	businessCore, businessCloser := createCore(next, next.logPath, businessLevel)
	ginCore, ginCloser := createCore(next, next.ginLogPath, ginLevel)

	loggerMu.Lock()
	global, _, _, businessManaged, ginManaged := ensureLoggersLocked()
	copyOfNext := next
	currentSettings = &copyOfNext
	loggerMu.Unlock()

	businessErr := businessManaged.replace(businessCore, businessCloser)
	ginErr := ginManaged.replace(ginCore, ginCloser)
	zap.ReplaceGlobals(global)
	return errors.Join(businessErr, ginErr)
}

func createCore(value settings, path string, level zapcore.Level) (zapcore.Core, io.Closer) {
	if value.development {
		encoder := zap.NewDevelopmentEncoderConfig()
		return sampledCore(zapcore.NewConsoleEncoder(encoder), zapcore.AddSync(value.output), level),
			newFlushCloser(value.output)
	}
	encoder := zap.NewProductionEncoderConfig()
	encoder.EncodeTime = zapcore.ISO8601TimeEncoder
	writer := &lumberjack.Logger{
		Filename: path, MaxSize: int(value.cfg.MaxSize), MaxBackups: 0,
		MaxAge: int(value.cfg.MaxAge), LocalTime: true,
	}
	return sampledCore(zapcore.NewJSONEncoder(encoder), zapcore.AddSync(writer), level), writer
}

func sampledCore(encoder zapcore.Encoder, writer zapcore.WriteSyncer, level zapcore.Level) zapcore.Core {
	return zapcore.NewSamplerWithOptions(zapcore.NewCore(encoder, writer, level), time.Second, 4, 1)
}

func ensureLoggersLocked() (*zap.Logger, *zap.Logger, *zap.Logger, *managedCore, *managedCore) {
	if logger != nil && ginLogger != nil && ginErrorLogger != nil && loggerCore != nil && ginLoggerCore != nil {
		return logger, ginLogger, ginErrorLogger, loggerCore, ginLoggerCore
	}
	businessCore := newManagedCore()
	accessCore := newManagedCore()
	businessLogger := zap.New(businessCore, zap.AddCaller())
	accessLogger := zap.New(accessCore, zap.AddCaller()).With(zap.String("logger", "gin"))
	errorLogger := accessLogger.With(zap.String("stream", "stderr"))
	logger = businessLogger
	loggerCore = businessCore
	ginLogger = accessLogger
	ginErrorLogger = errorLogger
	ginLoggerCore = accessCore
	return businessLogger, accessLogger, errorLogger, businessCore, accessCore
}

// GetLogger returns the current process-global business logger.
func GetLogger() *zap.Logger {
	loggerMu.Lock()
	defer loggerMu.Unlock()
	businessLogger, _, _, _, _ := ensureLoggersLocked()
	return businessLogger
}

// GetGinLogger returns the current Gin access logger.
func GetGinLogger() *zap.Logger {
	loggerMu.Lock()
	defer loggerMu.Unlock()
	_, accessLogger, _, _, _ := ensureLoggersLocked()
	return accessLogger
}

// GetGinErrorLogger returns the current Gin error and recovery logger.
func GetGinErrorLogger() *zap.Logger {
	loggerMu.Lock()
	defer loggerMu.Unlock()
	_, _, errorLogger, _, _ := ensureLoggersLocked()
	return errorLogger
}

func parseLogLevel(level config.LogLevel) zapcore.Level {
	switch strings.ToLower(strings.TrimSpace(string(level))) {
	case "debug":
		return zap.DebugLevel
	case "warn", "warning":
		return zap.WarnLevel
	case "error":
		return zap.ErrorLevel
	default:
		return zap.InfoLevel
	}
}

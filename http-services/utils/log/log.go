package log

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"http-services/config"
	"http-services/utils/runmodel"

	"github.com/fsnotify/fsnotify"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	mu sync.RWMutex

	logger   *zap.Logger
	loggerLJ *lumberjack.Logger

	ginLogger      *zap.Logger
	ginErrorLogger *zap.Logger
	ginLoggerLJ    *lumberjack.Logger

	monitorDone chan struct{} // 用于停止监控 goroutine
	rotateDone  chan struct{} // 用于停止按天 Rotate goroutine
)

// Creating Dev logger
// DEV mode outputs logs to the terminal and is more readable
func createDevLogger(level zapcore.Level) *zap.Logger {
	encoder := zap.NewDevelopmentEncoderConfig()
	core := zapcore.NewTee(
		zapcore.NewSamplerWithOptions(
			zapcore.NewCore(zapcore.NewConsoleEncoder(encoder), os.Stdout, level), time.Second, 4, 1),
	)
	return zap.New(core, zap.AddCaller())
}

// Creating product logger
// The product pattern outputs logs to a file and is architecturally structured, in json format.
func createProductLogger(fileName string, level zapcore.Level) (*zap.Logger, *lumberjack.Logger) {
	fileEncoder := zap.NewProductionEncoderConfig()
	fileEncoder.EncodeTime = zapcore.ISO8601TimeEncoder
	lj := &lumberjack.Logger{
		Filename: fileName,
		MaxSize:  config.LogMaxSize,
		// 不限制备份文件数量：只按 max_age 做清理（并保留 max_size 兜底轮转）。
		MaxBackups: 0,
		MaxAge:     config.LogMaxAge,
		LocalTime:  true,
	}
	fileWriter := zapcore.AddSync(lj)
	core := zapcore.NewTee(
		zapcore.NewSamplerWithOptions(
			zapcore.NewCore(zapcore.NewJSONEncoder(fileEncoder), fileWriter, level), time.Second, 4, 1),
	)
	return zap.New(core, zap.AddCaller()), lj
}

// SetLogger to prevent zap persistence problems after files are deleted
func SetLogger() {
	mu.Lock()
	defer mu.Unlock()
	businessLevel := parseLogLevel(config.LogLevel)
	ginLevel := businessLevel
	if strings.TrimSpace(config.GinLogLevel) != "" {
		ginLevel = parseLogLevel(config.GinLogLevel)
	}

	// Get log mode
	switch {
	case runmodel.IsDev():
		logger = createDevLogger(businessLevel)
		loggerLJ = nil

		ginLogger = createDevLogger(ginLevel).With(zap.String("logger", "gin"))
		ginErrorLogger = ginLogger.With(zap.String("stream", "stderr"))
		ginLoggerLJ = nil
	case runmodel.IsRelease():
		logger, loggerLJ = createProductLogger(config.LogPath, businessLevel)

		ginLogger, ginLoggerLJ = createProductLogger(ginLogPath(), ginLevel)
		ginLogger = ginLogger.With(zap.String("logger", "gin"))
		ginErrorLogger = ginLogger.With(zap.String("stream", "stderr"))
	default:
		// 默认视作开发模式，避免测试/包初始化阶段创建文件与目录
		logger = createDevLogger(businessLevel)
		loggerLJ = nil

		ginLogger = createDevLogger(ginLevel).With(zap.String("logger", "gin"))
		ginErrorLogger = ginLogger.With(zap.String("stream", "stderr"))
		ginLoggerLJ = nil
	}
	zap.ReplaceGlobals(logger)
}

// Listen to log files
// When the log file is deleted manually, we will automatically create a new one.
func monitorFile(done <-chan struct{}) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		zap.L().Error("File listening error", zap.Error(err))
		return
	}
	defer func() {
		if closeErr := watcher.Close(); closeErr != nil {
			zap.L().Warn("close log file watcher failed", zap.Error(closeErr))
		}
	}()
	// 监控日志目录，避免因日志文件轮转（rename）导致 watcher 失效。
	err = watcher.Add(config.LogDir)
	if err != nil {
		zap.L().Error("File listening error", zap.Error(err))
	}
	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			if !event.Has(fsnotify.Remove) && !event.Has(fsnotify.Rename) {
				continue
			}

			// 只关心我们管理的日志文件。
			if !isManagedLogPath(event.Name) {
				continue
			}

			// lumberjack 轮转会先 rename 再创建新文件，这里延迟检查，
			// 避免对正常轮转误触发 SetLogger。
			path := event.Name
			go func() {
				time.Sleep(300 * time.Millisecond)
				if _, statErr := os.Stat(path); statErr == nil {
					return
				}
				zap.L().Warn("log file missing, reopening logger", zap.String("path", path))
				SetLogger()
			}()
		case watcherErr, ok := <-watcher.Errors:
			if !ok {
				return
			}
			zap.L().Error("file listening error", zap.Error(watcherErr))
		case <-done:
			// 收到停止信号，退出监控
			return
		}
	}
}

func GetLogger() *zap.Logger {
	mu.RLock()
	l := logger
	mu.RUnlock()
	if l != nil {
		return l
	}

	SetLogger()
	mu.RLock()
	defer mu.RUnlock()
	return logger
}

// GetGinLogger 返回 gin access log 使用的 logger（独立文件）。
func GetGinLogger() *zap.Logger {
	mu.RLock()
	l := ginLogger
	mu.RUnlock()
	if l != nil {
		return l
	}

	SetLogger()
	mu.RLock()
	defer mu.RUnlock()
	return ginLogger
}

// GetGinErrorLogger 返回 gin panic/recovery 等错误输出使用的 logger（独立文件）。
func GetGinErrorLogger() *zap.Logger {
	mu.RLock()
	l := ginErrorLogger
	mu.RUnlock()
	if l != nil {
		return l
	}

	SetLogger()
	mu.RLock()
	defer mu.RUnlock()
	return ginErrorLogger
}

// zapWriter 是一个 io.Writer 实现，用于将框架类日志（如 gin/gorm）转发到 zap
type zapWriter struct {
	getLogger func() *zap.Logger
	level     zapcore.Level
}

// Write 实现 io.Writer 接口，将写入内容作为消息输出到 zap
func (w *zapWriter) Write(p []byte) (n int, err error) {
	if w == nil || w.getLogger == nil {
		// 在 logger 提供器缺失的场景中避免 panic，同时不阻塞调用方
		return len(p), nil
	}

	l := w.getLogger()
	if l == nil {
		// 在 logger 尚未初始化的场景中避免 panic，同时不阻塞调用方
		return len(p), nil
	}

	msg := strings.TrimRight(string(p), "\r\n")
	if ce := l.Check(w.level, msg); ce != nil {
		ce.Write()
	}
	return len(p), nil
}

// NewZapWriter 创建一个基于 zap 的 io.Writer，方便将第三方日志重定向到统一的 zap 日志管道
func NewZapWriter(l *zap.Logger, level zapcore.Level) *zapWriter {
	if l == nil {
		return NewZapWriterFunc(GetLogger, level)
	}
	return NewZapWriterFunc(func() *zap.Logger { return l }, level)
}

// NewZapWriterFunc 创建一个动态 logger 的 io.Writer。
// 典型用法：第三方组件在运行期需要切换 logger（例如日志文件被删除后重建）。
func NewZapWriterFunc(getLogger func() *zap.Logger, level zapcore.Level) *zapWriter {
	if getLogger == nil {
		getLogger = GetLogger
	}
	return &zapWriter{
		getLogger: getLogger,
		level:     level,
	}
}

func init() {
	// init 时不初始化 logger，等待 main 中设置好 RunModel 后再初始化
	// SetLogger() 会在 GetLogger() 第一次调用时执行
}

// StartMonitor 启动日志文件监控（需在初始化后调用）
// 注意：仅在生产模式下启动监控，开发模式输出到终端，不需要监控
func StartMonitor() {
	// 只在生产模式下启动文件监控
	if runmodel.IsRelease() {
		mu.Lock()
		if monitorDone != nil || rotateDone != nil {
			mu.Unlock()
			return
		}
		monitorDone = make(chan struct{})
		rotateDone = make(chan struct{})
		mu.Unlock()

		go monitorFile(monitorDone)
		go rotateDaily(rotateDone)
	}
}

// StopMonitor 停止日志文件监控并刷新日志缓冲区（应用关闭时调用）
func StopMonitor() {
	mu.Lock()
	md := monitorDone
	rd := rotateDone
	monitorDone = nil
	rotateDone = nil
	bl := logger
	gl := ginLogger
	mu.Unlock()

	// 停止后台 goroutine
	if md != nil {
		close(md)
	}
	if rd != nil {
		close(rd)
	}

	// 刷新日志缓冲区
	if bl != nil {
		_ = bl.Sync()
	}
	if gl != nil {
		_ = gl.Sync()
	}
}

func rotateDaily(done <-chan struct{}) {
	if done == nil {
		return
	}

	for {
		now := time.Now()
		next := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location())
		wait := time.Until(next)
		if wait <= 0 {
			wait = time.Second
		}

		t := time.NewTimer(wait)
		select {
		case <-t.C:
			rotateAll()
		case <-done:
			if !t.Stop() {
				select {
				case <-t.C:
				default:
				}
			}
			return
		}
	}
}

func rotateAll() {
	mu.RLock()
	blj := loggerLJ
	glj := ginLoggerLJ
	mu.RUnlock()

	if blj != nil {
		if err := blj.Rotate(); err != nil {
			zap.L().Warn("rotate business log failed", zap.Error(err))
		}
	}
	if glj != nil {
		if err := glj.Rotate(); err != nil {
			zap.L().Warn("rotate gin log failed", zap.Error(err))
		}
	}
}

func ginLogPath() string {
	// 默认与业务日志同目录：log/<程序名>.gin.log
	return filepath.Join(config.LogDir, fmt.Sprintf("%s.gin.log", config.SelfName))
}

func isManagedLogPath(path string) bool {
	clean := filepath.Clean(path)
	if clean == filepath.Clean(config.LogPath) {
		return true
	}
	if clean == filepath.Clean(ginLogPath()) {
		return true
	}
	return false
}

func parseLogLevel(levelStr string) zapcore.Level {
	switch strings.ToLower(strings.TrimSpace(levelStr)) {
	case "debug":
		return zap.DebugLevel
	case "info":
		return zap.InfoLevel
	case "warn", "warning":
		return zap.WarnLevel
	case "error":
		return zap.ErrorLevel
	default:
		zap.L().Warn("未识别的日志级别，使用默认 info 级别",
			zap.String("input_level", levelStr),
		)
		return zap.InfoLevel
	}
}

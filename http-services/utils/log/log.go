package log

import (
	"os"
	"time"

	"http-services/config"
	runmodel "http-services/utils/run-model"

	"github.com/fsnotify/fsnotify"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	logger      *zap.Logger
	monitorDone chan struct{} // 用于停止监控 goroutine
)

// Creating Dev logger
// DEV mode outputs logs to the terminal and is more readable
func createDevLogger() *zap.Logger {
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
		MaxSize:    config.LogMaxSize,
		MaxBackups: config.LogMaxBackups,
		MaxAge:     config.LogMaxAge,
	})
	core := zapcore.NewTee(
		zapcore.NewSamplerWithOptions(
			zapcore.NewCore(zapcore.NewJSONEncoder(fileEncoder), fileWriter, zap.InfoLevel), time.Second, 4, 1),
	)
	return zap.New(core, zap.AddCaller())
}

// SetLogger to prevent zap persistence problems after files are deleted
func SetLogger() {
	// Get log mode
	switch {
	case runmodel.IsDev():
		logger = createDevLogger()
	case runmodel.IsRelease():
		logger = createProductLogger(config.LogPath)
	default:
		logger = createProductLogger(config.LogPath)
	}
	zap.ReplaceGlobals(logger)
}

// Listen to log files
// When the log file is deleted manually, we will automatically create a new one.
func monitorFile() {
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
				zap.L().Warn("the log file was deleted")
				SetLogger()
			}
			if event.Has(fsnotify.Rename) {
				zap.L().Warn("log files are renamed and new files are monitored")
				SetLogger()
			}
		case err := <-watcher.Errors:
			zap.L().Error("file listening error", zap.Error(err))
		case <-monitorDone:
			// 收到停止信号，退出监控
			return
		}
	}
}

func GetLogger() *zap.Logger {
	// 如果 logger 还未初始化，先初始化
	if logger == nil {
		SetLogger()
	}
	return logger
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
		monitorDone = make(chan struct{})
		go monitorFile()
	}
}

// StopMonitor 停止日志文件监控并刷新日志缓冲区（应用关闭时调用）
func StopMonitor() {
	// 停止文件监控 goroutine
	if monitorDone != nil {
		close(monitorDone)
	}

	// 刷新日志缓冲区
	if logger != nil {
		_ = logger.Sync()
	}
}

// FromContext 从 gin.Context 中获取带上下文信息的 logger
// 如果 context 中没有 logger，则返回全局 logger
// 这个函数应该在业务处理器中使用，以获取包含 trace_id、method、path 等上下文信息的 logger
//
// 使用示例:
//
//	func Handler(c *gin.Context) {
//	    logger := log.FromContext(c)
//	    logger.Info("处理用户请求", zap.String("user_id", userID))
//	}
func FromContext(c *gin.Context) *zap.Logger {
	// 尝试从 context 获取 logger
	if loggerVal, exists := c.Get("logger"); exists {
		if contextLogger, ok := loggerVal.(*zap.Logger); ok {
			return contextLogger
		}
	}

	// 如果没有上下文 logger，返回全局 logger
	// 这种情况通常发生在测试或者中间件执行顺序问题
	return GetLogger()
}

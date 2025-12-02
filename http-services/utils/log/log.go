package log

import (
	"net/http"
	"os"
	"strings"
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

// BoundParamsKey 用于在 gin.Context 中存放已绑定的业务参数。
// 目前由 middleware.CheckParam 进行写入，WithRequest 进行读取，仅用于日志记录。
const BoundParamsKey = "__bound_params__"

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
		// 默认视作开发模式，避免测试/包初始化阶段创建文件与目录
		logger = createDevLogger()
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

// zapWriter 是一个 io.Writer 实现，用于将框架类日志（如 gin/gorm）转发到 zap
type zapWriter struct {
	logger *zap.Logger
	level  zapcore.Level
}

// Write 实现 io.Writer 接口，将写入内容作为消息输出到 zap
func (w *zapWriter) Write(p []byte) (n int, err error) {
	if w == nil || w.logger == nil {
		// 在 logger 尚未初始化的场景中避免 panic，同时不阻塞调用方
		return len(p), nil
	}

	msg := strings.TrimRight(string(p), "\r\n")
	if ce := w.logger.Check(w.level, msg); ce != nil {
		ce.Write()
	}
	return len(p), nil
}

// NewZapWriter 创建一个基于 zap 的 io.Writer，方便将第三方日志重定向到统一的 zap 日志管道
func NewZapWriter(l *zap.Logger, level zapcore.Level) *zapWriter {
	if l == nil {
		l = GetLogger()
	}
	return &zapWriter{
		logger: l,
		level:  level,
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

// WithRequest 从 gin.Context 中获取带请求参数信息的 logger。
// 仅在需要排查问题时调用，避免对所有请求都记录参数。
// 注意：为避免影响后续绑定与大体积请求处理，这里只记录：
//   - 查询参数（query）
//   - 已解析的表单参数（PostForm / MultipartForm.Value）
//   - 路径参数（path params）
//   - 通过中间件预绑定并挂载在 Context 上的业务参数（key: "__bound_params__"）
//
// 如需记录完整请求体（body），建议在专用中间件中提前拷贝并存入 context。
func WithRequest(c *gin.Context) *zap.Logger {
	base := FromContext(c)

	// 在单元测试或特殊场景中，Context 可能尚未完全初始化，
	// 此时直接返回基础 logger，避免空指针异常。
	if c == nil || c.Request == nil {
		return base
	}

	fields := []zap.Field{
		zap.String("method", c.Request.Method),
	}

	if c.Request.URL != nil {
		fields = append(fields, zap.String("path", c.Request.URL.Path))
		// 查询参数
		if rawQuery := c.Request.URL.RawQuery; rawQuery != "" {
			fields = append(fields, zap.String("query", rawQuery))
		}
	}

	// 已解析的表单参数（不会主动触发 ParseForm，避免多次读取 Body）
	if c.Request.Method == http.MethodPost || c.Request.Method == http.MethodPut || c.Request.Method == http.MethodPatch {
		// 普通表单
		if len(c.Request.PostForm) > 0 {
			fields = append(fields, zap.Any("form", c.Request.PostForm))
		}
		// multipart 表单
		if c.Request.MultipartForm != nil && len(c.Request.MultipartForm.Value) > 0 {
			fields = append(fields, zap.Any("multipart_form", c.Request.MultipartForm.Value))
		}
	}

	// 路径参数
	if len(c.Params) > 0 {
		pathParams := make(map[string]string, len(c.Params))
		for _, p := range c.Params {
			pathParams[p.Key] = p.Value
		}
		fields = append(fields, zap.Any("path_params", pathParams))
	}

	// 已绑定的业务参数（例如通过 middleware.CheckParam 绑定的 JSON / 表单参数）
	if bound, exists := c.Get(BoundParamsKey); exists && bound != nil {
		fields = append(fields, zap.Any("params", bound))
	}

	return base.With(fields...)
}

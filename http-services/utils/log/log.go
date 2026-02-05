package log

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
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
	mu sync.RWMutex

	logger   *zap.Logger
	loggerLJ *lumberjack.Logger

	ginLogger      *zap.Logger
	ginErrorLogger *zap.Logger
	ginLoggerLJ    *lumberjack.Logger

	monitorDone chan struct{} // 用于停止监控 goroutine
	rotateDone  chan struct{} // 用于停止按天 Rotate goroutine
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
func createProductLogger(fileName string) (*zap.Logger, *lumberjack.Logger) {
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
			zapcore.NewCore(zapcore.NewJSONEncoder(fileEncoder), fileWriter, zap.InfoLevel), time.Second, 4, 1),
	)
	return zap.New(core, zap.AddCaller()), lj
}

// SetLogger to prevent zap persistence problems after files are deleted
func SetLogger() {
	mu.Lock()
	defer mu.Unlock()

	// Get log mode
	switch {
	case runmodel.IsDev():
		logger = createDevLogger()
		loggerLJ = nil

		ginLogger = createDevLogger().With(zap.String("logger", "gin"))
		ginErrorLogger = ginLogger.With(zap.String("stream", "stderr"))
		ginLoggerLJ = nil
	case runmodel.IsRelease():
		logger, loggerLJ = createProductLogger(config.LogPath)

		ginLogger, ginLoggerLJ = createProductLogger(ginLogPath())
		ginLogger = ginLogger.With(zap.String("logger", "gin"))
		ginErrorLogger = ginLogger.With(zap.String("stream", "stderr"))
	default:
		// 默认视作开发模式，避免测试/包初始化阶段创建文件与目录
		logger = createDevLogger()
		loggerLJ = nil

		ginLogger = createDevLogger().With(zap.String("logger", "gin"))
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
	defer watcher.Close()
	// 监控日志目录，避免因日志文件轮转（rename）导致 watcher 失效。
	err = watcher.Add(config.LogDir)
	if err != nil {
		zap.L().Error("File listening error", zap.Error(err))
	}
	for {
		select {
		case event := <-watcher.Events:
			if !(event.Has(fsnotify.Remove) || event.Has(fsnotify.Rename)) {
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
		case err := <-watcher.Errors:
			zap.L().Error("file listening error", zap.Error(err))
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

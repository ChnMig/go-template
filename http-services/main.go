package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"http-services/api"
	"http-services/api/middleware"
	"http-services/config"
	"http-services/db"
	"http-services/db/msqldb"
	"http-services/db/rdb"
	"http-services/utils/log"
	"http-services/utils/pathtool"
	"http-services/utils/pidfile"
	"http-services/utils/runmodel"

	"github.com/alecthomas/kong"
	"go.uber.org/zap"
)

var CLI struct {
	Dev     bool `help:"Run in development mode" short:"d"`
	Version bool `help:"Show version information" short:"v"`
	Migrate bool `help:"Run database migrations and exit" short:"m"`
}

var (
	Version   = "dev"
	BuildTime = "unknown"
	GitCommit = "unknown"
)

func main() {
	// 解析命令行参数
	ctx := kong.Parse(&CLI,
		kong.Name("http-services"),
		kong.Description("HTTP API services"),
		kong.UsageOnError(),
	)

	// 显示版本信息
	if CLI.Version {
		fmt.Printf("Version:    %s\n", Version)
		fmt.Printf("Build Time: %s\n", BuildTime)
		fmt.Printf("Git Commit: %s\n", GitCommit)
		os.Exit(0)
	}

	// 从配置文件加载配置
	if err := config.LoadConfig(); err != nil {
		fmt.Printf("Failed to load configuration: %v\n", err)
		ctx.Exit(1)
	}

	// 设置运行模式（必须在初始化日志之前）
	runmodel.Detect(CLI.Dev)

	// 初始化日志（在设置好 RunModel 之后）
	// 仅在 release 模式创建日志目录，避免在测试/子包初始化时散落空 log 目录
	if config.RunModel == config.RunModelRelease {
		_ = pathtool.CreateDir(config.LogDir)
	}
	log.GetLogger()
	log.StartMonitor() // 启动日志文件监控

	// 启动配置热重载（在日志初始化之后）
	config.WatchConfig(func() {
		log.SetLogger()
		zap.L().Info("Configuration reloaded",
			zap.Int("port", config.ListenPort),
			zap.Duration("jwt_expiration", config.JWTExpiration),
			zap.Bool("rate_limit_enabled", config.EnableRateLimit),
		)
	})

	// 校验配置
	config.CheckConfig(
		config.JWTKey,
		int64(config.JWTExpiration),
	)

	if CLI.Migrate {
		zap.L().Info("Running database migrations...")
		if err := db.MigrateAll(); err != nil {
			zap.L().Fatal("Database migration failed", zap.Error(err))
		}
		msqldb.CloseClient()
		rdb.CloseClient()
		middleware.CleanupAllLimiters()
		log.StopMonitor()
		zap.L().Info("Database migration completed successfully")
		ctx.Exit(0)
	}

	zap.L().Info("Starting HTTP service",
		zap.String("mode", config.RunModel),
		zap.Int("port", config.ListenPort),
		zap.String("version", Version),
	)

	// 初始化 API 路由
	r := api.InitApi()

	// 创建 HTTP 服务器（使用配置化的超时参数）
	addr := fmt.Sprintf(":%d", config.ListenPort)
	srv := &http.Server{
		Addr:           addr,
		Handler:        r,
		ReadTimeout:    config.ReadTimeout,
		WriteTimeout:   config.WriteTimeout,
		IdleTimeout:    config.IdleTimeout,
		MaxHeaderBytes: config.MaxHeaderBytes,
	}

	// 监听停止信号（尽早注册，避免启动阶段收到信号时错过清理流程）
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)

	pidFilePath := config.PidFile
	pid := os.Getpid()
	// 独占写入 pid 文件，避免多个进程同时持有同一路径。
	if pidFilePath != "" {
		if err := pidfile.Write(pidFilePath, pid); err != nil {
			zap.L().Fatal("写入 pid 文件失败",
				zap.String("pid_file", pidFilePath),
				zap.Error(err),
			)
		}
		zap.L().Info("PID 文件已写入",
			zap.String("pid_file", pidFilePath),
			zap.Int("pid", pid),
		)
	}

	serverErrCh := make(chan error, 1)
	go func() {
		zap.L().Info("Server is starting...")
		err := srv.ListenAndServe()

		if err != nil && err != http.ErrServerClosed {
			serverErrCh <- err
		}
	}()

	exitCode := 0
	select {
	case sig := <-quit:
		zap.L().Info("Received stop signal, shutting down gracefully", zap.String("signal", sig.String()))
	case err := <-serverErrCh:
		exitCode = 1
		zap.L().Error("HTTP 服务异常退出，开始执行清理与退出",
			zap.Error(err),
		)
	}

	// 创建带超时的 context 用于优雅关闭（使用配置化的超时时间）
	shutdownCtx, cancel := context.WithTimeout(context.Background(), config.ShutdownTimeout)

	// 优雅关闭服务器
	if err := srv.Shutdown(shutdownCtx); err != nil {
		if !errors.Is(err, http.ErrServerClosed) {
			zap.L().Error("Server forced to shutdown", zap.Error(err))
		}
		// 即使服务器强制关闭，也要尝试清理资源
	}

	// 清理资源
	zap.L().Info("Cleaning up resources...")
	middleware.CleanupAllLimiters() // 清理限流器
	msqldb.CloseClient()            // 关闭 MySQL 连接池（如已初始化）
	rdb.CloseClient()               // 关闭 Redis 连接池（如已初始化）
	log.StopMonitor()               // 停止日志监控并刷新缓冲区

	// 仅删除仍由当前进程持有的 pid 文件。
	if pidFilePath != "" {
		if err := pidfile.Remove(pidFilePath, pid); err != nil {
			zap.L().Warn("删除 pid 文件失败",
				zap.String("pid_file", pidFilePath),
				zap.Error(err),
			)
		}
	}

	cancel()

	zap.L().Info("Server exited", zap.Int("exit_code", exitCode))
	ctx.Exit(exitCode)
}

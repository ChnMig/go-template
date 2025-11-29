package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"http-services/api"
	"http-services/api/middleware"
	"http-services/config"
	"http-services/utils/acme"
	"http-services/utils/log"
	pathtool "http-services/utils/path-tool"

	"github.com/alecthomas/kong"
	"go.uber.org/zap"
)

var CLI struct {
	Dev     bool `help:"Run in development mode" short:"d"`
	Version bool `help:"Show version information" short:"v"`
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
	if CLI.Dev {
		config.RunModel = config.RunModelDevValue
	} else {
		// 从环境变量检测运行模式（如果没有指定 --dev）
		model := os.Getenv(config.RunModelKey)
		switch model {
		case config.RunModelDevValue:
			config.RunModel = config.RunModelDevValue
		case config.RunModelRelease:
			config.RunModel = config.RunModelRelease
		default:
			config.RunModel = config.RunModelRelease
		}
	}

	// 初始化日志（在设置好 RunModel 之后）
	// 仅在 release 模式创建日志目录，避免在测试/子包初始化时散落空 log 目录
	if config.RunModel == config.RunModelRelease {
		_ = pathtool.CreateDir(config.LogDir)
	}
	log.GetLogger()
	log.StartMonitor() // 启动日志文件监控

	// 启动配置热重载（在日志初始化之后）
	config.WatchConfig(func() {
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

	// 根据配置为服务挂载可选的 ACME 自动 TLS 能力
	acmeCtx := acme.Setup(srv)

	// 在 goroutine 中启动服务器
	if acmeCtx.Enabled && acmeCtx.HTTPServer != nil {
		// 启动 ACME HTTP 挑战服务器（80 端口）
		go func() {
			zap.L().Info("ACME HTTP 挑战服务器启动", zap.String("addr", acmeCtx.HTTPServer.Addr))
			if err := acmeCtx.HTTPServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				zap.L().Error("ACME HTTP 挑战服务器异常退出", zap.Error(err))
			}
		}()
	}

	go func() {
		if acmeCtx.Enabled {
			zap.L().Info("Server is starting with ACME TLS...",
				zap.String("addr", srv.Addr),
				zap.String("domain", config.ACMEDomain),
			)
			if err := srv.ListenAndServeTLS("", ""); err != nil && err != http.ErrServerClosed {
				zap.L().Fatal("Failed to start ACME TLS server", zap.Error(err))
			}
		} else {
			zap.L().Info("Server is starting...")
			if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				zap.L().Fatal("Failed to start server", zap.Error(err))
			}
		}
	}()

	// 监听停止信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	sig := <-quit
	zap.L().Info("Received stop signal, shutting down gracefully", zap.String("signal", sig.String()))

	// 创建带超时的 context 用于优雅关闭（使用配置化的超时时间）
	shutdownCtx, cancel := context.WithTimeout(context.Background(), config.ShutdownTimeout)
	defer cancel()

	// 优雅关闭服务器
	if err := srv.Shutdown(shutdownCtx); err != nil {
		zap.L().Error("Server forced to shutdown", zap.Error(err))
		// 即使服务器强制关闭，也要尝试清理资源
	}

	// 关闭 ACME HTTP 挑战服务器
	if acmeCtx.Enabled && acmeCtx.HTTPServer != nil {
		if err := acmeCtx.HTTPServer.Shutdown(shutdownCtx); err != nil {
			zap.L().Error("ACME HTTP 挑战服务器关闭失败", zap.Error(err))
		}
	}

	// 清理资源
	zap.L().Info("Cleaning up resources...")
	middleware.CleanupAllLimiters() // 清理限流器
	log.StopMonitor()               // 停止日志监控并刷新缓冲区

	zap.L().Info("Server exited gracefully")
	ctx.Exit(0)
}

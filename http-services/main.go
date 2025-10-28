package main

import (
	"context"
	"fmt"
	"http-services/api"
	"http-services/config"
	"http-services/utils/log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

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
		os.Exit(1)
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
	log.GetLogger()
	log.StartMonitor() // 启动日志文件监控

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

	// 创建 HTTP 服务器
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", config.ListenPort),
		Handler: r,
	}

	// 在 goroutine 中启动服务器
	go func() {
		zap.L().Info("Server is starting...")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			zap.L().Fatal("Failed to start server", zap.Error(err))
		}
	}()

	// 监听停止信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	sig := <-quit
	zap.L().Info("Received stop signal, shutting down gracefully", zap.String("signal", sig.String()))

	// 创建带超时的 context 用于优雅关闭
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 优雅关闭服务器
	if err := srv.Shutdown(shutdownCtx); err != nil {
		zap.L().Error("Server forced to shutdown", zap.Error(err))
		ctx.Exit(1)
	}

	zap.L().Info("Server exited gracefully")
	ctx.Exit(0)
}

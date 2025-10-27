package main

import (
	"fmt"
	"go-services/api"
	"go-services/config"
	"go-services/util/log"
	"os"
	"os/signal"
	"syscall"

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

	// 设置运行模式
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

	// 初始化日志
	log.GetLogger()

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
	go func() {
		if err := r.Run(fmt.Sprintf(":%d", config.ListenPort)); err != nil {
			zap.L().Fatal("Failed to start server", zap.Error(err))
		}
	}()

	// 监听停止信号
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGTERM, syscall.SIGINT)
	sig := <-sigs
	zap.L().Info("Received stop signal, shutting down", zap.String("signal", sig.String()))

	// 执行清理工作
	ctx.Exit(0)
}

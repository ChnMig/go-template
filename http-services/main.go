package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"http-services/api"
	"http-services/api/middleware"
	"http-services/config"
	"http-services/db"
	"http-services/db/msqldb"
	"http-services/db/rdb"
	"http-services/utils/log"
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
	command := kong.Parse(
		&CLI,
		kong.Name("http-services"),
		kong.Description("HTTP API services"),
		kong.UsageOnError(),
	)

	if CLI.Version {
		fmt.Printf("Version:    %s\n", Version)
		fmt.Printf("Build Time: %s\n", BuildTime)
		fmt.Printf("Git Commit: %s\n", GitCommit)
		return
	}

	if err := config.LoadConfig(); err != nil {
		fmt.Printf("Failed to load configuration: %v\n", err)
		command.Exit(1)
	}
	runmodel.Detect(CLI.Dev)
	if runmodel.IsRelease() {
		if err := os.MkdirAll(config.LogDir, 0o750); err != nil {
			fmt.Printf("Failed to create log directory: %v\n", err)
			command.Exit(1)
		}
	}
	log.GetLogger()
	log.StartMonitor()
	config.WatchConfig(func() {
		log.SetLogger()
		zap.L().Info(
			"Configuration reloaded",
			zap.Int("port", config.ListenPort),
			zap.Bool("rate_limit_enabled", config.EnableRateLimit),
		)
	})

	if CLI.Migrate {
		zap.L().Info("Running database migrations...")
		if err := db.MigrateAll(); err != nil {
			zap.L().Error("Database migration failed", zap.Error(err))
			msqldb.CloseClient()
			rdb.CloseClient()
			middleware.CleanupAllLimiters()
			log.StopMonitor()
			command.Exit(1)
		}
		zap.L().Info("Database migration completed successfully")
		msqldb.CloseClient()
		rdb.CloseClient()
		middleware.CleanupAllLimiters()
		log.StopMonitor()
		return
	}

	router := api.InitApi()
	server := &http.Server{
		Addr:           net.JoinHostPort(config.ListenHost, strconv.Itoa(config.ListenPort)),
		Handler:        router,
		ReadTimeout:    config.ReadTimeout,
		WriteTimeout:   config.WriteTimeout,
		IdleTimeout:    config.IdleTimeout,
		MaxHeaderBytes: config.MaxHeaderBytes,
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	defer signal.Stop(quit)

	pid := os.Getpid()
	if config.PidFile != "" {
		if err := pidfile.Write(config.PidFile, pid); err != nil {
			zap.L().Error("写入 pid 文件失败", zap.String("pid_file", config.PidFile), zap.Error(err))
			msqldb.CloseClient()
			rdb.CloseClient()
			middleware.CleanupAllLimiters()
			log.StopMonitor()
			command.Exit(1)
		}
		zap.L().Info("PID 文件已写入", zap.String("pid_file", config.PidFile), zap.Int("pid", pid))
	}

	serverErrCh := make(chan error, 1)
	go func() {
		zap.L().Info("Server is starting...", zap.String("addr", server.Addr), zap.String("version", Version))
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErrCh <- err
		}
	}()

	exitCode := 0
	select {
	case received := <-quit:
		zap.L().Info("Received stop signal, shutting down gracefully", zap.String("signal", received.String()))
	case err := <-serverErrCh:
		exitCode = 1
		zap.L().Error("HTTP 服务异常退出，开始执行清理与退出", zap.Error(err))
	}

	shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), config.ShutdownTimeout)
	if err := server.Shutdown(shutdownCtx); err != nil && !errors.Is(err, http.ErrServerClosed) {
		exitCode = 1
		zap.L().Error("Server forced to shutdown", zap.Error(err))
	}
	cancelShutdown()

	zap.L().Info("Cleaning up resources...")
	middleware.CleanupAllLimiters()
	msqldb.CloseClient()
	rdb.CloseClient()
	if config.PidFile != "" {
		if err := pidfile.Remove(config.PidFile, pid); err != nil {
			zap.L().Warn("删除 pid 文件失败", zap.String("pid_file", config.PidFile), zap.Error(err))
		}
	}
	zap.L().Info("Server exited", zap.Int("exit_code", exitCode))
	log.StopMonitor()
	command.Exit(exitCode)
}

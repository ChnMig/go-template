package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"http-services/bootstrap"

	"github.com/alecthomas/kong"
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
	cliContext := kong.Parse(&CLI,
		kong.Name("http-services"),
		kong.Description("HTTP API services"),
		kong.UsageOnError(),
	)

	// 显示版本信息
	if CLI.Version {
		fmt.Printf("Version:    %s\n", Version)
		fmt.Printf("Build Time: %s\n", BuildTime)
		fmt.Printf("Git Commit: %s\n", GitCommit)
		return
	}

	runContext, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	err := bootstrap.Run(runContext, bootstrap.Options{
		PID:         os.Getpid(),
		Development: CLI.Dev,
		Migrate:     CLI.Migrate,
	})
	stop()
	if err != nil {
		fmt.Fprintf(os.Stderr, "http-services failed: %v\n", err)
		cliContext.Exit(1)
	}
}

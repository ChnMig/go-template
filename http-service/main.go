package main

import (
	"fmt"
	"go-services/config"
	"go-services/util/log"
	runmodel "go-services/util/run-model"
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/zap"
)

func main() {
	// 从配置文件加载配置
	if err := config.LoadConfig(); err != nil {
		fmt.Printf("Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// Ensure zap is initialized first
	log.GetLogger()

	// 校验配置
	config.CheckConfig(
		config.JWTKey,
		int64(config.JWTExpiration),
	)

	// Set the running mode of the program
	runmodel.Detection()
	// End of monitoring
	func() {
		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, syscall.SIGTERM, syscall.SIGINT)
		for {
			switch <-sigs {
			case syscall.SIGTERM, syscall.SIGINT:
				zap.L().Error("i picked up a stop signal.")
				return
			}
		}
	}()
}

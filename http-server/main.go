package main

import (
	"http-server/util/log"
	runmodel "http-server/util/run-model"
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/zap"
)

func main() {
	// Ensure zap is initialized first
	log.GetLogger()
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

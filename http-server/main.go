package main

import (
	"http-server/util/log"
	"time"

	"go.uber.org/zap"
)

func main() {
	// Ensure zap is initialized first
	log.GetLogger()
	for i := 0; ; i++ {
		zap.L().Error("test", zap.Int("i", i))
		time.Sleep(time.Second * 1)
	}
}

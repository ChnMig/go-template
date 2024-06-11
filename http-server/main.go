package main

import (
	"http-server/util/log"
	runmodel "http-server/util/run-model"
)

func main() {
	// Ensure zap is initialized first
	log.GetLogger()
	// Set the running mode of the program
	runmodel.Detection()
}

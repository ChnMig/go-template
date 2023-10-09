package main

import "http-server/util/log"

func main() {
	// Ensure zap is initialized first
	log.GetLogger()
}

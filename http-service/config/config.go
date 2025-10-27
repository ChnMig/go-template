package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	pathtool "go-services/util/path-tool"
)

// Here are some basic configurations
// These configurations are usually generic
var (
	// listen
	ListenPort = 8080 // api listen port
	// run model
	RunModelKey      = "model"
	RunModel         = ""
	RunModelDevValue = "dev"
	RunModelRelease  = "release"
	// path
	SelfName = filepath.Base(os.Args[0])      // own file name
	AbsPath  = pathtool.GetCurrentDirectory() // current directory
	// log
	LogDir        = filepath.Join(pathtool.GetCurrentDirectory(), "log")   // log directory
	LogPath       = filepath.Join(LogDir, fmt.Sprintf("%s.log", SelfName)) // self log path
	LogMaxSize    = 50                                                     // M
	LogMaxBackups = 3                                                      // backups
	LogMaxAge     = 30                                                     // days
	LogModelDev   = "dev"                                                  // dev model
)

// These configurations need to be modified as needed
var (
	// jWT
	JWTKey        = "N#xiAuAq!B!$d2Acq99Rz*Q*8&E" // Key must be regenerated, otherwise there will be security risks
	JWTExpiration = time.Hour * 12
)

func init() {
	pathtool.CreateDir(LogDir)
}

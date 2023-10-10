package config

import (
	"fmt"
	"os"
	"path/filepath"

	pathtool "http-server/util/path-tool"
)

// Here are some basic configurations
// These configurations are usually generic
var (
	SelfName = filepath.Base(os.Args[0])                              // own file name
	AbsPath  = pathtool.GetCurrentDirectory()                         // current directory
	LogDir   = filepath.Join(pathtool.GetCurrentDirectory(), "log")   // log directory
	LogPath  = filepath.Join(LogDir, fmt.Sprintf("%s.log", SelfName)) // self log path
)

// These configurations vary according to actual usage scenarios
var (
	ListenPort = 8080 // api listen port
)

func init() {
	pathtool.CreateDir(LogDir)
}

package runmodel

import (
	"os"

	"http-services/config"
)

// Detect 根据 --dev 和 model 环境变量确定日志运行模式。
func Detect(cliDev bool) string {
	if cliDev {
		config.RunModel = config.RunModelDevValue
		return config.RunModel
	}

	switch os.Getenv(config.RunModelKey) {
	case config.RunModelDevValue:
		config.RunModel = config.RunModelDevValue
	default:
		config.RunModel = config.RunModelRelease
	}
	return config.RunModel
}

func IsDev() bool {
	return config.RunModel == config.RunModelDevValue
}

func IsRelease() bool {
	return config.RunModel == config.RunModelRelease
}

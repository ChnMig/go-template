package runmodel

import (
	"os"

	"http-services/config"
)

// Detect 根据命令行参数和环境变量确定运行模式。
// 命令行 --dev 优先级最高；未显式指定时默认 release。
func Detect(cliDev bool) string {
	if cliDev {
		config.RunModel = config.RunModelDevValue
		return config.RunModel
	}

	switch os.Getenv(config.RunModelKey) {
	case config.RunModelDevValue:
		config.RunModel = config.RunModelDevValue
	case config.RunModelRelease:
		config.RunModel = config.RunModelRelease
	default:
		config.RunModel = config.RunModelRelease
	}
	return config.RunModel
}

// Detection 保留旧函数名的兼容入口，默认只根据环境变量判断。
func Detection() {
	Detect(false)
}

func IsDev() bool {
	return config.RunModel == config.RunModelDevValue
}

func IsRelease() bool {
	return config.RunModel == config.RunModelRelease
}

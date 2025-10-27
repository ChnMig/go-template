package config

import (
	"go.uber.org/zap"
)

// CheckConfig 校验关键配置项，缺失则 fatal 并记录日志
func CheckConfig(
	JWTKey string,
	JWTExpiration int64,
) {
	if JWTKey == "" {
		zap.L().Fatal("JWTKey 配置缺失")
	}
	if JWTExpiration == 0 {
		zap.L().Fatal("JWTExpiration 配置缺失")
	}
}

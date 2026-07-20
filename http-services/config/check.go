package config

import (
	"fmt"

	"go.uber.org/zap"
)

const (
	// 最小 JWT 密钥长度
	minJWTKeyLength = 32
)

var unsafeDefaultKeys = map[string]struct{}{
	"YOUR_SECRET_KEY_HERE":                        {},
	"YOUR_SECRET_KEY_HERE_AT_LEAST_32_CHARACTERS": {},
}

func validateJWTConfig(jwtKey string, jwtExpiration int64) error {
	if jwtKey == "" {
		return fmt.Errorf("JWTKey 配置缺失，请在 config.yaml 中设置")
	}
	if _, ok := unsafeDefaultKeys[jwtKey]; ok {
		return fmt.Errorf("JWT 密钥仍使用示例占位值，请修改为强密钥")
	}
	if len(jwtKey) < minJWTKeyLength {
		return fmt.Errorf("JWT 密钥长度不足：当前 %d，至少需要 %d 个字符", len(jwtKey), minJWTKeyLength)
	}
	if jwtExpiration <= 0 {
		return fmt.Errorf("JWTExpiration 配置缺失或无效，请设置 jwt.expiration")
	}
	return nil
}

// ValidateConfig 校验关键配置项并向调用方返回错误。
func ValidateConfig(jwtKey string, jwtExpiration int64) error {
	return validateJWTConfig(jwtKey, jwtExpiration)
}

// CheckConfig 校验关键配置项，缺失或不安全则 fatal 并记录日志。
func CheckConfig(jwtKey string, jwtExpiration int64) {
	if err := ValidateConfig(jwtKey, jwtExpiration); err != nil {
		zap.L().Fatal("关键配置校验失败", zap.Error(err))
	}
}

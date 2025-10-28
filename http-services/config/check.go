package config

import (
	"go.uber.org/zap"
)

const (
	// 最小 JWT 密钥长度
	minJWTKeyLength = 32
	// 不安全的默认密钥
	unsafeDefaultKey = "YOUR_SECRET_KEY_HERE"
)

// CheckConfig 校验关键配置项，缺失或不安全则 fatal 并记录日志
func CheckConfig(
	JWTKey string,
	JWTExpiration int64,
) {
	// 检查 JWT 密钥是否为空
	if JWTKey == "" {
		zap.L().Fatal("JWTKey 配置缺失，请在 config.yaml 中设置")
	}

	// 检查是否使用了默认的不安全密钥
	if JWTKey == unsafeDefaultKey {
		zap.L().Fatal("JWT 密钥仍使用示例值，存在严重安全风险！请修改 config.yaml 中的 jwt.key 为强密钥")
	}

	// 检查密钥长度是否足够
	if len(JWTKey) < minJWTKeyLength {
		zap.L().Fatal("JWT 密钥长度不足",
			zap.Int("current_length", len(JWTKey)),
			zap.Int("min_required", minJWTKeyLength),
			zap.String("suggestion", "请使用至少32字符的强密钥"),
		)
	}

	// 检查过期时间是否设置
	if JWTExpiration == 0 {
		zap.L().Fatal("JWTExpiration 配置缺失，请在 config.yaml 中设置 jwt.expiration")
	}
}

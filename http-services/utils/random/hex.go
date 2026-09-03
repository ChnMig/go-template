package random

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"errors"
)

// Hex 返回 nBytes 个随机字节的 hex 字符串（小写），长度为 2*nBytes。
func Hex(nBytes int) (string, error) {
	if nBytes <= 0 {
		return "", errors.New("nBytes must be positive")
	}
	b := make([]byte, nBytes)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// Base64URL 返回 nBytes 个随机字节的无填充 Base64URL 字符串。
// 适用于需要放入 URL、但不能暴露业务身份信息的不可枚举安全令牌。
func Base64URL(nBytes int) (string, error) {
	if nBytes <= 0 {
		return "", errors.New("nBytes must be positive")
	}
	b := make([]byte, nBytes)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

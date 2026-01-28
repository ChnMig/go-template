package random

import (
	"crypto/rand"
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

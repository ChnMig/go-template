// Package id 提供进程内唯一标识生成工具。
package id

import (
	"crypto/md5"
	"fmt"

	"github.com/google/uuid"
)

// GenerateUUIDv7 生成符合 RFC 9562 的 UUIDv7。
func GenerateUUIDv7() (uuid.UUID, error) {
	generated, err := uuid.NewV7()
	if err != nil {
		return uuid.Nil, fmt.Errorf("generate UUIDv7: %w", err)
	}
	return generated, nil
}

// GenerateUUIDv7MD5 生成 UUIDv7 标准字符串的 32 位小写 MD5 表示。
func GenerateUUIDv7MD5() (string, error) {
	generated, err := GenerateUUIDv7()
	if err != nil {
		return "", err
	}
	digest := md5.Sum([]byte(generated.String()))
	return fmt.Sprintf("%x", digest), nil
}

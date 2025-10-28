package id

import (
	"crypto/md5"
	"fmt"

	"github.com/sony/sonyflake"
)

var flake *sonyflake.Sonyflake

// IssueID 生成唯一 ID (基于 Sony 改进的 Snowflake 算法)
// https://github.com/sony/sonyflake
func IssueID() string {
	id, _ := flake.NextID()
	return fmt.Sprintf("%v", id)
}

// GenerateID 生成唯一 ID (基于 Sonyflake + MD5)
func GenerateID() string {
	keyID := IssueID()
	id := fmt.Sprintf("%x", md5.Sum([]byte(keyID)))
	return id
}

func init() {
	flake = sonyflake.NewSonyflake(sonyflake.Settings{})
}

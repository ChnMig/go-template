package pathtool

import (
	"os"
	"path/filepath"
	"strings"
)

// GetCurrentDirectory 返回当前工作目录，获取失败时回退到可执行文件所在目录。
func GetCurrentDirectory() string {
	dir, err := os.Getwd()
	if err != nil {
		dir, _ = filepath.Abs(filepath.Dir(os.Args[0]))
	}
	return strings.Replace(dir, "\\", "/", -1)
}

// PathExists 判断路径是否存在。
func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

// CreateDir 创建目录，目录已存在时直接返回。
func CreateDir(path string) error {
	exist, err := PathExists(path)
	if err != nil {
		return err
	}
	if exist {
		return nil
	}
	return os.MkdirAll(path, os.ModePerm)
}

// CreateFile 创建空文件，文件已存在时直接返回。
func CreateFile(path string) error {
	dir := filepath.Dir(path)
	if err := CreateDir(dir); err != nil {
		return err
	}
	exist, err := PathExists(path)
	if err != nil {
		return err
	}
	if exist {
		return nil
	}
	return os.WriteFile(path, []byte(""), 0o644)
}

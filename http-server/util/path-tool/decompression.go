package pathtool

import (
	"archive/tar"
	"compress/gzip"
	"debug/elf"
	"io"
	"os"
	"path/filepath"

	"go.uber.org/zap"
)

func IsELF(file string) bool {
	f, err := elf.Open(file)
	if err != nil {
		zap.L().Error("ELF文件打开错误", zap.Error(err))
		return false
	}
	defer f.Close()
	switch f.FileHeader.Class {
	case elf.ELFCLASS64:
		return true
	case elf.ELFCLASS32:
		return true
	default:
		return false
	}
}

// tar.gz 解压
// replacement = true 解压目录有相同文件名存在则覆盖
func Decompression(originFile, outPath string, replacement bool) error {
	file, err := os.Open(originFile)
	if err != nil {
		zap.L().Error("压缩文件读取失败", zap.Error(err))
		return err
	}
	defer file.Close()
	gzipReader, err := gzip.NewReader(file)
	if err != nil {
		zap.L().Error("压缩文件读取失败", zap.Error(err))
		return err
	}
	defer gzipReader.Close()
	CreateDir(outPath)
	tarReader := tar.NewReader(gzipReader)
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			zap.L().Error("压缩文件解压失败", zap.Error(err))
			return err
		}
		switch header.Typeflag {
		case tar.TypeDir:
			abs := filepath.Join(outPath, header.Name)
			ok, err := PathExists(abs)
			if err != nil {
				zap.L().Error("压缩文件路径获取失败", zap.Error(err))
				continue
			}
			if !ok {
				if err := os.Mkdir(filepath.Join(outPath, header.Name), 0755); err != nil {
					zap.L().Error("压缩文件路径创建失败", zap.Error(err))
					return err
				}
			}
		case tar.TypeReg:
			abs := filepath.Join(outPath, header.Name)
			ok, err := PathExists(abs)
			if err != nil {
				zap.L().Error("压缩文件路径判断失败", zap.Error(err))
				continue
			}
			if ok {
				if replacement {
					// 替换
					os.RemoveAll(abs)
				} else {
					// 不做处理
					continue
				}
			}
			fileWriter, err := os.Create(filepath.Join(outPath, header.Name))
			if err != nil {
				zap.L().Error("压缩文件创建文件失败", zap.Error(err))
				return err
			}
			defer fileWriter.Close()
			if _, err := io.Copy(fileWriter, tarReader); err != nil {
				zap.L().Error("压缩文件copy文件失败", zap.Error(err))
				return err
			}
			if IsELF(filepath.Join(outPath, header.Name)) {
				// 加权限
				os.Chmod(filepath.Join(outPath, header.Name), os.ModePerm)
			}
		}
	}
	return nil
}

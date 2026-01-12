package tlsfile

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"path/filepath"
	"sync/atomic"
	"time"

	"http-services/config"
	pathtool "http-services/utils/path-tool"

	"github.com/fsnotify/fsnotify"
	"go.uber.org/zap"
)

// Context 描述基于本地证书文件的 TLS 运行时信息。
// Enabled 为 true 时表示当前进程启用了证书文件 TLS 模式。
type Context struct {
	Enabled bool
}

var currentCert atomic.Value // *tls.Certificate

// loadCertificate 从指定路径加载证书与私钥，并更新全局证书指针。
func loadCertificate(certPath, keyPath string) error {
	cert, err := tls.LoadX509KeyPair(certPath, keyPath)
	if err != nil {
		return err
	}
	currentCert.Store(&cert)
	return nil
}

// getCurrentCertificate 返回当前生效的 TLS 证书；仅用于内部与测试。
func getCurrentCertificate() *tls.Certificate {
	value := currentCert.Load()
	if value == nil {
		return nil
	}
	cert, ok := value.(*tls.Certificate)
	if !ok {
		return nil
	}
	return cert
}

// Setup 根据全局配置为 HTTP 服务器挂载基于本地证书文件的 TLS 能力，并启动文件变更监听实现证书热更新。
// - 当未启用 TLS 证书文件模式时，仅返回 Disabled 的上下文，不修改传入的 server；
// - 当启用时：
//   - 解析证书与私钥路径（支持相对路径，相对 config.AbsPath）；
//   - 加载证书写入全局缓存；
//   - 设置 server.TLSConfig.GetCertificate 回调；
//   - 使用 fsnotify 监听证书与私钥文件变更，变更时自动重新加载。
func Setup(server *http.Server) *Context {
	ctx := &Context{
		Enabled: false,
	}

	if !config.EnableTLS {
		return ctx
	}

	certPath := config.TLSCertFile
	keyPath := config.TLSKeyFile

	if certPath == "" || keyPath == "" {
		zap.L().Fatal("已启用 TLS 证书文件模式，但未配置证书或私钥路径",
			zap.String("tls_cert_file", certPath),
			zap.String("tls_key_file", keyPath),
		)
	}

	// 解析路径：相对路径基于程序所在目录
	if !filepath.IsAbs(certPath) {
		certPath = filepath.Join(config.AbsPath, certPath)
	}
	if !filepath.IsAbs(keyPath) {
		keyPath = filepath.Join(config.AbsPath, keyPath)
	}

	// 确保证书文件所在目录存在（证书轮换时通常不会改目录，但这里保持与其他路径处理一致）
	_ = pathtool.CreateDir(filepath.Dir(certPath))
	_ = pathtool.CreateDir(filepath.Dir(keyPath))

	if err := loadCertificate(certPath, keyPath); err != nil {
		zap.L().Fatal("加载 TLS 证书失败",
			zap.String("cert_file", certPath),
			zap.String("key_file", keyPath),
			zap.Error(err),
		)
	}

	// 配置 TLS：通过 GetCertificate 委托到当前缓存的证书，实现后续热更新。
	server.TLSConfig = &tls.Config{
		GetCertificate: func(info *tls.ClientHelloInfo) (*tls.Certificate, error) {
			cert := getCurrentCertificate()
			if cert == nil {
				return nil, fmt.Errorf("no TLS certificate loaded")
			}
			return cert, nil
		},
	}

	ctx.Enabled = true

	// 启动文件监听，用于证书热更新
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		zap.L().Error("创建 TLS 证书文件监听失败，后续证书将无法自动热更新",
			zap.Error(err),
		)
		return ctx
	}

	watchDirs := map[string]struct{}{
		filepath.Dir(certPath): {},
		filepath.Dir(keyPath):  {},
	}
	for dir := range watchDirs {
		if err := watcher.Add(dir); err != nil {
			zap.L().Error("监听 TLS 证书目录失败",
				zap.String("dir", dir),
				zap.Error(err),
			)
		}
	}

	watchFiles := map[string]struct{}{
		certPath: {},
		keyPath:  {},
	}

	go func() {
		defer watcher.Close()
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if _, ok := watchFiles[event.Name]; !ok {
					continue
				}
				// 证书或私钥文件发生写入/创建/重命名等变更时尝试重新加载
				if event.Has(fsnotify.Write) || event.Has(fsnotify.Create) || event.Has(fsnotify.Rename) || event.Has(fsnotify.Remove) || event.Has(fsnotify.Chmod) {
					// 简单的防抖处理，避免文件仍在写入过程中
					time.Sleep(200 * time.Millisecond)
					if err := loadCertificate(certPath, keyPath); err != nil {
						zap.L().Error("重新加载 TLS 证书失败",
							zap.String("cert_file", certPath),
							zap.String("key_file", keyPath),
							zap.String("event", event.String()),
							zap.Error(err),
						)
					} else {
						zap.L().Info("TLS 证书已重新加载",
							zap.String("cert_file", certPath),
							zap.String("key_file", keyPath),
							zap.String("event", event.String()),
						)
					}
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				zap.L().Error("TLS 证书文件监听错误", zap.Error(err))
			}
		}
	}()

	return ctx
}

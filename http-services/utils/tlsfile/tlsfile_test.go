package tlsfile

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net/http"
	"os"
	"testing"
	"time"

	"http-services/config"
)

// createTestCertFiles 动态生成自签名证书与私钥文件，用于单元测试。
func createTestCertFiles(t *testing.T) (string, string) {
	t.Helper()

	// 生成密钥
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("生成测试 RSA 密钥失败: %v", err)
	}

	// 生成自签名证书
	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			CommonName:   "localhost",
			Organization: []string{"http-services-test"},
		},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(time.Hour), // 有效期 1 小时即可
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		DNSNames:              []string{"localhost"},
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		t.Fatalf("生成测试证书失败: %v", err)
	}

	// 写入证书文件
	certFile, err := os.CreateTemp("", "tlsfile-test-cert-*.pem")
	if err != nil {
		t.Fatalf("创建测试证书文件失败: %v", err)
	}
	t.Cleanup(func() { _ = os.Remove(certFile.Name()) })

	if err := pem.Encode(certFile, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes}); err != nil {
		t.Fatalf("写入测试证书 PEM 失败: %v", err)
	}
	_ = certFile.Close()

	// 写入私钥文件
	keyFile, err := os.CreateTemp("", "tlsfile-test-key-*.pem")
	if err != nil {
		t.Fatalf("创建测试私钥文件失败: %v", err)
	}
	t.Cleanup(func() { _ = os.Remove(keyFile.Name()) })

	privBytes := x509.MarshalPKCS1PrivateKey(priv)
	if err := pem.Encode(keyFile, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: privBytes}); err != nil {
		t.Fatalf("写入测试私钥 PEM 失败: %v", err)
	}
	_ = keyFile.Close()

	return certFile.Name(), keyFile.Name()
}

// TestSetup_Disabled 验证在未启用 TLS 证书文件模式时不会修改 server。
func TestSetup_Disabled(t *testing.T) {
	config.EnableTLS = false
	config.TLSCertFile = ""
	config.TLSKeyFile = ""

	srv := &http.Server{}

	ctx := Setup(srv)
	if ctx == nil {
		t.Fatalf("Setup() 返回 nil")
	}
	if ctx.Enabled {
		t.Errorf("ctx.Enabled = true, want false when TLS disabled")
	}
	if srv.TLSConfig != nil {
		t.Errorf("srv.TLSConfig != nil, want nil when TLS disabled")
	}
}

// TestSetup_Enabled 验证在启用 TLS 证书文件模式时会为 server 配置 TLSConfig。
func TestSetup_Enabled(t *testing.T) {
	certPath, keyPath := createTestCertFiles(t)

	config.EnableTLS = true
	config.TLSCertFile = certPath
	config.TLSKeyFile = keyPath

	srv := &http.Server{
		Addr: ":0",
	}

	ctx := Setup(srv)
	if ctx == nil {
		t.Fatalf("Setup() 返回 nil")
	}
	if !ctx.Enabled {
		t.Fatalf("ctx.Enabled = false, want true when TLS enabled")
	}
	if srv.TLSConfig == nil {
		t.Fatalf("srv.TLSConfig is nil, want non-nil when TLS enabled")
	}

	// 验证可以通过 GetCertificate 获取证书
	if getCurrentCertificate() == nil {
		t.Fatalf("getCurrentCertificate() 返回 nil，期望已加载证书")
	}
	if _, err := srv.TLSConfig.GetCertificate(nil); err != nil {
		t.Fatalf("GetCertificate 返回错误: %v", err)
	}
}

// TestSetup_ReloadOnFileChange 验证当证书文件内容被外部修改时，会自动重新加载新证书。
func TestSetup_ReloadOnFileChange(t *testing.T) {
	certPath, keyPath := createTestCertFiles(t)

	config.EnableTLS = true
	config.TLSCertFile = certPath
	config.TLSKeyFile = keyPath

	srv := &http.Server{
		Addr: ":0",
	}

	ctx := Setup(srv)
	if ctx == nil {
		t.Fatalf("Setup() 返回 nil")
	}
	if !ctx.Enabled {
		t.Fatalf("ctx.Enabled = false, want true when TLS enabled")
	}

	// 获取初始证书指针
	originalCert := getCurrentCertificate()
	if originalCert == nil {
		t.Fatalf("初始证书为空，期望已加载证书")
	}

	// 生成新的证书与私钥，并覆盖到原路径，模拟外部更新
	newCertPath, newKeyPath := createTestCertFiles(t)
	newCertBytes, err := os.ReadFile(newCertPath)
	if err != nil {
		t.Fatalf("读取新证书文件失败: %v", err)
	}
	newKeyBytes, err := os.ReadFile(newKeyPath)
	if err != nil {
		t.Fatalf("读取新私钥文件失败: %v", err)
	}
	if err := os.WriteFile(certPath, newCertBytes, 0o600); err != nil {
		t.Fatalf("覆盖写入证书文件失败: %v", err)
	}
	if err := os.WriteFile(keyPath, newKeyBytes, 0o600); err != nil {
		t.Fatalf("覆盖写入私钥文件失败: %v", err)
	}

	// 等待后台 fsnotify 监听与重新加载逻辑生效
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		updated := getCurrentCertificate()
		if updated != nil && updated != originalCert {
			// 证书指针已发生变化，认为热更新成功
			return
		}
		time.Sleep(100 * time.Millisecond)
	}

	t.Fatalf("TLS 证书未在预期时间内完成热更新")
}

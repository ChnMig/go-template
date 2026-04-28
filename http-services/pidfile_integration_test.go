package main

import (
	"bytes"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestPidFileLifecycle(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过集成测试（-short）")
	}

	tmpDir := t.TempDir()
	binPath := filepath.Join(tmpDir, "http-services-testbin")

	buildCmd := exec.Command("go", "build", "-o", binPath, ".")
	var buildOut bytes.Buffer
	buildCmd.Stdout = &buildOut
	buildCmd.Stderr = &buildOut
	if err := buildCmd.Run(); err != nil {
		t.Fatalf("构建测试二进制失败: %v\n%s", err, buildOut.String())
	}

	pidPath := filepath.Join(tmpDir, "http-services.pid")
	port := reserveLocalPort(t)

	cmd := exec.Command(binPath, "--dev")
	cmd.Dir = tmpDir
	cmd.Env = append(os.Environ(),
		"HTTP_SERVICES_SERVER_PORT="+port,
		"HTTP_SERVICES_SERVER_SHUTDOWN_TIMEOUT=1s",
		"HTTP_SERVICES_JWT_KEY=0123456789abcdef0123456789abcdef",
	)

	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	if err := cmd.Start(); err != nil {
		t.Fatalf("启动服务失败: %v", err)
	}
	waitCh := make(chan error, 1)
	go func() { waitCh <- cmd.Wait() }()

	t.Cleanup(func() {
		if cmd.Process != nil && cmd.ProcessState == nil {
			_ = cmd.Process.Kill()
			_ = <-waitCh
		}
	})

	wantPID := strconv.Itoa(cmd.Process.Pid)
	deadline := time.Now().Add(10 * time.Second)
	for {
		data, err := os.ReadFile(pidPath)
		if err == nil {
			gotPID := strings.TrimSpace(string(data))
			if gotPID == wantPID {
				break
			}
		}

		select {
		case err := <-waitCh:
			t.Fatalf("服务在写入 pid 文件前退出: %v\nwant=%s，当前输出：\n%s", err, wantPID, out.String())
		default:
		}

		if time.Now().After(deadline) {
			t.Fatalf("pid 文件未在预期时间内写入或内容不正确，want=%s，当前输出：\n%s", wantPID, out.String())
		}
		time.Sleep(10 * time.Millisecond)
	}

	if err := cmd.Process.Signal(os.Interrupt); err != nil {
		t.Fatalf("发送中断信号失败: %v", err)
	}

	select {
	case err := <-waitCh:
		if err != nil {
			t.Fatalf("服务退出失败: %v\n输出：\n%s", err, out.String())
		}
	case <-time.After(5 * time.Second):
		_ = cmd.Process.Kill()
		_ = <-waitCh
		t.Fatalf("服务未在预期时间内退出\n输出：\n%s", out.String())
	}

	if _, err := os.Stat(pidPath); !os.IsNotExist(err) {
		t.Fatalf("服务退出后 pid 文件应被删除，Stat() err=%v", err)
	}
}

func reserveLocalPort(t *testing.T) string {
	t.Helper()

	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Skipf("当前环境无法分配本地监听端口，跳过进程生命周期集成测试: %v", err)
	}
	defer l.Close()

	addr, ok := l.Addr().(*net.TCPAddr)
	if !ok {
		t.Fatalf("监听地址类型 = %T，期望 *net.TCPAddr", l.Addr())
	}
	return strconv.Itoa(addr.Port)
}

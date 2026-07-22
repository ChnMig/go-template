// Package bootstrap 负责应用资源组装、运行和反向清理。
package bootstrap

import (
	"context"
	"errors"
	"net"
	"net/http"
	"reflect"
	"time"

	"http-services/config"
)

var (
	ErrInvalidDependencies = errors.New("bootstrap: invalid dependencies")
	ErrServeReturned       = errors.New("bootstrap: HTTP server returned unexpectedly")
)

// RuntimeConfig 是启动时冻结的 HTTP 生命周期配置。
type RuntimeConfig struct {
	Address         string
	PIDFile         string
	ShutdownTimeout time.Duration
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
	MaxHeaderBytes  int
	HTTP            config.HTTPConfig
}

// Server 是 bootstrap 管理的最小 HTTP server 接口。
type Server interface {
	Serve(net.Listener) error
	Shutdown(context.Context) error
	Close() error
}

// Dependencies 提供生产实现和测试替身之间的窄边界。
type Dependencies struct {
	Initialize func(bool) (RuntimeConfig, error)
	Migrate    func() error
	NewHandler func(RuntimeConfig) (http.Handler, error)
	NewServer  func(RuntimeConfig, http.Handler) Server
	Listen     func(string, string) (net.Listener, error)
	WritePID   func(string, int) error
	RemovePID  func(string) error
	Cleanup    func()
}

// Options 控制一次独立的应用运行。
type Options struct {
	Dependencies Dependencies
	PID          int
	Development  bool
	Migrate      bool
}

func validateDependencies(dependencies Dependencies) error {
	value := reflect.ValueOf(dependencies)
	for index := range value.NumField() {
		if value.Field(index).IsNil() {
			return errors.Join(ErrInvalidDependencies, errors.New(value.Type().Field(index).Name))
		}
	}
	return nil
}

func isNil(value any) bool {
	if value == nil {
		return true
	}
	reflected := reflect.ValueOf(value)
	return reflected.Kind() >= reflect.Chan && reflected.Kind() <= reflect.Slice && reflected.IsNil()
}

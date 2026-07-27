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

type RuntimeConfig struct {
	Address         string
	PIDFile         string
	MySQLDSN        string
	Redis           config.RedisConfig
	ShutdownTimeout time.Duration
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
	MaxHeaderBytes  int
	HTTP            config.HTTPConfig
}

type Resource interface {
	Ping(context.Context) error
	Close() error
}

type Resources struct {
	MySQL Resource
	Redis Resource
}

type Worker interface {
	Run(context.Context) error
}

type Server interface {
	Serve(net.Listener) error
	Shutdown(context.Context) error
	Close() error
}

type Dependencies struct {
	Initialize func(bool) (RuntimeConfig, error)
	NewMySQL   func(context.Context, string) (Resource, error)
	NewRedis   func(context.Context, config.RedisConfig) (Resource, error)
	Migrate    func(context.Context, Resource) error
	NewHandler func(RuntimeConfig, Resources) (http.Handler, error)
	NewWorker  func(RuntimeConfig, Resources) (Worker, error)
	NewServer  func(RuntimeConfig, http.Handler) Server
	Listen     func(string, string) (net.Listener, error)
	WritePID   func(string, int) error
	RemovePID  func(string, int) error
	Cleanup    func() error
}

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

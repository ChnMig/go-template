package bootstrap

import (
	"context"
	"errors"
	"net"
	"net/http"
	"sync"
	"testing"
	"time"

	"http-services/config"
)

var (
	errListen   = errors.New("listen failed")
	errPID      = errors.New("pid failed")
	errMigrate  = errors.New("migration failed")
	errServe    = errors.New("serve failed")
	errShutdown = errors.New("shutdown failed")
	errHandler  = errors.New("handler failed")
)

func TestRunMigrationInitializesMigratesAndCleansUp(t *testing.T) {
	// Given
	fixture := newFixture()
	fixture.migrateErr = errMigrate

	// When
	err := Run(t.Context(), Options{Development: true, Migrate: true, PID: 42, Dependencies: fixture.dependencies()})

	// Then
	if !errors.Is(err, errMigrate) {
		t.Fatalf("Run() error = %v, want %v", err, errMigrate)
	}
	fixture.requireEvents(t, "initialize", "mysql.new", "migrate", "mysql.close", "cleanup")
}

func TestRunHandlerFailureDoesNotStartServerOrListener(t *testing.T) {
	// Given
	fixture := newFixture()
	fixture.handlerErr = errHandler

	// When
	err := Run(t.Context(), Options{PID: 42, Dependencies: fixture.dependencies()})

	// Then
	if !errors.Is(err, errHandler) {
		t.Fatalf("Run() error = %v, want %v", err, errHandler)
	}
	fixture.requireEvents(t, "initialize", "mysql.new", "redis.new", "handler", "redis.close", "mysql.close", "cleanup")
}

func TestRunListenerFailureDoesNotWritePID(t *testing.T) {
	// Given
	fixture := newFixture()
	fixture.listenErr = errListen

	// When
	err := Run(t.Context(), Options{PID: 42, Dependencies: fixture.dependencies()})

	// Then
	if !errors.Is(err, errListen) {
		t.Fatalf("Run() error = %v, want %v", err, errListen)
	}
	fixture.requireEvents(
		t,
		"initialize", "mysql.new", "redis.new", "handler", "server", "listen",
		"redis.close", "mysql.close", "cleanup",
	)
}

func TestRunPIDFailureClosesAcquiredListener(t *testing.T) {
	// Given
	fixture := newFixture()
	fixture.pidErr = errPID

	// When
	err := Run(t.Context(), Options{PID: 42, Dependencies: fixture.dependencies()})

	// Then
	if !errors.Is(err, errPID) {
		t.Fatalf("Run() error = %v, want %v", err, errPID)
	}
	fixture.requireEvents(
		t,
		"initialize", "mysql.new", "redis.new", "handler", "server", "listen", "pid.write", "listener.close",
		"redis.close", "mysql.close", "cleanup",
	)
}

func TestRunCancellationShutsDownServerAndRemovesPID(t *testing.T) {
	// Given
	fixture := newFixture()
	ctx, cancel := context.WithCancel(t.Context())
	fixture.server.onServe = cancel

	// When
	err := Run(ctx, Options{PID: 42, Dependencies: fixture.dependencies()})
	// Then
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	fixture.requireEvents(
		t,
		"initialize", "mysql.new", "redis.new", "handler", "server", "listen", "pid.write", "serve", "shutdown",
		"listener.close", "pid.remove", "redis.close", "mysql.close", "cleanup",
	)
}

func TestRunUnexpectedServerReturnFailsAndCleansUp(t *testing.T) {
	// Given
	fixture := newFixture()
	fixture.server.returnImmediately = true
	fixture.server.serveErr = errServe

	// When
	err := Run(t.Context(), Options{PID: 42, Dependencies: fixture.dependencies()})

	// Then
	if !errors.Is(err, ErrServeReturned) || !errors.Is(err, errServe) {
		t.Fatalf("Run() error = %v, want ErrServeReturned and %v", err, errServe)
	}
	fixture.requireEvents(
		t,
		"initialize", "mysql.new", "redis.new", "handler", "server", "listen", "pid.write", "serve", "server.close",
		"listener.close", "pid.remove", "redis.close", "mysql.close", "cleanup",
	)
}

func TestRunShutdownFailureForcesCloseAndReturnsBothErrors(t *testing.T) {
	// Given
	fixture := newFixture()
	fixture.server.shutdownErr = errShutdown
	ctx, cancel := context.WithCancel(t.Context())
	fixture.server.onServe = cancel

	// When
	err := Run(ctx, Options{PID: 42, Dependencies: fixture.dependencies()})

	// Then
	if !errors.Is(err, errShutdown) {
		t.Fatalf("Run() error = %v, want %v", err, errShutdown)
	}
	fixture.requireEvents(
		t,
		"initialize", "mysql.new", "redis.new", "handler", "server", "listen", "pid.write", "serve", "shutdown",
		"server.close", "listener.close", "pid.remove", "redis.close", "mysql.close", "cleanup",
	)
}

type fixture struct {
	mu         sync.Mutex
	events     []string
	listenErr  error
	pidErr     error
	migrateErr error
	handlerErr error
	listener   *fakeListener
	server     *fakeServer
	mysql      *fakeResource
	redis      *fakeResource
}

func newFixture() *fixture {
	fixture := &fixture{}
	fixture.listener = &fakeListener{record: fixture.record}
	fixture.server = &fakeServer{record: fixture.record, done: make(chan struct{})}
	fixture.mysql = &fakeResource{name: "mysql", record: fixture.record}
	fixture.redis = &fakeResource{name: "redis", record: fixture.record}
	return fixture
}

func (f *fixture) dependencies() Dependencies {
	return Dependencies{
		Initialize: func(bool) (RuntimeConfig, error) {
			f.record("initialize")
			return RuntimeConfig{
				Address: "127.0.0.1:0", PIDFile: "service.pid", ShutdownTimeout: time.Second,
				MySQLDSN: "mysql-dsn", Redis: config.RedisConfig{Host: "redis:6379"},
			}, nil
		},
		NewMySQL: func(context.Context, string) (Resource, error) {
			f.record("mysql.new")
			return f.mysql, nil
		},
		NewRedis: func(context.Context, config.RedisConfig) (Resource, error) {
			f.record("redis.new")
			return f.redis, nil
		},
		Migrate: func(context.Context, Resource) error { f.record("migrate"); return f.migrateErr },
		NewHandler: func(RuntimeConfig, Resources) (http.Handler, error) {
			f.record("handler")
			return http.NewServeMux(), f.handlerErr
		},
		NewWorker: func(RuntimeConfig, Resources) (Worker, error) { return disabledWorker{}, nil },
		NewServer: func(RuntimeConfig, http.Handler) Server { f.record("server"); return f.server },
		Listen: func(string, string) (net.Listener, error) {
			f.record("listen")
			return f.listener, f.listenErr
		},
		WritePID:  func(string, int) error { f.record("pid.write"); return f.pidErr },
		RemovePID: func(string, int) error { f.record("pid.remove"); return nil },
		Cleanup:   func() error { f.record("cleanup"); return nil },
	}
}

type fakeResource struct {
	name   string
	record func(string)
}

func (*fakeResource) Ping(context.Context) error { return nil }
func (r *fakeResource) Close() error {
	r.record(r.name + ".close")
	return nil
}

func (f *fixture) record(event string) {
	f.mu.Lock()
	f.events = append(f.events, event)
	f.mu.Unlock()
}

func (f *fixture) requireEvents(t *testing.T, want ...string) {
	t.Helper()
	f.mu.Lock()
	defer f.mu.Unlock()
	if len(f.events) != len(want) {
		t.Fatalf("events = %v, want %v", f.events, want)
	}
	for index := range want {
		if f.events[index] != want[index] {
			t.Fatalf("events = %v, want %v", f.events, want)
		}
	}
}

type fakeListener struct {
	record func(string)
	once   sync.Once
}

func (l *fakeListener) Accept() (net.Conn, error) { return nil, net.ErrClosed }
func (l *fakeListener) Addr() net.Addr            { return fakeAddress("listener") }
func (l *fakeListener) Close() error {
	l.once.Do(func() { l.record("listener.close") })
	return nil
}

type fakeAddress string

func (a fakeAddress) Network() string { return string(a) }
func (a fakeAddress) String() string  { return string(a) }

type fakeServer struct {
	record            func(string)
	done              chan struct{}
	onServe           func()
	serveErr          error
	shutdownErr       error
	returnImmediately bool
	once              sync.Once
}

func (s *fakeServer) Serve(net.Listener) error {
	s.record("serve")
	if s.onServe != nil {
		s.onServe()
	}
	if s.returnImmediately {
		return s.serveErr
	}
	<-s.done
	return http.ErrServerClosed
}

func (s *fakeServer) Shutdown(context.Context) error {
	s.record("shutdown")
	if s.shutdownErr == nil {
		s.once.Do(func() { close(s.done) })
	}
	return s.shutdownErr
}

func (s *fakeServer) Close() error {
	s.record("server.close")
	s.once.Do(func() { close(s.done) })
	return nil
}

package bootstrap

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"reflect"

	"http-services/api"
	"http-services/config"
	"http-services/db"
	"http-services/db/msqldb"
	"http-services/db/rdb"
	"http-services/utils/log"
	"http-services/utils/pidfile"
	"http-services/utils/runmodel"

	"go.uber.org/zap"
)

func reflectDependenciesZero(dependencies Dependencies) bool {
	return reflect.ValueOf(dependencies).IsZero()
}

func defaultDependencies() Dependencies {
	var logWatcher *config.LogWatcher
	return Dependencies{
		Initialize: func(development bool) (RuntimeConfig, error) {
			runtimeConfig, watcher, err := initializeRuntime(development)
			if err == nil {
				logWatcher = watcher
			}
			return runtimeConfig, err
		},
		NewMySQL: func(ctx context.Context, dsn string) (Resource, error) {
			return msqldb.New(ctx, dsn)
		},
		NewRedis: func(ctx context.Context, redisConfig config.RedisConfig) (Resource, error) {
			return rdb.New(ctx, redisConfig)
		},
		Migrate: defaultMigrate,
		NewHandler: func(runtimeConfig RuntimeConfig, _ Resources) (http.Handler, error) {
			return api.NewRouter(api.DefaultOptions(runtimeConfig.HTTP))
		},
		NewWorker: func(RuntimeConfig, Resources) (Worker, error) { return disabledWorker{}, nil },
		NewServer: newHTTPServer,
		Listen:    net.Listen,
		WritePID:  pidfile.Write,
		RemovePID: pidfile.Remove,
		Cleanup: func() error {
			return cleanupRuntime(logWatcher)
		},
	}
}

func defaultMigrate(ctx context.Context, resource Resource) error {
	client, ok := resource.(*msqldb.Client)
	if !ok || client == nil || client.Database() == nil {
		return fmt.Errorf("%w: migration mysql", ErrInvalidDependencies)
	}
	return db.MigrateAll(ctx, client.Database())
}

func initializeRuntime(development bool) (RuntimeConfig, *config.LogWatcher, error) {
	cfg, err := config.Load()
	if err != nil {
		return RuntimeConfig{}, nil, fmt.Errorf("load config: %w", err)
	}
	runmodel.Detect(development)
	if err := log.SetLogger(cfg.Log, runmodel.IsDev(), os.Stdout); err != nil {
		return RuntimeConfig{}, nil, fmt.Errorf("initialize logger: %w", err)
	}
	log.StartMonitor()
	watcher, err := config.WatchLogConfig(cfg.Log, func(next config.LogConfig) error {
		config.UpdateLogConfig(next)
		if setErr := log.SetLogger(next, runmodel.IsDev(), os.Stdout); setErr != nil {
			return setErr
		}
		zap.L().Info("Log configuration reloaded",
			zap.String("level", string(next.Level)), zap.String("gin_level", string(next.GinLevel)))
		return nil
	})
	if err != nil {
		return RuntimeConfig{}, nil, errors.Join(fmt.Errorf("watch config: %w", err), log.Close())
	}
	return RuntimeConfig{
		Address: cfg.Server.Address(), PIDFile: cfg.Server.PIDFile,
		MySQLDSN: cfg.Database.MySQLDSN, Redis: cfg.Redis,
		ShutdownTimeout: cfg.Server.ShutdownTimeout, ReadTimeout: cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout, IdleTimeout: cfg.Server.IdleTimeout,
		MaxHeaderBytes: cfg.Server.MaxHeaderBytes, HTTP: cfg.Server.HTTPConfig(),
	}, watcher, nil
}

func newHTTPServer(runtimeConfig RuntimeConfig, handler http.Handler) Server {
	return &http.Server{
		Addr: runtimeConfig.Address, Handler: handler, ReadTimeout: runtimeConfig.ReadTimeout,
		WriteTimeout: runtimeConfig.WriteTimeout, IdleTimeout: runtimeConfig.IdleTimeout,
		MaxHeaderBytes: runtimeConfig.MaxHeaderBytes,
	}
}

func cleanupRuntime(watcher *config.LogWatcher) error {
	var watcherErr error
	if watcher != nil {
		watcherErr = watcher.Close()
	}
	return errors.Join(watcherErr, log.Close())
}

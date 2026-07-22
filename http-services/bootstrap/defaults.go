package bootstrap

import (
	"fmt"
	"net"
	"net/http"
	"reflect"

	"http-services/api"
	"http-services/config"
	"http-services/db"
	"http-services/db/msqldb"
	"http-services/db/rdb"
	"http-services/utils/log"
	"http-services/utils/pathtool"
	"http-services/utils/pidfile"
	"http-services/utils/runmodel"

	"go.uber.org/zap"
)

func reflectDependenciesZero(dependencies Dependencies) bool {
	return reflect.ValueOf(dependencies).IsZero()
}

func defaultDependencies() Dependencies {
	return Dependencies{
		Initialize: initializeRuntime,
		Migrate:    db.MigrateAll,
		NewHandler: func(runtimeConfig RuntimeConfig) (http.Handler, error) {
			return api.NewRouter(api.DefaultOptions(runtimeConfig.HTTP))
		},
		NewServer: newHTTPServer,
		Listen:    net.Listen,
		WritePID:  pidfile.Write,
		RemovePID: pidfile.Remove,
		Cleanup:   cleanupRuntime,
	}
}

func initializeRuntime(development bool) (RuntimeConfig, error) {
	if err := config.LoadConfig(); err != nil {
		return RuntimeConfig{}, fmt.Errorf("load config: %w", err)
	}
	runmodel.Detect(development)
	if config.RunModel == config.RunModelRelease {
		if err := pathtool.CreateDir(config.LogDir); err != nil {
			return RuntimeConfig{}, fmt.Errorf("create log directory: %w", err)
		}
	}
	log.GetLogger()
	log.StartMonitor()
	if err := config.ValidateConfig(config.JWTKey, int64(config.JWTExpiration)); err != nil {
		log.StopMonitor()
		return RuntimeConfig{}, fmt.Errorf("validate config: %w", err)
	}
	if err := config.WatchConfig(func() {
		log.SetLogger()
		logConfig := config.CurrentLogConfig()
		zap.L().Info("Log configuration reloaded",
			zap.String("level", logConfig.Level),
			zap.String("gin_level", logConfig.GinLevel),
		)
	}); err != nil {
		log.StopMonitor()
		return RuntimeConfig{}, fmt.Errorf("watch config: %w", err)
	}

	return RuntimeConfig{
		Address:         fmt.Sprintf(":%d", config.ListenPort),
		PIDFile:         config.PidFile,
		ShutdownTimeout: config.ShutdownTimeout,
		ReadTimeout:     config.ReadTimeout,
		WriteTimeout:    config.WriteTimeout,
		IdleTimeout:     config.IdleTimeout,
		MaxHeaderBytes:  config.MaxHeaderBytes,
		HTTP:            config.SnapshotHTTPConfig(),
	}, nil
}

func newHTTPServer(runtimeConfig RuntimeConfig, handler http.Handler) Server {
	return &http.Server{
		Addr:           runtimeConfig.Address,
		Handler:        handler,
		ReadTimeout:    runtimeConfig.ReadTimeout,
		WriteTimeout:   runtimeConfig.WriteTimeout,
		IdleTimeout:    runtimeConfig.IdleTimeout,
		MaxHeaderBytes: runtimeConfig.MaxHeaderBytes,
	}
}

func cleanupRuntime() {
	config.StopWatchConfig()
	msqldb.CloseClient()
	rdb.CloseClient()
	log.StopMonitor()
}

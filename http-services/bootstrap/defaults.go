package bootstrap

import (
	"fmt"
	"net"
	"net/http"
	"reflect"

	"http-services/api"
	"http-services/api/middleware"
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
		NewHandler: func() http.Handler { return api.InitApi() },
		NewServer:  newHTTPServer,
		Listen:     net.Listen,
		WritePID:   pidfile.Write,
		RemovePID:  pidfile.Remove,
		Cleanup:    cleanupRuntime,
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
	config.WatchConfig(func() {
		log.SetLogger()
		zap.L().Info("Configuration reloaded",
			zap.Int("port", config.ListenPort),
			zap.Duration("jwt_expiration", config.JWTExpiration),
			zap.Bool("rate_limit_enabled", config.EnableRateLimit),
		)
	})

	return RuntimeConfig{
		Address:         fmt.Sprintf(":%d", config.ListenPort),
		PIDFile:         config.PidFile,
		ShutdownTimeout: config.ShutdownTimeout,
		ReadTimeout:     config.ReadTimeout,
		WriteTimeout:    config.WriteTimeout,
		IdleTimeout:     config.IdleTimeout,
		MaxHeaderBytes:  config.MaxHeaderBytes,
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
	middleware.CleanupAllLimiters()
	msqldb.CloseClient()
	rdb.CloseClient()
	log.StopMonitor()
}

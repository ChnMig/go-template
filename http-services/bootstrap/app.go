package bootstrap

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strings"
)

type application struct {
	dependencies Dependencies
	runtime      RuntimeConfig
	mysql        Resource
	redis        Resource
	pid          int
}

func Run(ctx context.Context, options Options) error {
	if ctx == nil || options.PID <= 0 {
		return fmt.Errorf("%w: options", ErrInvalidDependencies)
	}
	dependencies := options.Dependencies
	if reflectDependenciesZero(dependencies) {
		dependencies = defaultDependencies()
	}
	if err := validateDependencies(dependencies); err != nil {
		return err
	}
	runtimeConfig, err := dependencies.Initialize(options.Development)
	if err != nil {
		return fmt.Errorf("bootstrap initialize: %w", err)
	}
	app := &application{dependencies: dependencies, runtime: runtimeConfig, pid: options.PID}
	if options.Migrate {
		if err := app.acquireMySQL(ctx); err != nil {
			return app.cleanup(err, false)
		}
		return app.cleanup(app.migrate(ctx), false)
	}
	if err := app.acquire(ctx); err != nil {
		return app.cleanup(err, false)
	}
	resources := Resources{MySQL: app.mysql, Redis: app.redis}
	handler, err := dependencies.NewHandler(runtimeConfig, resources)
	if err != nil {
		return app.cleanup(fmt.Errorf("bootstrap build handler: %w", err), false)
	}
	if isNil(handler) {
		return app.cleanup(fmt.Errorf("%w: handler", ErrInvalidDependencies), false)
	}
	worker, err := dependencies.NewWorker(runtimeConfig, resources)
	if err != nil {
		return app.cleanup(fmt.Errorf("bootstrap build worker: %w", err), false)
	}
	if isNil(worker) {
		return app.cleanup(fmt.Errorf("%w: worker", ErrInvalidDependencies), false)
	}
	server := dependencies.NewServer(runtimeConfig, handler)
	if isNil(server) {
		return app.cleanup(fmt.Errorf("%w: server", ErrInvalidDependencies), false)
	}
	listener, err := dependencies.Listen("tcp", runtimeConfig.Address)
	if err != nil {
		return app.cleanup(fmt.Errorf("bootstrap listen: %w", err), false)
	}
	if isNil(listener) {
		return app.cleanup(fmt.Errorf("%w: listener", ErrInvalidDependencies), false)
	}
	pidWritten := false
	if strings.TrimSpace(runtimeConfig.PIDFile) != "" {
		if err := dependencies.WritePID(runtimeConfig.PIDFile, options.PID); err != nil {
			operationErr := errors.Join(fmt.Errorf("bootstrap write pid: %w", err), closeListener(listener))
			return app.cleanup(operationErr, false)
		}
		pidWritten = true
	}
	operationErr := serveApplication(ctx, runtimeConfig.ShutdownTimeout, server, listener, worker)
	operationErr = errors.Join(operationErr, closeListener(listener))
	return app.cleanup(operationErr, pidWritten)
}

func (app *application) acquire(ctx context.Context) error {
	if err := app.acquireMySQL(ctx); err != nil {
		return err
	}
	if strings.TrimSpace(app.runtime.Redis.Host) == "" {
		return nil
	}
	resource, err := app.dependencies.NewRedis(ctx, app.runtime.Redis)
	if err != nil {
		return fmt.Errorf("bootstrap acquire redis: %w", err)
	}
	if isNil(resource) {
		return fmt.Errorf("%w: redis", ErrInvalidDependencies)
	}
	app.redis = resource
	return nil
}

func (app *application) acquireMySQL(ctx context.Context) error {
	if strings.TrimSpace(app.runtime.MySQLDSN) == "" {
		return nil
	}
	resource, err := app.dependencies.NewMySQL(ctx, app.runtime.MySQLDSN)
	if err != nil {
		return fmt.Errorf("bootstrap acquire mysql: %w", err)
	}
	if isNil(resource) {
		return fmt.Errorf("%w: mysql", ErrInvalidDependencies)
	}
	app.mysql = resource
	return nil
}

func (app *application) migrate(ctx context.Context) error {
	if app.mysql == nil {
		return fmt.Errorf("%w: migration mysql", ErrInvalidDependencies)
	}
	if err := app.dependencies.Migrate(ctx, app.mysql); err != nil {
		return fmt.Errorf("bootstrap migrate: %w", err)
	}
	return nil
}

func (app *application) cleanup(operationErr error, removePID bool) error {
	errs := []error{operationErr}
	if removePID {
		errs = append(errs, wrapOperation("remove pid", app.dependencies.RemovePID(app.runtime.PIDFile, app.pid)))
	}
	if app.redis != nil {
		errs = append(errs, wrapOperation("close redis", app.redis.Close()))
	}
	if app.mysql != nil {
		errs = append(errs, wrapOperation("close mysql", app.mysql.Close()))
	}
	errs = append(errs, wrapOperation("cleanup runtime", app.dependencies.Cleanup()))
	return errors.Join(errs...)
}

func closeListener(listener net.Listener) error {
	if err := listener.Close(); err != nil && !errors.Is(err, net.ErrClosed) {
		return fmt.Errorf("bootstrap close listener: %w", err)
	}
	return nil
}

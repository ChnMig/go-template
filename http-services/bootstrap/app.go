package bootstrap

import (
	"context"
	"errors"
	"fmt"
	"net"
)

// Run 初始化资源，执行迁移或运行 HTTP 服务，并在退出时反向清理。
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
	defer dependencies.Cleanup()

	if options.Migrate {
		if err := dependencies.Migrate(); err != nil {
			return fmt.Errorf("bootstrap migrate: %w", err)
		}
		return nil
	}

	handler, err := dependencies.NewHandler(runtimeConfig)
	if err != nil {
		return fmt.Errorf("bootstrap build handler: %w", err)
	}
	if isNil(handler) {
		return fmt.Errorf("%w: handler", ErrInvalidDependencies)
	}
	server := dependencies.NewServer(runtimeConfig, handler)
	if isNil(server) {
		return fmt.Errorf("%w: server", ErrInvalidDependencies)
	}
	listener, err := dependencies.Listen("tcp", runtimeConfig.Address)
	if err != nil {
		return fmt.Errorf("bootstrap listen: %w", err)
	}
	if isNil(listener) {
		return fmt.Errorf("%w: listener", ErrInvalidDependencies)
	}

	pidWritten := false
	if runtimeConfig.PIDFile != "" {
		if err := dependencies.WritePID(runtimeConfig.PIDFile, options.PID); err != nil {
			return errors.Join(fmt.Errorf("bootstrap write pid: %w", err), closeListener(listener))
		}
		pidWritten = true
	}

	operationErr := serve(ctx, runtimeConfig.ShutdownTimeout, server, listener)
	closeErr := closeListener(listener)
	var removeErr error
	if pidWritten {
		if err := dependencies.RemovePID(runtimeConfig.PIDFile); err != nil {
			removeErr = fmt.Errorf("bootstrap remove pid: %w", err)
		}
	}
	return errors.Join(operationErr, closeErr, removeErr)
}

func closeListener(listener net.Listener) error {
	if err := listener.Close(); err != nil && !errors.Is(err, net.ErrClosed) {
		return fmt.Errorf("bootstrap close listener: %w", err)
	}
	return nil
}

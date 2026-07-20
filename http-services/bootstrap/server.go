package bootstrap

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"time"
)

func serve(ctx context.Context, timeout time.Duration, server Server, listener net.Listener) error {
	serveResult := make(chan error, 1)
	go func() { serveResult <- server.Serve(listener) }()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), timeout)
		shutdownErr := server.Shutdown(shutdownCtx)
		cancel()
		var closeErr error
		if shutdownErr != nil {
			closeErr = wrapServerClose(server.Close())
		}
		serveErr := <-serveResult
		return errors.Join(wrapOperation("shutdown", shutdownErr), closeErr, classifyServe(serveErr, true))
	case serveErr := <-serveResult:
		return errors.Join(classifyServe(serveErr, false), wrapServerClose(server.Close()))
	}
}

func classifyServe(err error, controlled bool) error {
	if controlled && errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	if err == nil {
		return ErrServeReturned
	}
	return fmt.Errorf("%w: %w", ErrServeReturned, err)
}

func wrapServerClose(err error) error {
	if err == nil || errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return fmt.Errorf("bootstrap force close server: %w", err)
}

func wrapOperation(operation string, err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("bootstrap %s: %w", operation, err)
}

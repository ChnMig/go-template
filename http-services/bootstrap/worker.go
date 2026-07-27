package bootstrap

import (
	"context"
	"errors"
	"net"
	"time"
)

var errWorkerReturned = errors.New("bootstrap: worker returned unexpectedly")

type disabledWorker struct{}

func (disabledWorker) Run(ctx context.Context) error {
	<-ctx.Done()
	return nil
}

func serveApplication(ctx context.Context, timeout time.Duration, server Server, listener net.Listener, worker Worker) error {
	runCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	workerResult := make(chan error, 1)
	serverResult := make(chan error, 1)
	go func() { workerResult <- worker.Run(runCtx) }()
	go func() { serverResult <- serve(runCtx, timeout, server, listener) }()
	select {
	case workerErr := <-workerResult:
		cancel()
		serverErr := <-serverResult
		if workerErr == nil && ctx.Err() == nil {
			workerErr = errWorkerReturned
		}
		return errors.Join(workerErr, serverErr)
	case serverErr := <-serverResult:
		cancel()
		return errors.Join(serverErr, <-workerResult)
	}
}

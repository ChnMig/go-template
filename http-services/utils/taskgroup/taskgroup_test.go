package taskgroup

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestRun_ReturnsErrorsInTaskOrder(t *testing.T) {
	firstErr := errors.New("first failed")
	secondErr := errors.New("second failed")
	firstStarted := make(chan struct{})
	secondReturned := make(chan struct{})
	releaseFirst := make(chan struct{})
	result := make(chan []error, 1)

	go func() {
		result <- Run(context.Background(),
			CancelOnError("first", func(context.Context) error {
				close(firstStarted)
				<-releaseFirst
				return firstErr
			}),
			CancelOnError("second", func(context.Context) error {
				<-firstStarted
				close(secondReturned)
				return secondErr
			}),
		)
	}()

	<-secondReturned
	close(releaseFirst)
	errs := <-result
	if len(errs) != 2 || !errors.Is(errs[0], firstErr) || !errors.Is(errs[1], secondErr) {
		t.Fatalf("errors are not input-ordered: %#v", errs)
	}
}

func TestRun_CancelOnErrorCancelsPeerAndWaitsForExit(t *testing.T) {
	rootErr := errors.New("root failed")
	peerStarted := make(chan struct{})
	peerCanceled := make(chan struct{})
	releasePeer := make(chan struct{})
	result := make(chan []error, 1)

	go func() {
		result <- Run(context.Background(),
			CancelOnError("root", func(context.Context) error {
				<-peerStarted
				return rootErr
			}),
			CancelOnError("peer", func(ctx context.Context) error {
				close(peerStarted)
				<-ctx.Done()
				close(peerCanceled)
				<-releasePeer
				return ctx.Err()
			}),
		)
	}()

	<-peerCanceled
	select {
	case <-result:
		t.Fatal("Run returned before the canceled peer exited")
	default:
	}
	close(releasePeer)
	errs := <-result
	if !errors.Is(errs[0], rootErr) || errs[1] != nil {
		t.Fatalf("unexpected errors after internal cancellation: %#v", errs)
	}
}

func TestRun_ContinueOnErrorKeepsPeerRunning(t *testing.T) {
	softErr := errors.New("soft failed")
	softReturned := make(chan struct{})
	peerStarted := make(chan struct{})
	peerCanceled := make(chan struct{}, 1)
	releasePeer := make(chan struct{})
	result := make(chan []error, 1)

	go func() {
		result <- Run(context.Background(),
			ContinueOnError("soft", func(context.Context) error {
				close(softReturned)
				return softErr
			}),
			ContinueOnError("peer", func(ctx context.Context) error {
				close(peerStarted)
				select {
				case <-ctx.Done():
					peerCanceled <- struct{}{}
					return ctx.Err()
				case <-releasePeer:
					return nil
				}
			}),
		)
	}()

	<-softReturned
	<-peerStarted
	select {
	case <-peerCanceled:
		t.Fatal("ContinueOnError canceled its peer")
	case <-time.After(250 * time.Millisecond):
	}
	close(releasePeer)
	errs := <-result
	if !errors.Is(errs[0], softErr) || errs[1] != nil {
		t.Fatalf("unexpected continue-on-error results: %#v", errs)
	}
}

func TestRun_RecoversPanicWithMetadataAndCancelsPeer(t *testing.T) {
	panicValue := errors.New("worker exploded")
	peerStarted := make(chan struct{})
	peerCanceled := make(chan struct{})

	errs := Run(context.Background(),
		CancelOnError("panicking task", func(context.Context) error {
			<-peerStarted
			panic(panicValue)
		}),
		ContinueOnError("peer", func(ctx context.Context) error {
			close(peerStarted)
			<-ctx.Done()
			close(peerCanceled)
			return ctx.Err()
		}),
	)

	var panicErr *PanicError
	if !errors.As(errs[0], &panicErr) {
		t.Fatalf("expected PanicError, got %v", errs[0])
	}
	if panicErr.TaskIndex() != 0 || panicErr.TaskName() != "panicking task" || panicErr.Recovered() != panicValue || len(panicErr.Stack()) == 0 {
		t.Fatalf("unexpected panic metadata: %#v", panicErr)
	}
	stack := panicErr.Stack()
	stack[0] = 0
	if panicErr.Stack()[0] == 0 {
		t.Fatal("PanicError.Stack returned mutable internal storage")
	}
	if errs[1] != nil {
		t.Fatalf("peer cancellation replaced panic root: %v", errs[1])
	}
	select {
	case <-peerCanceled:
	default:
		t.Fatal("panic did not cancel peer")
	}
}

func TestRun_ContinueOnErrorStillCancelsForContextError(t *testing.T) {
	peerStarted := make(chan struct{})
	peerCanceled := make(chan struct{})

	errs := Run(context.Background(),
		ContinueOnError("deadline", func(context.Context) error {
			<-peerStarted
			return context.DeadlineExceeded
		}),
		ContinueOnError("peer", func(ctx context.Context) error {
			close(peerStarted)
			<-ctx.Done()
			close(peerCanceled)
			return ctx.Err()
		}),
	)

	if !errors.Is(errs[0], context.DeadlineExceeded) || errs[1] != nil {
		t.Fatalf("unexpected context-error results: %#v", errs)
	}
	select {
	case <-peerCanceled:
	default:
		t.Fatal("context error did not cancel peer")
	}
}

func TestRun_PreservesWorkerCancellationWhenParentIsCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	firstStarted := make(chan struct{})
	secondStarted := make(chan struct{})
	result := make(chan []error, 1)

	go func() {
		result <- Run(ctx,
			ContinueOnError("first", func(ctx context.Context) error {
				close(firstStarted)
				<-ctx.Done()
				return ctx.Err()
			}),
			ContinueOnError("second", func(ctx context.Context) error {
				close(secondStarted)
				<-ctx.Done()
				return ctx.Err()
			}),
		)
	}()

	<-firstStarted
	<-secondStarted
	cancel()
	errs := <-result
	if !errors.Is(errs[0], context.Canceled) || !errors.Is(errs[1], context.Canceled) {
		t.Fatalf("parent cancellation was filtered: %#v", errs)
	}
}

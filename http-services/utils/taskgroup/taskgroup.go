package taskgroup

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"runtime/debug"
	"sync"
)

// Task is a named concurrent operation constructed with CancelOnError or ContinueOnError.
type Task struct {
	name          string
	cancelOnError bool
	run           func(context.Context) error
}

// CancelOnError creates a task whose returned error cancels its peers.
func CancelOnError(name string, run func(context.Context) error) Task {
	return Task{name: name, cancelOnError: true, run: run}
}

// ContinueOnError creates a task whose ordinary returned error does not cancel its peers.
func ContinueOnError(name string, run func(context.Context) error) Task {
	return Task{name: name, run: run}
}

// PanicError records a recovered task panic and its stack.
type PanicError struct {
	taskIndex int
	taskName  string
	recovered any
	stack     []byte
}

func (e *PanicError) Error() string {
	return fmt.Sprintf("task %d (%q) panicked: %v\n%s", e.taskIndex, e.taskName, e.recovered, e.stack)
}

// TaskIndex returns the task's input position.
func (e *PanicError) TaskIndex() int {
	return e.taskIndex
}

// TaskName returns the task name supplied to its constructor.
func (e *PanicError) TaskName() string {
	return e.taskName
}

// Recovered returns the value captured by recover.
func (e *PanicError) Recovered() any {
	return e.recovered
}

// Stack returns a copy of the recovered panic stack.
func (e *PanicError) Stack() []byte {
	return bytes.Clone(e.stack)
}

// Run starts every task, waits for all of them, and returns errors in input order.
func Run(ctx context.Context, tasks ...Task) []error {
	taskCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	errs := make([]error, len(tasks))
	rootIndex := -1
	var cancelOnce sync.Once
	var waitGroup sync.WaitGroup
	waitGroup.Add(len(tasks))
	for index, task := range tasks {
		go func() {
			defer waitGroup.Done()
			err, panicked := call(taskCtx, index, task)
			errs[index] = err
			if err == nil || (!panicked && !task.cancelOnError && !isContextError(err)) {
				return
			}
			cancelOnce.Do(func() {
				rootIndex = index
				cancel()
			})
		}()
	}
	waitGroup.Wait()

	if ctx.Err() == nil && rootIndex >= 0 {
		for index, err := range errs {
			if index != rootIndex && errors.Is(err, context.Canceled) {
				errs[index] = nil
			}
		}
	}
	return errs
}

func call(ctx context.Context, index int, task Task) (err error, panicked bool) {
	defer func() {
		if recovered := recover(); recovered != nil {
			err = &PanicError{taskIndex: index, taskName: task.name, recovered: recovered, stack: debug.Stack()}
			panicked = true
		}
	}()
	return task.run(ctx), false
}

func isContextError(err error) bool {
	return errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded)
}

package apperrors

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"testing"
)

func TestIsClientCancellation(t *testing.T) {
	failure := errors.New("database unavailable")
	cases := []struct {
		name string
		err  error
		want bool
	}{
		{name: "nil"},
		{name: "canceled", err: context.Canceled, want: true},
		{name: "wrapped canceled", err: fmt.Errorf("query failed: %w", context.Canceled), want: true},
		{name: "joined canceled", err: errors.Join(context.Canceled, fmt.Errorf("query failed: %w", context.Canceled)), want: true},
		{name: "multiple wrapped canceled", err: fmt.Errorf("failures: %w / %w", context.Canceled, context.Canceled), want: true},
		{name: "nested canceled", err: fmt.Errorf("outer: %w", errors.Join(errors.Join(context.Canceled), context.Canceled)), want: true},
		{name: "custom canceled sentinel", err: cancellationError{}, want: true},
		{name: "custom wrapped sentinel", err: fmt.Errorf("query failed: %w", cancellationError{}), want: true},
		{name: "deadline", err: context.DeadlineExceeded},
		{name: "wrapped deadline", err: fmt.Errorf("query failed: %w", context.DeadlineExceeded)},
		{name: "canceled text", err: errors.New("context canceled")},
		{name: "wrapped canceled text", err: fmt.Errorf("query failed: %w", errors.New("context canceled"))},
		{name: "failure", err: failure},
		{name: "joined failure", err: errors.Join(context.Canceled, failure)},
		{name: "failure first", err: errors.Join(failure, context.Canceled)},
		{name: "joined deadline", err: errors.Join(context.Canceled, context.DeadlineExceeded)},
		{name: "multiple wrapped failure", err: fmt.Errorf("failures: %w / %w", context.Canceled, failure)},
		{name: "nested failure", err: fmt.Errorf("outer: %w", errors.Join(context.Canceled, errors.Join(context.Canceled, failure)))},
		{name: "custom match wrapping failure", err: cancellationWrapper{cause: failure}},
		{name: "nil join", err: errors.Join(nil)},
		{name: "nil children", err: multiError{nil, context.Canceled, nil}, want: true},
		{name: "only nil children", err: multiError{nil, nil}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := IsClientCancellation(tc.err); got != tc.want {
				t.Fatalf("IsClientCancellation(%v) = %v, want %v", tc.err, got, tc.want)
			}
		})
	}
}

func TestIsClientCancellation_TypedNil(t *testing.T) {
	cases := []struct {
		name string
		err  error
	}{
		{name: "url error", err: (*url.Error)(nil)},
		{name: "custom pointer", err: (*cancellationError)(nil)},
		{name: "map", err: cancellationMapError(nil)},
		{name: "slice", err: cancellationSliceError(nil)},
		{name: "func", err: cancellationFuncError(nil)},
		{name: "chan", err: cancellationChanError(nil)},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			variants := map[string]error{
				"direct":           tc.err,
				"wrapped":          fmt.Errorf("query failed: %w", tc.err),
				"joined":           errors.Join(context.Canceled, tc.err),
				"custom wrapped":   cancellationWrapper{cause: tc.err},
				"multiple wrapped": fmt.Errorf("failures: %w / %w", context.Canceled, tc.err),
			}
			for name, err := range variants {
				t.Run(name, func(t *testing.T) {
					if IsClientCancellation(err) {
						t.Fatal("typed nil errors must not be suppressed as cancellation")
					}
				})
			}
		})
	}
}

func TestIsClientCancellation_NonNilCustomSentinels(t *testing.T) {
	for _, err := range []error{
		&cancellationError{},
		cancellationMapError{},
		cancellationSliceError{},
		cancellationFuncError(func() {}),
		make(cancellationChanError),
	} {
		if !IsClientCancellation(err) {
			t.Fatalf("expected non-nil %T cancellation sentinel to match", err)
		}
	}
}

type cancellationError struct{}

func (cancellationError) Error() string       { return "client disconnected" }
func (cancellationError) Is(other error) bool { return other == context.Canceled }

type cancellationMapError map[string]bool

func (cancellationMapError) Error() string       { return "client disconnected" }
func (cancellationMapError) Is(other error) bool { return other == context.Canceled }

type cancellationSliceError []string

func (cancellationSliceError) Error() string       { return "client disconnected" }
func (cancellationSliceError) Is(other error) bool { return other == context.Canceled }

type cancellationFuncError func()

func (cancellationFuncError) Error() string       { return "client disconnected" }
func (cancellationFuncError) Is(other error) bool { return other == context.Canceled }

type cancellationChanError chan bool

func (cancellationChanError) Error() string       { return "client disconnected" }
func (cancellationChanError) Is(other error) bool { return other == context.Canceled }

type cancellationWrapper struct{ cause error }

func (cancellationWrapper) Error() string       { return "operation canceled" }
func (cancellationWrapper) Is(other error) bool { return other == context.Canceled }
func (e cancellationWrapper) Unwrap() error     { return e.cause }

type multiError []error

func (multiError) Error() string     { return "multiple errors" }
func (e multiError) Unwrap() []error { return e }

package log

import (
	"bytes"
	"strings"
	"sync/atomic"
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type countingCloser struct {
	closed atomic.Int32
}

func (closer *countingCloser) Close() error {
	closer.closed.Add(1)
	return nil
}

func TestManagedCoreReplaceKeepsDerivedLoggersLiveAndClosesOldWriter(t *testing.T) {
	managed := newManagedCore()
	logger := zap.New(managed).With(zap.String("trace_id", "trace-1"))
	firstOutput := &bytes.Buffer{}
	secondOutput := &bytes.Buffer{}
	firstCloser := &countingCloser{}
	secondCloser := &countingCloser{}

	if err := managed.replace(testJSONCore(firstOutput), firstCloser); err != nil {
		t.Fatalf("install first core: %v", err)
	}
	logger.Info("first")
	if err := managed.replace(testJSONCore(secondOutput), secondCloser); err != nil {
		t.Fatalf("install second core: %v", err)
	}
	logger.Info("second")

	if firstCloser.closed.Load() != 1 {
		t.Fatalf("first closer count = %d, want 1", firstCloser.closed.Load())
	}
	if secondCloser.closed.Load() != 0 {
		t.Fatalf("second closer count = %d, want 0", secondCloser.closed.Load())
	}
	if !strings.Contains(firstOutput.String(), "first") || strings.Contains(firstOutput.String(), "second") {
		t.Fatalf("unexpected first output: %s", firstOutput.String())
	}
	if !strings.Contains(secondOutput.String(), "second") || !strings.Contains(secondOutput.String(), "trace-1") {
		t.Fatalf("unexpected second output: %s", secondOutput.String())
	}

	if err := managed.close(); err != nil {
		t.Fatalf("close managed core: %v", err)
	}
	if secondCloser.closed.Load() != 1 {
		t.Fatalf("second closer count after close = %d, want 1", secondCloser.closed.Load())
	}
}

func testJSONCore(output *bytes.Buffer) zapcore.Core {
	return zapcore.NewCore(
		zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
		zapcore.AddSync(output),
		zap.DebugLevel,
	)
}

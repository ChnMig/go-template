package log

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"http-services/config"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func TestCancellationCoreFiltersOnlyCancellationErrors(t *testing.T) {
	var nilURLError *url.Error
	for _, tc := range []struct {
		name   string
		level  zapcore.Level
		fields []zap.Field
		logged bool
	}{
		{"canceled", zapcore.ErrorLevel, []zap.Field{zap.Error(context.Canceled)}, false},
		{"wrapped", zapcore.ErrorLevel, []zap.Field{zap.Error(fmt.Errorf("query: %w", context.Canceled))}, false},
		{"joined cancellations", zapcore.ErrorLevel, []zap.Field{zap.Error(errors.Join(context.Canceled, fmt.Errorf("peer: %w", context.Canceled)))}, false},
		{"warning", zapcore.WarnLevel, []zap.Field{zap.Error(context.Canceled)}, false},
		{"debug", zapcore.DebugLevel, []zap.Field{zap.Error(context.Canceled)}, true},
		{"info", zapcore.InfoLevel, []zap.Field{zap.Error(context.Canceled)}, true},
		{"dpanic", zapcore.DPanicLevel, []zap.Field{zap.Error(context.Canceled)}, true},
		{"deadline", zapcore.ErrorLevel, []zap.Field{zap.Error(context.DeadlineExceeded)}, true},
		{"unexpected", zapcore.ErrorLevel, []zap.Field{zap.Error(errors.New("database unavailable"))}, true},
		{"message only", zapcore.ErrorLevel, []zap.Field{zap.String("error", "context canceled")}, true},
		{"mixed join", zapcore.ErrorLevel, []zap.Field{zap.Error(errors.Join(context.Canceled, errors.New("database unavailable")))}, true},
		{"multiple errors", zapcore.ErrorLevel, []zap.Field{zap.Error(context.Canceled), zap.NamedError("cause", errors.New("database unavailable"))}, true},
		{"nil error", zapcore.ErrorLevel, []zap.Field{zap.Error(nil)}, true},
		{"typed nil error", zapcore.ErrorLevel, []zap.Field{zap.Error(nilURLError)}, true},
		{"canceled plus typed nil", zapcore.ErrorLevel, []zap.Field{zap.Error(context.Canceled), zap.NamedError("cause", nilURLError)}, true},
		{"panicking is", zapcore.ErrorLevel, []zap.Field{zap.Error(cancellationInspectionIsPanicError{})}, true},
		{"panicking unwrap", zapcore.ErrorLevel, []zap.Field{zap.Error(cancellationInspectionPanicError{})}, true},
		{"no error", zapcore.ErrorLevel, nil, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			core, output := cancellationTestCore(zapcore.DebugLevel)
			l := zap.New(newCancellationCore(core))
			l.Log(tc.level, "operation failed", tc.fields...)
			if got := output.Len() > 0; got != tc.logged {
				t.Fatalf("logged = %v, want %v: %s", got, tc.logged, output.String())
			}
			if tc.logged && cancellationLogEntries(t, output)[0].Level != tc.level.String() {
				t.Fatal("changed the level of a retained log")
			}
		})
	}
}

func TestCancellationCoreWithTypedNilError(t *testing.T) {
	var nilURLError *url.Error
	core, output := cancellationTestCore(zapcore.ErrorLevel)
	l := zap.New(newCancellationCore(core)).With(zap.Error(nilURLError))
	l.Error("typed nil failure")
	entries := cancellationLogEntries(t, output)
	if len(entries) != 1 || entries[0].Message != "typed nil failure" {
		t.Fatalf("typed nil error should remain visible without panicking: %s", output.String())
	}
}

type cancellationInspectionPanicError struct{}

func (cancellationInspectionPanicError) Error() string { return "unexpected failure" }
func (cancellationInspectionPanicError) Unwrap() error { panic("cannot inspect error") }

func TestCancellationCoreWithDoesNotHideOtherFailures(t *testing.T) {
	core, output := cancellationTestCore(zapcore.DebugLevel)
	base := zap.New(newCancellationCore(core))
	canceled := base.With(zap.Error(context.Canceled)).Named("query").With(zap.String("trace_id", "cancel-test"))
	canceled.Error("canceled query")
	if output.Len() != 0 {
		t.Fatal("error attached through With was not filtered")
	}
	canceled.Error("real failure", zap.NamedError("cause", errors.New("database unavailable")))
	canceled.With(zap.NamedError("cause", context.DeadlineExceeded)).Error("timeout")
	base.Error("parent failure")
	base.Sugar().Errorw("sugared cancellation", "error", context.Canceled)
	entries := cancellationLogEntries(t, output)
	if len(entries) != 3 {
		t.Fatalf("expected real failures to survive and parent to stay independent: %s", output.String())
	}
	entry := entries[0]
	if entry.LoggerName != "query" || entry.TraceID != "cancel-test" {
		t.Fatalf("lost bound request fields: %v", entry)
	}
}

func TestCancellationCorePreservesSamplingAndThreshold(t *testing.T) {
	core, output := cancellationTestCore(zapcore.ErrorLevel)
	l := zap.New(zapcore.NewSamplerWithOptions(newCancellationCore(core), time.Hour, 1, 0))
	l.Warn("below threshold", zap.Error(errors.New("unexpected")))
	l.Error("canceled", zap.Error(context.Canceled))
	l.Error("real failure", zap.Error(errors.New("database unavailable")))
	l.Error("real failure", zap.Error(errors.New("database unavailable")))
	entries := cancellationLogEntries(t, output)
	if len(entries) != 1 || entries[0].Message != "real failure" {
		t.Fatalf("sampling, threshold or cancellation filtering changed: %s", output.String())
	}
}

func cancellationTestCore(level zapcore.Level) (zapcore.Core, *bytes.Buffer) {
	var output bytes.Buffer
	core := zapcore.NewCore(zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()), zapcore.AddSync(&output), level)
	return core, &output
}

type cancellationTestEntry struct {
	Level      string `json:"level"`
	Message    string `json:"msg"`
	LoggerName string `json:"logger"`
	TraceID    string `json:"trace_id"`
}

func cancellationLogEntries(t *testing.T, output *bytes.Buffer) []cancellationTestEntry {
	t.Helper()
	var entries []cancellationTestEntry
	for _, line := range bytes.Split(bytes.TrimSpace(output.Bytes()), []byte("\n")) {
		if len(line) == 0 {
			continue
		}
		var entry cancellationTestEntry
		if err := json.Unmarshal(line, &entry); err != nil {
			t.Fatalf("decode log: %v: %s", err, line)
		}
		entries = append(entries, entry)
	}
	return entries
}

func TestProductLoggerFiltersCancellationAcrossLayers(t *testing.T) {
	oldMaxSize, oldMaxAge := config.LogMaxSize, config.LogMaxAge
	t.Cleanup(func() { config.LogMaxSize, config.LogMaxAge = oldMaxSize, oldMaxAge })
	config.LogMaxSize, config.LogMaxAge = 1, 1
	path := filepath.Join(t.TempDir(), "cancellation.log")
	l, writer := createProductLogger(path, zapcore.ErrorLevel)
	t.Cleanup(func() { _ = writer.Close() })
	undo := zap.ReplaceGlobals(l)
	t.Cleanup(undo)
	// 模拟同一次取消依次传播到 DB、领域与 handler。
	for _, layer := range []string{"db", "domain", "handler"} {
		zap.L().With(zap.String("layer", layer)).Error("operation failed", zap.Error(context.Canceled))
	}
	zap.L().Error("real failure", zap.Error(errors.New("database unavailable")))
	if err := l.Sync(); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(data), "operation failed") || !strings.Contains(string(data), `"level":"error"`) || !strings.Contains(string(data), "database unavailable") {
		t.Fatalf("unexpected production logs: %s", data)
	}
}

type cancellationInspectionIsPanicError struct{}

func (cancellationInspectionIsPanicError) Error() string { return "unexpected failure" }
func (cancellationInspectionIsPanicError) Is(error) bool { panic("cannot inspect error") }

func TestCancellationCoreWithPanickingError(t *testing.T) {
	for _, err := range []error{cancellationInspectionPanicError{}, cancellationInspectionIsPanicError{}} {
		core, output := cancellationTestCore(zapcore.DebugLevel)
		logger := zap.New(newCancellationCore(core)).With(zap.Error(err))
		logger.Error("real failure", zap.NamedError("cause", context.Canceled))
		if entries := cancellationLogEntries(t, output); len(entries) != 1 || entries[0].Level != "error" {
			t.Fatalf("inspection panic must preserve error entry: %s", output.String())
		}
	}
}

func TestCancellationCorePreservesCriticalEntries(t *testing.T) {
	for _, level := range []zapcore.Level{zapcore.DPanicLevel, zapcore.PanicLevel, zapcore.FatalLevel} {
		core, output := cancellationTestCore(zapcore.DebugLevel)
		if err := newCancellationCore(core).Write(zapcore.Entry{Level: level, Message: "critical"}, []zap.Field{zap.Error(context.Canceled)}); err != nil {
			t.Fatal(err)
		}
		if entries := cancellationLogEntries(t, output); len(entries) != 1 || entries[0].Level != level.String() {
			t.Fatalf("critical entry must remain visible: %s", output.String())
		}
	}
}

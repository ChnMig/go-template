package log

import (
	"io"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type flushCloser struct {
	flusher interface{ Flush() error }
}

func (closer flushCloser) Close() error {
	return closer.flusher.Flush()
}

func newFlushCloser(writer io.Writer) io.Closer {
	flusher, ok := writer.(interface{ Flush() error })
	if !ok {
		return nil
	}
	return flushCloser{flusher: flusher}
}

// zapWriter forwards framework-style io.Writer records into zap.
type zapWriter struct {
	getLogger func() *zap.Logger
	level     zapcore.Level
}

// Write implements io.Writer.
func (writer *zapWriter) Write(value []byte) (int, error) {
	if writer == nil || writer.getLogger == nil {
		return len(value), nil
	}
	logger := writer.getLogger()
	if logger == nil {
		return len(value), nil
	}
	message := strings.TrimRight(string(value), "\r\n")
	if checked := logger.Check(writer.level, message); checked != nil {
		checked.Write()
	}
	return len(value), nil
}

// NewZapWriter creates a writer backed by one logger.
func NewZapWriter(logger *zap.Logger, level zapcore.Level) *zapWriter {
	if logger == nil {
		return NewZapWriterFunc(GetLogger, level)
	}
	return NewZapWriterFunc(func() *zap.Logger { return logger }, level)
}

// NewZapWriterFunc creates a writer backed by a dynamic global logger provider.
func NewZapWriterFunc(getLogger func() *zap.Logger, level zapcore.Level) *zapWriter {
	if getLogger == nil {
		getLogger = GetLogger
	}
	return &zapWriter{getLogger: getLogger, level: level}
}

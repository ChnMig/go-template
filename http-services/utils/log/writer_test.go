package log

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

func Test_NewZapWriterFunc_uses_the_latest_logger_and_trims_line_endings(t *testing.T) {
	// Given
	firstCore, firstObserved := observer.New(zap.InfoLevel)
	current := zap.New(firstCore)
	writer := NewZapWriterFunc(func() *zap.Logger { return current }, zapcore.InfoLevel)
	firstLine := []byte("first line\r\n")

	// When
	written, err := writer.Write(firstLine)
	secondCore, secondObserved := observer.New(zap.InfoLevel)
	current = zap.New(secondCore)
	_, secondErr := writer.Write([]byte("second line\n"))

	// Then
	require.NoError(t, err)
	require.NoError(t, secondErr)
	require.Equal(t, len(firstLine), written)
	require.Len(t, firstObserved.All(), 1)
	require.Len(t, secondObserved.All(), 1)
	require.Equal(t, "first line", firstObserved.All()[0].Message)
	require.Equal(t, "second line", secondObserved.All()[0].Message)
}

package middleware_test

import (
	"bytes"
	"encoding/json"
	"testing"

	"http-services/api/middleware"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func requireSemanticEnvelope(t *testing.T, body []byte, code int, status string) {
	t.Helper()
	var envelope struct {
		Code      int    `json:"code"`
		Status    string `json:"status"`
		Timestamp int64  `json:"timestamp"`
	}
	require.NoError(t, json.Unmarshal(body, &envelope))
	require.Equal(t, code, envelope.Code)
	require.Equal(t, status, envelope.Status)
	require.Positive(t, envelope.Timestamp)
}

func newTestLogger(t *testing.T, output *bytes.Buffer) middleware.LoggerProvider {
	t.Helper()
	encoder := zap.NewProductionEncoderConfig()
	logger := zap.New(zapcore.NewCore(
		zapcore.NewJSONEncoder(encoder), zapcore.AddSync(output), zap.DebugLevel,
	))
	return func() *zap.Logger { return logger }
}

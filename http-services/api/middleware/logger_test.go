package middleware_test

import (
	"bytes"
	"encoding/json"
	"testing"

	"http-services/config"
	serviceLog "http-services/utils/log"

	"github.com/stretchr/testify/require"
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

func installGlobalTestLogger(t *testing.T, output *bytes.Buffer) {
	t.Helper()
	require.NoError(t, serviceLog.SetLogger(config.LogConfig{
		MaxSize: config.LogFileSizeMB(1), MaxAge: config.LogRetentionDays(1),
		Level: config.LogLevelDebug, GinLevel: config.LogLevelDebug,
	}, true, output))
	t.Cleanup(func() { require.NoError(t, serviceLog.Close()) })
}

package middleware_test

import (
	"bufio"
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"http-services/api/middleware"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

type logRecord struct {
	Message       string  `json:"msg"`
	TraceID       string  `json:"trace_id"`
	Method        string  `json:"method"`
	Path          string  `json:"path"`
	ClientIP      string  `json:"client_ip"`
	Status        float64 `json:"status"`
	ResponseBytes float64 `json:"response_bytes"`
}

func Test_AccessLog_writes_safe_structured_final_status(t *testing.T) {
	// Given
	var logOutput bytes.Buffer
	logger := newTestLogger(t, &logOutput)
	const traceID = "018f47a5-7b8c-7c11-8000-123456789abc"
	router := gin.New()
	router.Use(middleware.TraceID())
	router.Use(middleware.AccessLogWithLogger(logger))
	router.Use(middleware.RecoveryWithLogger(logger))
	router.POST("/probe", func(*gin.Context) { panic("secret-panic-value") })
	request := httptest.NewRequest(http.MethodPost, "/probe?token=query-secret", strings.NewReader("body-secret"))
	request.RemoteAddr = "192.0.2.10:4321"
	request.Header.Set(middleware.TraceIDHeader, traceID)
	request.Header.Set("Authorization", "Bearer header-secret")
	recorder := httptest.NewRecorder()

	// When
	router.ServeHTTP(recorder, request)

	// Then
	records := decodeLogRecords(t, logOutput.Bytes())
	require.Len(t, records, 2)
	accessRecord := records[1]
	require.Equal(t, "http.request", accessRecord.Message)
	require.Equal(t, traceID, accessRecord.TraceID)
	require.Equal(t, "POST", accessRecord.Method)
	require.Equal(t, "/probe", accessRecord.Path)
	require.Equal(t, "192.0.2.10", accessRecord.ClientIP)
	require.Equal(t, http.StatusOK, int(accessRecord.Status))
	require.NotContains(t, logOutput.String(), "query-secret")
	require.NotContains(t, logOutput.String(), "body-secret")
	require.NotContains(t, logOutput.String(), "header-secret")
}

func Test_AccessLog_excludes_health_path(t *testing.T) {
	// Given
	var logOutput bytes.Buffer
	logger := newTestLogger(t, &logOutput)
	router := gin.New()
	router.Use(middleware.AccessLogWithLogger(logger))
	router.GET("/api/v1/open/health", func(context *gin.Context) { context.Status(http.StatusOK) })

	// When
	router.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/api/v1/open/health", nil))

	// Then
	require.Empty(t, logOutput.String())
}

func Test_AccessLog_normalizes_unwritten_response_size_to_zero(t *testing.T) {
	// Given
	var logOutput bytes.Buffer
	logger := newTestLogger(t, &logOutput)
	router := gin.New()
	router.Use(middleware.AccessLogWithLogger(logger))
	router.GET("/status", func(context *gin.Context) { context.Status(http.StatusNoContent) })

	// When
	router.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/status", nil))

	// Then
	records := decodeLogRecords(t, logOutput.Bytes())
	require.Len(t, records, 1)
	require.Equal(t, 0, int(records[0].ResponseBytes))
}

func decodeLogRecords(t *testing.T, output []byte) []logRecord {
	t.Helper()
	records := make([]logRecord, 0, 2)
	scanner := bufio.NewScanner(bytes.NewReader(output))
	for scanner.Scan() {
		var record logRecord
		require.NoError(t, json.Unmarshal(scanner.Bytes(), &record))
		records = append(records, record)
	}
	require.NoError(t, scanner.Err())
	return records
}

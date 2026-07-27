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

func Test_AccessLog_writes_structured_final_status_with_global_zap(t *testing.T) {
	// Given
	var logOutput bytes.Buffer
	installGlobalTestLogger(t, &logOutput)
	const traceID = "018f47a5-7b8c-7c11-8000-123456789abc"
	router := gin.New()
	router.Use(middleware.TraceID())
	router.Use(middleware.AccessLog())
	router.Use(middleware.Recovery())
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
	accessRecord := findLogRecord(t, records, "http.request")
	require.Equal(t, "http.request", accessRecord.Message)
	require.Equal(t, traceID, accessRecord.TraceID)
	require.Equal(t, "POST", accessRecord.Method)
	require.Equal(t, "/probe", accessRecord.Path)
	require.Equal(t, "192.0.2.10", accessRecord.ClientIP)
	require.Equal(t, http.StatusOK, int(accessRecord.Status))
	require.Contains(t, logOutput.String(), "query-secret")
	require.NotContains(t, logOutput.String(), "body-secret")
	require.NotContains(t, logOutput.String(), "header-secret")
}

func Test_AccessLog_excludes_health_path(t *testing.T) {
	// Given
	var logOutput bytes.Buffer
	installGlobalTestLogger(t, &logOutput)
	router := gin.New()
	router.Use(middleware.AccessLog())
	router.GET("/api/v1/open/health", func(context *gin.Context) { context.Status(http.StatusOK) })

	// When
	router.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/api/v1/open/health", nil))

	// Then
	require.Empty(t, logOutput.String())
}

func Test_AccessLog_normalizes_unwritten_response_size_to_zero(t *testing.T) {
	// Given
	var logOutput bytes.Buffer
	installGlobalTestLogger(t, &logOutput)
	router := gin.New()
	router.Use(middleware.AccessLog())
	router.GET("/status", func(context *gin.Context) { context.Status(http.StatusNoContent) })

	// When
	router.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/status", nil))

	// Then
	records := decodeLogRecords(t, logOutput.Bytes())
	require.Equal(t, 0, int(findLogRecord(t, records, "http.request").ResponseBytes))
}

func decodeLogRecords(t *testing.T, output []byte) []logRecord {
	t.Helper()
	records := make([]logRecord, 0, 2)
	scanner := bufio.NewScanner(bytes.NewReader(output))
	for scanner.Scan() {
		line := scanner.Bytes()
		start := bytes.IndexByte(line, '{')
		if start < 0 {
			continue
		}
		var record logRecord
		require.NoError(t, json.Unmarshal(line[start:], &record))
		for _, message := range []string{
			"http.request.started", "http.request.completed", "http.request",
			"http.panic_recovered", "http.connection_aborted", "Returning error response",
		} {
			if bytes.Contains(line[:start], []byte(message)) {
				record.Message = message
				break
			}
		}
		records = append(records, record)
	}
	require.NoError(t, scanner.Err())
	return records
}

func findLogRecord(t *testing.T, records []logRecord, message string) logRecord {
	t.Helper()
	for _, record := range records {
		if record.Message == message {
			return record
		}
	}
	t.Fatalf("log record %q not found", message)
	return logRecord{}
}

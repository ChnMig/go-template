package middleware_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"http-services/api/middleware"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func Test_Recovery_classifies_aborted_connections_without_stack_or_response(t *testing.T) {
	// Given
	var logOutput bytes.Buffer
	installGlobalTestLogger(t, &logOutput)
	router := gin.New()
	router.Use(middleware.Recovery())
	router.GET("/panic", func(*gin.Context) { panic(http.ErrAbortHandler) })
	recorder := httptest.NewRecorder()

	// When
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/panic", nil))

	// Then
	require.Empty(t, recorder.Body.String())
	require.Contains(t, logOutput.String(), "http.connection_aborted")
	require.NotContains(t, logOutput.String(), `"stack"`)
}

func Test_Recovery_returns_internal_envelope_before_commit(t *testing.T) {
	// Given
	var logOutput bytes.Buffer
	installGlobalTestLogger(t, &logOutput)
	const traceID = "018f47a5-7b8c-7c11-8000-123456789abc"
	router := gin.New()
	router.Use(middleware.TraceID())
	router.Use(middleware.Recovery())
	router.GET("/panic", func(*gin.Context) { panic("boom") })
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/panic", nil)
	request.Header.Set(middleware.TraceIDHeader, traceID)

	// When
	router.ServeHTTP(recorder, request)

	// Then
	require.Equal(t, http.StatusOK, recorder.Code)
	requireSemanticEnvelope(t, recorder.Body.Bytes(), http.StatusInternalServerError, "INTERNAL")
	require.Equal(t, traceID, recorder.Header().Get(middleware.TraceIDHeader))
	require.Equal(t, 1, strings.Count(logOutput.String(), "http.panic_recovered"))
	records := decodeLogRecords(t, logOutput.Bytes())
	require.Equal(t, http.StatusInternalServerError, int(findLogRecord(t, records, "http.panic_recovered").Status))
}

func Test_Recovery_preserves_committed_response_and_logs_once(t *testing.T) {
	// Given
	var logOutput bytes.Buffer
	installGlobalTestLogger(t, &logOutput)
	router := gin.New()
	router.Use(middleware.Recovery())
	router.GET("/panic", func(context *gin.Context) {
		context.String(http.StatusAccepted, "partial")
		panic("boom")
	})
	recorder := httptest.NewRecorder()

	// When
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/panic", nil))

	// Then
	require.Equal(t, http.StatusAccepted, recorder.Code)
	require.Equal(t, "partial", recorder.Body.String())
	require.Equal(t, 1, strings.Count(logOutput.String(), "http.panic_recovered"))
}

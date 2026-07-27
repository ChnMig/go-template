package middleware_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"http-services/api/middleware"
	serviceLog "http-services/utils/log"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func Test_TraceID_stores_global_request_context_logger(t *testing.T) {
	// Given
	var output bytes.Buffer
	installGlobalTestLogger(t, &output)
	router := gin.New()
	router.Use(middleware.TraceID())
	router.GET("/probe", func(context *gin.Context) {
		serviceLog.FromContext(context).Info("handler event")
		context.Status(http.StatusNoContent)
	})
	recorder := httptest.NewRecorder()

	// When
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/probe?token=secret", nil))

	// Then
	generated := recorder.Header().Get(middleware.TraceIDHeader)
	requireUUIDv7(t, generated)
	require.Contains(t, output.String(), generated)
	require.Contains(t, output.String(), `"method": "GET"`)
	require.Contains(t, output.String(), `"path": "/probe"`)
	require.NotContains(t, output.String(), "secret")
}

func Test_TraceID_preserves_valid_inbound_UUID(t *testing.T) {
	// Given
	validTraceID := "018f47a5-7b8c-7c11-8000-123456789abc"
	router := gin.New()
	router.Use(middleware.TraceID())
	router.GET("/probe", func(context *gin.Context) {
		traceID, ok := serviceLog.TraceID(context.Request.Context())
		require.True(t, ok)
		context.String(http.StatusOK, traceID)
	})
	request := httptest.NewRequest(http.MethodGet, "/probe", nil)
	request.Header.Set(middleware.TraceIDHeader, validTraceID)
	recorder := httptest.NewRecorder()

	// When
	router.ServeHTTP(recorder, request)

	// Then
	require.Equal(t, http.StatusOK, recorder.Code)
	require.Equal(t, validTraceID, recorder.Header().Get(middleware.TraceIDHeader))
	require.Equal(t, validTraceID, recorder.Body.String())
}

func Test_TraceID_replaces_malformed_header_with_UUIDv7(t *testing.T) {
	// Given
	router := gin.New()
	router.Use(middleware.TraceID())
	router.GET("/probe", func(context *gin.Context) { context.Status(http.StatusNoContent) })
	request := httptest.NewRequest(http.MethodGet, "/probe", nil)
	request.Header.Set(middleware.TraceIDHeader, "not-a-uuid")
	recorder := httptest.NewRecorder()

	// When
	router.ServeHTTP(recorder, request)

	// Then
	require.Equal(t, http.StatusNoContent, recorder.Code)
	requireUUIDv7(t, recorder.Header().Get(middleware.TraceIDHeader))
}

func Test_TraceID_replaces_noncanonical_UUID_header(t *testing.T) {
	// Given
	router := gin.New()
	router.Use(middleware.TraceID())
	router.GET("/probe", func(context *gin.Context) { context.Status(http.StatusNoContent) })
	request := httptest.NewRequest(http.MethodGet, "/probe", nil)
	request.Header.Set(middleware.TraceIDHeader, "018f47a57b8c7c118000123456789abc")
	recorder := httptest.NewRecorder()

	// When
	router.ServeHTTP(recorder, request)

	// Then
	requireUUIDv7(t, recorder.Header().Get(middleware.TraceIDHeader))
}

func Test_TraceID_replaces_other_noncanonical_UUID_forms(t *testing.T) {
	for _, inbound := range []string{
		"urn:uuid:018f47a5-7b8c-7c11-8000-123456789abc",
		"{018f47a5-7b8c-7c11-8000-123456789abc}",
	} {
		t.Run(inbound, func(t *testing.T) {
			// Given
			router := gin.New()
			router.Use(middleware.TraceID())
			router.GET("/probe", func(context *gin.Context) { context.Status(http.StatusNoContent) })
			request := httptest.NewRequest(http.MethodGet, "/probe", nil)
			request.Header.Set(middleware.TraceIDHeader, inbound)
			recorder := httptest.NewRecorder()

			// When
			router.ServeHTTP(recorder, request)

			// Then
			requireUUIDv7(t, recorder.Header().Get(middleware.TraceIDHeader))
		})
	}
}

func Test_TraceID_preserves_uppercase_canonical_UUID(t *testing.T) {
	// Given
	inbound := "018F47A5-7B8C-7C11-8000-123456789ABC"
	router := gin.New()
	router.Use(middleware.TraceID())
	router.GET("/probe", func(context *gin.Context) { context.Status(http.StatusNoContent) })
	request := httptest.NewRequest(http.MethodGet, "/probe", nil)
	request.Header.Set(middleware.TraceIDHeader, inbound)
	recorder := httptest.NewRecorder()

	// When
	router.ServeHTTP(recorder, request)

	// Then
	require.Equal(t, inbound, recorder.Header().Get(middleware.TraceIDHeader))
}

func requireUUIDv7(t *testing.T, value string) {
	t.Helper()
	parsed, err := uuid.Parse(value)
	require.NoError(t, err)
	require.Len(t, value, 36)
	require.Equal(t, uuid.Version(7), parsed.Version())
	require.Equal(t, uuid.RFC4122, parsed.Variant())
}

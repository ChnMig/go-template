package middleware_test

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"http-services/api/middleware"
	serviceLog "http-services/utils/log"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func Test_TraceID_stores_safe_request_context_logger(t *testing.T) {
	var output bytes.Buffer
	generated := uuid.MustParse("018f47a5-7b8c-7c11-8000-123456789abc")
	router := gin.New()
	router.Use(middleware.TraceIDWithDependencies(
		func() (uuid.UUID, error) { return generated, nil }, newTestLogger(t, &output),
	))
	router.GET("/probe", func(context *gin.Context) {
		serviceLog.FromContext(context).Info("handler event")
		context.Status(http.StatusNoContent)
	})
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/probe?token=secret", nil))

	require.Contains(t, output.String(), generated.String())
	require.Contains(t, output.String(), `"method":"GET"`)
	require.Contains(t, output.String(), `"path":"/probe"`)
	require.NotContains(t, output.String(), "secret")
}

func Test_TraceID_preserves_valid_inbound_UUID(t *testing.T) {
	// Given
	validTraceID := "018f47a5-7b8c-7c11-8000-123456789abc"
	factoryCalled := false
	router := gin.New()
	router.Use(middleware.TraceID(func() (uuid.UUID, error) {
		factoryCalled = true
		return uuid.Nil, nil
	}))
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
	require.False(t, factoryCalled)
}

func Test_TraceID_replaces_malformed_header_with_UUIDv7(t *testing.T) {
	// Given
	generated := uuid.MustParse("018f47a5-7b8c-7c11-8000-123456789abc")
	router := gin.New()
	router.Use(middleware.TraceID(func() (uuid.UUID, error) { return generated, nil }))
	router.GET("/probe", func(context *gin.Context) { context.Status(http.StatusNoContent) })
	request := httptest.NewRequest(http.MethodGet, "/probe", nil)
	request.Header.Set(middleware.TraceIDHeader, "not-a-uuid")
	recorder := httptest.NewRecorder()

	// When
	router.ServeHTTP(recorder, request)

	// Then
	require.Equal(t, http.StatusNoContent, recorder.Code)
	require.Equal(t, generated.String(), recorder.Header().Get(middleware.TraceIDHeader))
}

func Test_TraceID_replaces_noncanonical_UUID_header(t *testing.T) {
	// Given
	generated := uuid.MustParse("018f47a5-7b8c-7c11-8000-123456789abc")
	router := gin.New()
	router.Use(middleware.TraceID(func() (uuid.UUID, error) { return generated, nil }))
	router.GET("/probe", func(context *gin.Context) { context.Status(http.StatusNoContent) })
	request := httptest.NewRequest(http.MethodGet, "/probe", nil)
	request.Header.Set(middleware.TraceIDHeader, "018f47a57b8c7c118000123456789abc")
	recorder := httptest.NewRecorder()

	// When
	router.ServeHTTP(recorder, request)

	// Then
	require.Equal(t, generated.String(), recorder.Header().Get(middleware.TraceIDHeader))
}

func Test_TraceID_replaces_other_noncanonical_UUID_forms(t *testing.T) {
	for _, inbound := range []string{
		"urn:uuid:018f47a5-7b8c-7c11-8000-123456789abc",
		"{018f47a5-7b8c-7c11-8000-123456789abc}",
	} {
		t.Run(inbound, func(t *testing.T) {
			// Given
			generated := uuid.MustParse("019f4ace-61ab-7382-899a-8f94844f0135")
			router := gin.New()
			router.Use(middleware.TraceID(func() (uuid.UUID, error) { return generated, nil }))
			router.GET("/probe", func(context *gin.Context) { context.Status(http.StatusNoContent) })
			request := httptest.NewRequest(http.MethodGet, "/probe", nil)
			request.Header.Set(middleware.TraceIDHeader, inbound)
			recorder := httptest.NewRecorder()

			// When
			router.ServeHTTP(recorder, request)

			// Then
			require.Equal(t, generated.String(), recorder.Header().Get(middleware.TraceIDHeader))
		})
	}
}

func Test_TraceID_preserves_uppercase_canonical_UUID(t *testing.T) {
	// Given
	inbound := "018F47A5-7B8C-7C11-8000-123456789ABC"
	router := gin.New()
	router.Use(middleware.TraceID(func() (uuid.UUID, error) { return uuid.Nil, errors.New("must not generate") }))
	router.GET("/probe", func(context *gin.Context) { context.Status(http.StatusNoContent) })
	request := httptest.NewRequest(http.MethodGet, "/probe", nil)
	request.Header.Set(middleware.TraceIDHeader, inbound)
	recorder := httptest.NewRecorder()

	// When
	router.ServeHTTP(recorder, request)

	// Then
	require.Equal(t, inbound, recorder.Header().Get(middleware.TraceIDHeader))
}

func Test_TraceID_returns_internal_error_without_downstream_or_header_when_generation_fails(t *testing.T) {
	// Given
	downstreamCalled := false
	router := gin.New()
	router.Use(middleware.TraceID(func() (uuid.UUID, error) { return uuid.Nil, errors.New("entropy unavailable") }))
	router.GET("/probe", func(context *gin.Context) {
		downstreamCalled = true
		context.Status(http.StatusNoContent)
	})
	recorder := httptest.NewRecorder()

	// When
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/probe", nil))

	// Then
	require.Equal(t, http.StatusOK, recorder.Code)
	requireSemanticEnvelope(t, recorder.Body.Bytes(), http.StatusInternalServerError, "INTERNAL")
	require.Empty(t, recorder.Header().Get(middleware.TraceIDHeader))
	require.False(t, downstreamCalled)
}

package middleware_test

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"http-services/api/middleware"
	"http-services/config"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

type failingReadCloser struct {
	err error
}

func (reader failingReadCloser) Read([]byte) (int, error) {
	return 0, reader.err
}

func (failingReadCloser) Close() error {
	return nil
}

func Test_BodySizeLimit_rejects_known_oversize_before_handler(t *testing.T) {
	// Given
	installGlobalTestLogger(t, &bytes.Buffer{})
	handlerCalled := false
	router := gin.New()
	router.Use(middleware.BodySizeLimit(config.ByteSize(4)))
	router.POST("/probe", func(context *gin.Context) {
		handlerCalled = true
		context.Status(http.StatusNoContent)
	})
	request := httptest.NewRequest(http.MethodPost, "/probe", strings.NewReader("12345"))
	recorder := httptest.NewRecorder()

	// When
	router.ServeHTTP(recorder, request)

	// Then
	require.Equal(t, http.StatusOK, recorder.Code)
	requireSemanticEnvelope(t, recorder.Body.Bytes(), http.StatusBadRequest, "INVALID_ARGUMENT")
	require.False(t, handlerCalled)
}

func Test_BodySizeLimit_converts_propagated_stream_error_before_commit(t *testing.T) {
	// Given
	installGlobalTestLogger(t, &bytes.Buffer{})
	handlerCompleted := false
	router := gin.New()
	router.Use(middleware.BodySizeLimit(config.ByteSize(4)))
	router.POST("/probe", func(context *gin.Context) {
		_, err := io.ReadAll(context.Request.Body)
		if err != nil {
			require.NotNil(t, context.Error(err))
			return
		}
		handlerCompleted = true
	})
	request := httptest.NewRequest(http.MethodPost, "/probe", strings.NewReader("12345"))
	request.ContentLength = -1
	recorder := httptest.NewRecorder()

	// When
	router.ServeHTTP(recorder, request)

	// Then
	require.Equal(t, http.StatusOK, recorder.Code)
	requireSemanticEnvelope(t, recorder.Body.Bytes(), http.StatusBadRequest, "INVALID_ARGUMENT")
	require.False(t, handlerCompleted)
}

func Test_BodySizeLimit_aborts_later_handlers_when_stream_read_exceeds_limit(t *testing.T) {
	// Given
	installGlobalTestLogger(t, &bytes.Buffer{})
	secondHandlerCalled := false
	router := gin.New()
	router.Use(middleware.BodySizeLimit(config.ByteSize(4)))
	router.POST(
		"/probe",
		func(context *gin.Context) {
			_, err := io.ReadAll(context.Request.Body)
			require.Error(t, err)
		},
		func(context *gin.Context) {
			secondHandlerCalled = true
			context.String(299, "second-handler-ran")
		},
	)
	request := httptest.NewRequest(http.MethodPost, "/probe", strings.NewReader("12345"))
	request.ContentLength = -1
	recorder := httptest.NewRecorder()

	// When
	router.ServeHTTP(recorder, request)

	// Then
	require.Equal(t, http.StatusOK, recorder.Code)
	requireSemanticEnvelope(t, recorder.Body.Bytes(), http.StatusBadRequest, "INVALID_ARGUMENT")
	require.False(t, secondHandlerCalled)
}

func Test_BodySizeLimit_preserves_committed_response_and_logs_once(t *testing.T) {
	// Given
	var logOutput bytes.Buffer
	installGlobalTestLogger(t, &logOutput)
	router := gin.New()
	router.Use(middleware.BodySizeLimit(config.ByteSize(4)))
	router.POST("/probe", func(context *gin.Context) {
		context.String(http.StatusAccepted, "partial")
		_, err := io.ReadAll(context.Request.Body)
		if err != nil {
			require.NotNil(t, context.Error(err))
		}
	})
	request := httptest.NewRequest(http.MethodPost, "/probe", strings.NewReader("12345"))
	request.ContentLength = -1
	recorder := httptest.NewRecorder()

	// When
	router.ServeHTTP(recorder, request)

	// Then
	require.Equal(t, http.StatusAccepted, recorder.Code)
	require.Equal(t, "partial", recorder.Body.String())
	require.Equal(t, 1, strings.Count(logOutput.String(), "http.body_too_large_after_commit"))
}

func Test_BodySizeLimit_converts_stream_error_without_context_error(t *testing.T) {
	// Given
	installGlobalTestLogger(t, &bytes.Buffer{})
	router := gin.New()
	router.Use(middleware.BodySizeLimit(config.ByteSize(4)))
	router.POST("/probe", func(context *gin.Context) {
		_, err := io.ReadAll(context.Request.Body)
		require.Error(t, err)
	})
	request := httptest.NewRequest(http.MethodPost, "/probe", strings.NewReader("12345"))
	request.ContentLength = -1
	recorder := httptest.NewRecorder()

	// When
	router.ServeHTTP(recorder, request)

	// Then
	require.Equal(t, http.StatusOK, recorder.Code)
	requireSemanticEnvelope(t, recorder.Body.Bytes(), http.StatusBadRequest, "INVALID_ARGUMENT")
}

func Test_BodySizeLimit_aborts_later_handlers_after_committed_stream_overflow(t *testing.T) {
	// Given
	var logOutput bytes.Buffer
	installGlobalTestLogger(t, &logOutput)
	secondHandlerCalled := false
	router := gin.New()
	router.Use(middleware.BodySizeLimit(config.ByteSize(4)))
	router.POST(
		"/probe",
		func(context *gin.Context) {
			context.String(http.StatusAccepted, "partial")
			_, err := io.ReadAll(context.Request.Body)
			require.Error(t, err)
		},
		func(context *gin.Context) {
			secondHandlerCalled = true
			context.String(299, "second-handler-ran")
		},
	)
	request := httptest.NewRequest(http.MethodPost, "/probe", strings.NewReader("12345"))
	request.ContentLength = -1
	recorder := httptest.NewRecorder()

	// When
	router.ServeHTTP(recorder, request)

	// Then
	require.Equal(t, http.StatusAccepted, recorder.Code)
	require.Equal(t, "partial", recorder.Body.String())
	require.False(t, secondHandlerCalled)
	require.Equal(t, 1, strings.Count(logOutput.String(), "http.body_too_large_after_commit"))
}

func Test_BodySizeLimit_does_not_abort_on_non_limit_read_error(t *testing.T) {
	// Given
	readErr := errors.New("upstream read failed")
	installGlobalTestLogger(t, &bytes.Buffer{})
	secondHandlerCalled := false
	router := gin.New()
	router.Use(middleware.BodySizeLimit(config.ByteSize(4)))
	router.POST(
		"/probe",
		func(context *gin.Context) {
			_, err := io.ReadAll(context.Request.Body)
			require.ErrorIs(t, err, readErr)
		},
		func(context *gin.Context) {
			secondHandlerCalled = true
			context.String(299, "second-handler-ran")
		},
	)
	request := httptest.NewRequest(http.MethodPost, "/probe", failingReadCloser{err: readErr})
	request.ContentLength = -1
	recorder := httptest.NewRecorder()

	// When
	router.ServeHTTP(recorder, request)

	// Then
	require.Equal(t, 299, recorder.Code)
	require.Equal(t, "second-handler-ran", recorder.Body.String())
	require.True(t, secondHandlerCalled)
}

func Test_BodySizeLimit_accepts_bodies_at_or_below_limit(t *testing.T) {
	for _, testCase := range []struct {
		name          string
		body          string
		contentLength int64
	}{
		{name: "empty", body: "", contentLength: 0},
		{name: "known exact limit", body: "1234", contentLength: 4},
		{name: "streamed exact limit", body: "1234", contentLength: -1},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			// Given
			installGlobalTestLogger(t, &bytes.Buffer{})
			router := gin.New()
			router.Use(middleware.BodySizeLimit(config.ByteSize(4)))
			router.POST("/probe", func(context *gin.Context) {
				body, err := io.ReadAll(context.Request.Body)
				require.NoError(t, err)
				require.Equal(t, testCase.body, string(body))
				context.Status(http.StatusNoContent)
			})
			request := httptest.NewRequest(http.MethodPost, "/probe", strings.NewReader(testCase.body))
			request.ContentLength = testCase.contentLength
			recorder := httptest.NewRecorder()

			// When
			router.ServeHTTP(recorder, request)

			// Then
			require.Equal(t, http.StatusNoContent, recorder.Code)
		})
	}
}

func Test_BodySizeLimit_rejects_real_chunked_HTTP_request(t *testing.T) {
	// Given
	installGlobalTestLogger(t, &bytes.Buffer{})
	transferEncoding := make(chan []string, 1)
	router := gin.New()
	router.Use(middleware.BodySizeLimit(config.ByteSize(4)))
	router.POST("/probe", func(context *gin.Context) {
		transferEncoding <- append([]string(nil), context.Request.TransferEncoding...)
		_, err := io.ReadAll(context.Request.Body)
		if err != nil {
			require.NotNil(t, context.Error(err))
		}
	})
	server := httptest.NewServer(router)
	t.Cleanup(server.Close)
	body := struct{ io.Reader }{Reader: strings.NewReader("12345")}
	request, err := http.NewRequest(http.MethodPost, server.URL+"/probe", body)
	require.NoError(t, err)

	// When
	result, err := server.Client().Do(request)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, result.Body.Close()) })
	responseBody, err := io.ReadAll(result.Body)
	require.NoError(t, err)

	// Then
	require.Equal(t, http.StatusOK, result.StatusCode)
	require.Equal(t, []string{"chunked"}, <-transferEncoding)
	requireSemanticEnvelope(t, responseBody, http.StatusBadRequest, "INVALID_ARGUMENT")
}

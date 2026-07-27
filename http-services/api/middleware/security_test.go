package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"http-services/api/middleware"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func Test_SecurityHeaders_sets_deterministic_headers(t *testing.T) {
	// Given
	router := gin.New()
	router.Use(middleware.SecurityHeaders())
	router.GET("/probe", func(context *gin.Context) { context.Status(http.StatusNoContent) })
	recorder := httptest.NewRecorder()

	// When
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/probe", nil))

	// Then
	require.Equal(t, "nosniff", recorder.Header().Get("X-Content-Type-Options"))
	require.Equal(t, "DENY", recorder.Header().Get("X-Frame-Options"))
	require.Equal(t, "1; mode=block", recorder.Header().Get("X-XSS-Protection"))
}

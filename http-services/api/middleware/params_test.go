package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	httpLog "http-services/utils/log"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func Test_CheckJSONParam_binds_and_stores_params_for_logging(t *testing.T) {
	// Given
	context, _ := gin.CreateTestContext(httptest.NewRecorder())
	context.Request = httptest.NewRequest(http.MethodPost, "/ticket", strings.NewReader(`{"token":"jwt"}`))
	params := &struct {
		Token string `json:"token" binding:"required"`
	}{}

	// When
	ok := CheckJSONParam(params, context)

	// Then
	require.True(t, ok)
	require.Equal(t, "jwt", params.Token)
	bound, exists := context.Get(httpLog.BoundParamsKey)
	require.True(t, exists)
	require.Same(t, params, bound)
}

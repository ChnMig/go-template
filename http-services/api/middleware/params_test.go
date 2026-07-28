package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	serviceLog "http-services/utils/log"

	"github.com/gin-gonic/gin"
)

func TestCheckJSONParamStoresCompleteParamsForLogging(t *testing.T) {
	context, _ := gin.CreateTestContext(httptest.NewRecorder())
	context.Request = httptest.NewRequest(http.MethodPost, "/ticket", strings.NewReader(`{"token":"jwt-full-value"}`))
	params := &struct {
		Token string `json:"token" binding:"required"`
	}{}

	if !CheckJSONParam(params, context) {
		t.Fatal("CheckJSONParam() = false, want true")
	}
	if params.Token != "jwt-full-value" {
		t.Fatalf("Token = %q, want complete value", params.Token)
	}
	bound, exists := context.Get(serviceLog.BoundParamsKey)
	if !exists || bound != params {
		t.Fatalf("bound params = %#v, want same params pointer", bound)
	}
}

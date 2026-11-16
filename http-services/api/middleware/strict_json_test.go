package middleware

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"http-services/api/response"
	"http-services/utils/contextkey"
	serviceLog "http-services/utils/log"

	"github.com/gin-gonic/gin"
)

func TestStrictJSONBindingHTTPContract(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	for path, check := range map[string]func(any, *gin.Context) bool{
		"/strict": CheckStrictJSONParam,
		"/normal": CheckJSONParam,
	} {
		router.POST(path, func(c *gin.Context) {
			params := &struct {
				Name string `json:"name" binding:"required"`
			}{}
			if check(params, c) {
				response.ReturnOk(c, params)
			}
		})
	}
	for _, tc := range []struct {
		name, path, body string
		code             int
	}{
		{"valid", "/strict", `{"name":"alice"} `, response.OK.Code},
		{"unknown field", "/strict", `{"name":"alice","extra":1}`, response.INVALID_ARGUMENT.Code},
		{"second value", "/strict", `{"name":"alice"} {"name":"bob"}`, response.INVALID_ARGUMENT.Code},
		{"invalid suffix", "/strict", `{"name":"alice"} broken`, response.INVALID_ARGUMENT.Code},
		{"syntax error", "/strict", `{"name":`, response.INVALID_ARGUMENT.Code},
		{"required field", "/strict", `{}`, response.INVALID_ARGUMENT.Code},
		{"empty body", "/strict", "", response.INVALID_ARGUMENT.Code},
		{"normal binding still allows unknown fields", "/normal", `{"name":"alice","extra":1}`, response.OK.Code},
	} {
		t.Run(tc.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, tc.path, strings.NewReader(tc.body))
			request.Header.Set("Content-Type", "application/json")
			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, request)
			var result struct {
				Code   int               `json:"code"`
				Detail map[string]string `json:"detail"`
			}
			if err := json.Unmarshal(recorder.Body.Bytes(), &result); err != nil {
				t.Fatal(err)
			}
			if recorder.Code != http.StatusOK || result.Code != tc.code {
				t.Fatalf("status=%d body=%s, want code=%v", recorder.Code, recorder.Body.String(), tc.code)
			}
			if tc.code == response.OK.Code && result.Detail["name"] != "alice" {
				t.Fatalf("bound fields missing: %s", recorder.Body.String())
			}
		})
	}
}

func TestStrictJSONFailureKeepsRawInputAndSpecificError(t *testing.T) {
	logger, output := requestParamsTestLogger()
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Set(contextkey.Logger, logger)
	c.Set(serviceLog.BoundParamsKey, map[string]any{"stale": "old request"})
	c.Request = httptest.NewRequest(http.MethodPost, "/strict", strings.NewReader(`{"name":"alice","extra":"original value"}`))
	c.Request.Header.Set("Content-Type", "application/json")
	params := &struct {
		Name string `json:"name"`
	}{}
	if CheckStrictJSONParamWithMessage(params, c, "请求字段不正确，请检查提交内容") {
		t.Fatal("unknown field accepted")
	}
	if bound, _ := c.Get(serviceLog.BoundParamsKey); bound != nil {
		t.Fatalf("failed bind retained old parameters: %v", bound)
	}
	entries := filterRequestLogs(t, output, "msg", "invalid request parameters")
	if len(entries) != 1 || !strings.Contains(entries[0]["error"].(string), `unknown field "extra"`) {
		t.Fatalf("specific parsing error missing: %s", output.String())
	}
	body, ok := entries[0]["request_body"].(map[string]any)
	if !ok || body["extra"] != "original value" {
		t.Fatalf("original input missing: %s", output.String())
	}
	var result struct {
		Message string `json:"message"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if result.Message != "请求字段不正确，请检查提交内容" {
		t.Fatalf("custom message lost: %s", recorder.Body.String())
	}
}

func TestBindStrictJSONPreservesBodyLimitError(t *testing.T) {
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodPost, "/strict", strings.NewReader(`{"name":"alice"} `+strings.Repeat(" ", 32)))
	c.Request.Body = http.MaxBytesReader(recorder, c.Request.Body, 20)
	params := &struct {
		Name string `json:"name"`
	}{}
	err := BindStrictJSONParam(params, c)
	var limitError *http.MaxBytesError
	if !errors.As(err, &limitError) {
		t.Fatalf("body limit error lost: %v", err)
	}
	if c.IsAborted() || recorder.Body.Len() != 0 {
		t.Fatalf("BindStrictJSONParam wrote a response: %s", recorder.Body.String())
	}
}

package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"http-services/api/response"
	"http-services/config"
	"http-services/utils/contextkey"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func TestTokenVerifyRejectsInvalidAlgorithmAndMissingExpiration(t *testing.T) {
	gin.SetMode(gin.TestMode)
	previousKey := config.JWTKey
	config.JWTKey = "middleware-validation-key-at-least-32-characters"
	t.Cleanup(func() { config.JWTKey = previousKey })
	for _, tc := range []struct {
		name   string
		method jwt.SigningMethod
		exp    bool
		allow  bool
	}{
		{"HS256 with expiration", jwt.SigningMethodHS256, true, true},
		{"HS384", jwt.SigningMethodHS384, true, false},
		{"HS512", jwt.SigningMethodHS512, true, false},
		{"missing expiration", jwt.SigningMethodHS256, false, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			claims := jwt.MapClaims{"data": map[string]any{"user_id": "u1"}}
			if tc.exp {
				claims["exp"] = time.Now().Add(time.Hour).Unix()
			}
			raw, err := jwt.NewWithClaims(tc.method, claims).SignedString([]byte(config.JWTKey))
			if err != nil {
				t.Fatal(err)
			}
			called := false
			router := gin.New()
			router.Use(TokenVerify)
			router.GET("/protected", func(c *gin.Context) {
				called = true
				data, exists := c.Get(contextkey.JWTData)
				if !exists || data.(map[string]any)["user_id"] != "u1" {
					t.Fatalf("identity context changed: %v", data)
				}
				response.ReturnOk(c, nil)
			})
			request := httptest.NewRequest(http.MethodGet, "/protected", nil)
			request.Header.Set(AuthorizationHeader, raw)
			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, request)
			var result struct {
				Code int `json:"code"`
			}
			if err := json.Unmarshal(recorder.Body.Bytes(), &result); err != nil {
				t.Fatal(err)
			}
			want := response.UNAUTHENTICATED.Code
			if tc.allow {
				want = response.OK.Code
			}
			if recorder.Code != http.StatusOK || called != tc.allow || result.Code != want {
				t.Fatalf("handler called=%v status=%d body=%s", called, recorder.Code, recorder.Body.String())
			}
		})
	}
}

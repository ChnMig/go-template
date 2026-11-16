package authentication

import (
	"errors"
	"testing"
	"time"

	"http-services/config"

	"github.com/golang-jwt/jwt/v5"
)

func TestParseHS256Validation(t *testing.T) {
	previousKey := config.JWTKey
	config.JWTKey = "validation-test-key-at-least-32-characters"
	t.Cleanup(func() { config.JWTKey = previousKey })

	now := time.Now()
	for _, tc := range []struct {
		name   string
		method jwt.SigningMethod
		claims jwt.MapClaims
		key    string
		want   error
	}{
		{"valid", jwt.SigningMethodHS256, jwt.MapClaims{"exp": now.Add(time.Hour).Unix()}, config.JWTKey, nil},
		{"HS384", jwt.SigningMethodHS384, jwt.MapClaims{"exp": now.Add(time.Hour).Unix()}, config.JWTKey, jwt.ErrTokenSignatureInvalid},
		{"HS512", jwt.SigningMethodHS512, jwt.MapClaims{"exp": now.Add(time.Hour).Unix()}, config.JWTKey, jwt.ErrTokenSignatureInvalid},
		{"missing expiration", jwt.SigningMethodHS256, jwt.MapClaims{}, config.JWTKey, jwt.ErrTokenRequiredClaimMissing},
		{"expired", jwt.SigningMethodHS256, jwt.MapClaims{"exp": now.Add(-time.Hour).Unix()}, config.JWTKey, jwt.ErrTokenExpired},
		{"not active", jwt.SigningMethodHS256, jwt.MapClaims{"exp": now.Add(2 * time.Hour).Unix(), "nbf": now.Add(time.Hour).Unix()}, config.JWTKey, jwt.ErrTokenNotValidYet},
		{"wrong signature", jwt.SigningMethodHS256, jwt.MapClaims{"exp": now.Add(time.Hour).Unix()}, "different-key", jwt.ErrTokenSignatureInvalid},
	} {
		t.Run(tc.name, func(t *testing.T) {
			tc.claims["data"] = map[string]any{"user_id": "u1"}
			raw, err := jwt.NewWithClaims(tc.method, tc.claims).SignedString([]byte(tc.key))
			if err != nil {
				t.Fatal(err)
			}
			claims := &MapClaims{}
			token, err := ParseHS256(raw, claims)
			if tc.want != nil {
				if !errors.Is(err, tc.want) {
					t.Fatalf("ParseHS256 error = %v, want %v", err, tc.want)
				}
				if _, err := JWTDecrypt(raw); !errors.Is(err, tc.want) {
					t.Fatalf("JWTDecrypt error = %v, want %v", err, tc.want)
				}
				return
			}
			if err != nil || token == nil || !token.Valid || claims.Data["user_id"] != "u1" {
				t.Fatalf("valid token rejected or claims changed: token=%v claims=%+v error=%v", token, claims, err)
			}
		})
	}
}

func TestParseHS256RejectsUnconfiguredKey(t *testing.T) {
	previousKey := config.JWTKey
	t.Cleanup(func() { config.JWTKey = previousKey })
	for _, key := range []string{"", " \t\n"} {
		config.JWTKey = key
		raw, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"exp": time.Now().Add(time.Hour).Unix(),
		}).SignedString([]byte(key))
		if err != nil {
			t.Fatal(err)
		}
		if token, err := ParseHS256(raw, &MapClaims{}); err == nil || token != nil {
			t.Fatalf("unconfigured key accepted: token=%v error=%v", token, err)
		}
	}
}

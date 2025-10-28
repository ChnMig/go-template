package authentication

import (
	"testing"
	"time"

	"http-services/config"

	"github.com/golang-jwt/jwt/v5"
)

func init() {
	// 设置测试用的配置
	config.JWTKey = "test-secret-key-at-least-32-chars-long"
	config.JWTExpiration = 1 * time.Hour
}

func TestJWTIssueAndDecrypt(t *testing.T) {
	testData := map[string]interface{}{
		"user_id":  "123",
		"username": "test_user",
		"role":     "admin",
	}

	// 测试签发 token
	token, err := JWTIssue(testData)
	if err != nil {
		t.Fatalf("JWTIssue failed: %v", err)
	}

	if token == "" {
		t.Fatal("JWTIssue returned empty token")
	}

	// 测试解密 token
	data, err := JWTDecrypt(token)
	if err != nil {
		t.Fatalf("JWTDecrypt failed: %v", err)
	}

	// 验证数据
	if data["user_id"] != testData["user_id"] {
		t.Errorf("user_id mismatch: got %v, want %v", data["user_id"], testData["user_id"])
	}
	if data["username"] != testData["username"] {
		t.Errorf("username mismatch: got %v, want %v", data["username"], testData["username"])
	}
	if data["role"] != testData["role"] {
		t.Errorf("role mismatch: got %v, want %v", data["role"], testData["role"])
	}
}

func TestJWTDecryptInvalidToken(t *testing.T) {
	// 测试无效的 token
	_, err := JWTDecrypt("invalid.token.here")
	if err == nil {
		t.Error("JWTDecrypt should fail on invalid token")
	}
}

func TestJWTDecryptExpiredToken(t *testing.T) {
	// 创建一个已经过期的自定义 claims
	claims := MapClaims{
		Data: map[string]interface{}{"user_id": "123"},
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)), // 1小时前过期
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)), // 2小时前签发
			NotBefore: jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
			Issuer:    defaultIssuer,
			Subject:   defaultSubject,
			Audience:  jwt.ClaimStrings{defaultAudience},
		},
	}

	// 签发已过期的 token
	token, err := SignHS256(&claims)
	if err != nil {
		t.Fatalf("SignHS256 failed: %v", err)
	}

	// 尝试解密已过期的 token
	_, err = JWTDecrypt(token)
	if err == nil {
		t.Error("JWTDecrypt should fail on expired token")
	}
}

func TestPrepareRegisteredClaims(t *testing.T) {
	claims := &jwt.RegisteredClaims{}
	PrepareRegisteredClaims(claims)

	// 检查是否填充了默认值
	if claims.Issuer != defaultIssuer {
		t.Errorf("Issuer not set correctly: got %s, want %s", claims.Issuer, defaultIssuer)
	}

	if claims.Subject != defaultSubject {
		t.Errorf("Subject not set correctly: got %s, want %s", claims.Subject, defaultSubject)
	}

	if len(claims.Audience) == 0 || claims.Audience[0] != defaultAudience {
		t.Errorf("Audience not set correctly: got %v, want [%s]", claims.Audience, defaultAudience)
	}

	if claims.ID == "" {
		t.Error("ID not set")
	}

	if claims.IssuedAt == nil {
		t.Error("IssuedAt not set")
	}

	if claims.NotBefore == nil {
		t.Error("NotBefore not set")
	}

	if claims.ExpiresAt == nil {
		t.Error("ExpiresAt not set")
	}
}

func TestSignAndParseHS256(t *testing.T) {
	claims := &MapClaims{
		Data: map[string]interface{}{"user_id": "test-123"},
	}
	PrepareRegisteredClaims(&claims.RegisteredClaims)

	// 测试签名
	tokenString, err := SignHS256(claims)
	if err != nil {
		t.Fatalf("SignHS256 failed: %v", err)
	}

	// 测试解析
	parsedClaims := &MapClaims{}
	token, err := ParseHS256(tokenString, parsedClaims)
	if err != nil {
		t.Fatalf("ParseHS256 failed: %v", err)
	}

	if !token.Valid {
		t.Error("Token should be valid")
	}

	if parsedClaims.Data["user_id"] != claims.Data["user_id"] {
		t.Errorf("Data mismatch: got %v, want %v", parsedClaims.Data["user_id"], claims.Data["user_id"])
	}
}

package random

import (
	"encoding/base64"
	"testing"
)

func TestHex_Length(t *testing.T) {
	s, err := Hex(16)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if len(s) != 32 {
		t.Fatalf("expect len=32, got %d", len(s))
	}
}

func TestHex_Invalid(t *testing.T) {
	if _, err := Hex(0); err == nil {
		t.Fatalf("expect err")
	}
}

func TestBase64URL(t *testing.T) {
	s, err := Base64URL(16)
	if err != nil {
		t.Fatalf("Base64URL() error = %v", err)
	}
	if len(s) != 22 {
		t.Fatalf("Base64URL(16) length = %d, want 22", len(s))
	}
	decoded, err := base64.RawURLEncoding.DecodeString(s)
	if err != nil {
		t.Fatalf("Base64URL(16) returned invalid Base64URL: %v", err)
	}
	if len(decoded) != 16 {
		t.Fatalf("decoded length = %d, want 16", len(decoded))
	}
	if _, err := Base64URL(0); err == nil {
		t.Fatal("Base64URL(0) expected error")
	}
}

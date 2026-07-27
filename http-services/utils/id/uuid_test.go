package id

import (
	"regexp"
	"testing"

	"github.com/google/uuid"
)

func Test_GenerateUUIDv7_returns_RFC9562_version_7(t *testing.T) {
	// Given / When
	generated, err := GenerateUUIDv7()
	// Then
	if err != nil {
		t.Fatalf("GenerateUUIDv7() error = %v", err)
	}
	if generated.Version() != uuid.Version(7) {
		t.Fatalf("GenerateUUIDv7() version = %v, want 7", generated.Version())
	}
	if generated.Variant() != uuid.RFC4122 {
		t.Fatalf("GenerateUUIDv7() variant = %v, want RFC4122", generated.Variant())
	}
}

func Test_GenerateUUIDv7MD5_returns_32_lowercase_hex_characters(t *testing.T) {
	// Given / When
	generated, err := GenerateUUIDv7MD5()
	// Then
	if err != nil {
		t.Fatalf("GenerateUUIDv7MD5() error = %v", err)
	}
	if !regexp.MustCompile(`^[a-f0-9]{32}$`).MatchString(generated) {
		t.Fatalf("GenerateUUIDv7MD5() = %q, want 32 lowercase hexadecimal characters", generated)
	}
}

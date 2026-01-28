package random

import "testing"

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

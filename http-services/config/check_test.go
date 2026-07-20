package config

import (
	"strings"
	"testing"
)

func TestValidateJWTConfig(t *testing.T) {
	validKey := strings.Repeat("a", minJWTKeyLength)
	tests := []struct {
		name       string
		key        string
		expiration int64
		wantErr    bool
	}{
		{name: "rejects empty key", expiration: 1, wantErr: true},
		{name: "rejects yaml example placeholder", key: "YOUR_SECRET_KEY_HERE", expiration: 1, wantErr: true},
		{name: "rejects env example placeholder", key: "YOUR_SECRET_KEY_HERE_AT_LEAST_32_CHARACTERS", expiration: 1, wantErr: true},
		{name: "rejects short key", key: strings.Repeat("a", minJWTKeyLength-1), expiration: 1, wantErr: true},
		{name: "rejects zero expiration", key: validKey, wantErr: true},
		{name: "rejects negative expiration", key: validKey, expiration: -1, wantErr: true},
		{name: "accepts valid config", key: validKey, expiration: 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateConfig(tt.key, tt.expiration)
			if tt.wantErr && err == nil {
				t.Fatal("expected validation error")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("expected valid config, got %v", err)
			}
		})
	}
}

func TestUnsafeDefaultKeysContainsDocumentedPlaceholders(t *testing.T) {
	for _, key := range []string{
		"YOUR_SECRET_KEY_HERE",
		"YOUR_SECRET_KEY_HERE_AT_LEAST_32_CHARACTERS",
	} {
		t.Run(key, func(t *testing.T) {
			if _, ok := unsafeDefaultKeys[key]; !ok {
				t.Fatalf("documented placeholder %q is not rejected", key)
			}
		})
	}
}

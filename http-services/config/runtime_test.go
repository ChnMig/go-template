package config

import (
	"strings"
	"testing"
	"time"
)

func TestLoadReturnsValidatedStartupSnapshot(t *testing.T) {
	t.Chdir(t.TempDir())
	t.Setenv("HTTP_SERVICES_JWT_KEY", strings.Repeat("a", minJWTKeyLength))

	snapshot, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if snapshot.Server.Address() != "0.0.0.0:8080" {
		t.Fatalf("address = %q", snapshot.Server.Address())
	}
	if snapshot.JWT.Expiration != 12*time.Hour {
		t.Fatalf("JWT expiration = %v", snapshot.JWT.Expiration)
	}
	if snapshot.Redis.Host != "" {
		t.Fatalf("Redis host = %q, want disabled", snapshot.Redis.Host)
	}
}

package pidfile

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestWriteCreatesExclusivePrivatePIDFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "run", "service.pid")
	if err := Write(path, 1234); err != nil {
		t.Fatalf("Write() error = %v", err)
	}
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if string(content) != "1234\n" {
		t.Fatalf("content = %q, want %q", content, "1234\\n")
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("Stat() error = %v", err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("mode = %o, want 600", info.Mode().Perm())
	}
	if err := Write(path, 5678); !errors.Is(err, ErrExists) {
		t.Fatalf("second Write() error = %v, want ErrExists", err)
	}
}

func TestRemoveVerifiesPIDOwnershipAndIsIdempotent(t *testing.T) {
	path := filepath.Join(t.TempDir(), "service.pid")
	if err := Write(path, 1234); err != nil {
		t.Fatalf("Write() error = %v", err)
	}
	if err := Remove(path, 5678); !errors.Is(err, ErrOwnership) {
		t.Fatalf("Remove() error = %v, want ErrOwnership", err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("mismatched owner removed file: %v", err)
	}
	if err := Remove(path, 1234); err != nil {
		t.Fatalf("Remove() error = %v", err)
	}
	if err := Remove(path, 1234); err != nil {
		t.Fatalf("second Remove() error = %v", err)
	}
}

package msqldb

import (
	"errors"
	"testing"

	"http-services/config"
)

func TestInitRequiresMysqlDSN(t *testing.T) {
	oldDSN := config.MysqlDSN
	config.MysqlDSN = ""
	t.Cleanup(func() { config.MysqlDSN = oldDSN })

	if err := Init(); !errors.Is(err, ErrMissingMysqlDSN) {
		t.Fatalf("Init() error = %v, want %v", err, ErrMissingMysqlDSN)
	}
}

func TestClientRequiresMysqlDSN(t *testing.T) {
	oldDSN := config.MysqlDSN
	config.MysqlDSN = ""
	t.Cleanup(func() { config.MysqlDSN = oldDSN })

	client, err := Client()
	if client != nil {
		t.Fatalf("Client() client = %v, want nil", client)
	}
	if !errors.Is(err, ErrMissingMysqlDSN) {
		t.Fatalf("Client() error = %v, want %v", err, ErrMissingMysqlDSN)
	}
}

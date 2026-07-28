package rdb

import (
	"errors"
	"testing"

	"http-services/config"
)

func TestInitRequiresRedisHost(t *testing.T) {
	oldHost := config.RedisHost
	config.RedisHost = ""
	t.Cleanup(func() { config.RedisHost = oldHost })

	if err := Init(); !errors.Is(err, ErrMissingRedisHost) {
		t.Fatalf("Init() error = %v, want %v", err, ErrMissingRedisHost)
	}
}

func TestClientRequiresRedisHost(t *testing.T) {
	oldHost := config.RedisHost
	config.RedisHost = ""
	t.Cleanup(func() { config.RedisHost = oldHost })

	client, err := Client()
	if client != nil {
		t.Fatalf("Client() client = %v, want nil", client)
	}
	if !errors.Is(err, ErrMissingRedisHost) {
		t.Fatalf("Client() error = %v, want %v", err, ErrMissingRedisHost)
	}
}

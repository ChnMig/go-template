package rdb

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"http-services/config"

	"github.com/stretchr/testify/require"
)

func TestNewReturnsMissingHostError(t *testing.T) {
	client, err := New(t.Context(), config.RedisConfig{})
	require.Nil(t, client)
	require.ErrorIs(t, err, ErrMissingHost)
}

func TestNewClientUsesChildTimeoutAndClosesOnFailure(t *testing.T) {
	pingErr := errors.New("ping failure")
	var deadline time.Time
	database := &recordingRedisDatabase{ping: func(ctx context.Context) error {
		var ok bool
		deadline, ok = ctx.Deadline()
		require.True(t, ok)
		return pingErr
	}}
	started := time.Now()

	client, err := newClient(t.Context(), validRedisConfig(), func(clientOptions) redisDatabase { return database })

	require.Nil(t, client)
	require.ErrorIs(t, err, ErrPing)
	require.ErrorIs(t, err, pingErr)
	require.WithinDuration(t, started.Add(connectionTimeout), deadline, 50*time.Millisecond)
	require.EqualValues(t, 1, database.closeCalls.Load())
}

func TestOptionsFromConfigMapsSettings(t *testing.T) {
	cfg := validRedisConfig()
	options := optionsFromConfig(cfg)
	require.Equal(t, cfg.Host, options.address)
	require.Equal(t, cfg.Password, options.password)
	require.Equal(t, cfg.KeyPrefix, options.keyPrefix)
	require.Equal(t, defaultPoolSize, options.poolSize)
	require.Equal(t, defaultMinIdleConns, options.minIdleConns)
}

func TestClientCloseIsExactlyOnceAndPreservesError(t *testing.T) {
	closeErr := errors.New("close failure")
	database := &recordingRedisDatabase{close: func() error { return closeErr }}
	client := &Client{database: database}
	const callers = 32
	errorsByCaller := make(chan error, callers)
	var waitGroup sync.WaitGroup
	waitGroup.Add(callers)
	for range callers {
		go func() {
			defer waitGroup.Done()
			errorsByCaller <- client.Close()
		}()
	}
	waitGroup.Wait()
	close(errorsByCaller)

	for err := range errorsByCaller {
		require.ErrorIs(t, err, ErrClose)
		require.ErrorIs(t, err, closeErr)
	}
	require.EqualValues(t, 1, database.closeCalls.Load())
}

func validRedisConfig() config.RedisConfig {
	return config.RedisConfig{Host: "127.0.0.1:6379", Password: "secret", KeyPrefix: "test:"}
}

type recordingRedisDatabase struct {
	ping       func(context.Context) error
	close      func() error
	closeCalls atomic.Int32
}

func (d *recordingRedisDatabase) Ping(ctx context.Context) error {
	if d.ping != nil {
		return d.ping(ctx)
	}
	return nil
}

func (d *recordingRedisDatabase) Close() error {
	d.closeCalls.Add(1)
	if d.close != nil {
		return d.close()
	}
	return nil
}

package msqldb

import (
	"context"
	"database/sql"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestNewReturnsMissingDSNError(t *testing.T) {
	client, err := New(t.Context(), "")
	require.Nil(t, client)
	require.ErrorIs(t, err, ErrMissingDSN)
}

func TestNewReturnsOpenErrorForInvalidDSN(t *testing.T) {
	client, err := New(t.Context(), "%%%")
	require.Nil(t, client)
	require.ErrorIs(t, err, ErrOpen)
}

func TestVerifyConnectionUsesChildTimeoutAndClosesOnFailure(t *testing.T) {
	pingErr := errors.New("ping failure")
	var deadline time.Time
	var closes atomic.Int32
	client := &Client{sqlDB: &recordingDatabase{
		ping: func(ctx context.Context) error {
			var ok bool
			deadline, ok = ctx.Deadline()
			require.True(t, ok)
			return pingErr
		},
		close: func() error { closes.Add(1); return nil },
	}}
	started := time.Now()

	result, err := verifyConnection(t.Context(), 150*time.Millisecond, client)

	require.Nil(t, result)
	require.ErrorIs(t, err, ErrPing)
	require.ErrorIs(t, err, pingErr)
	require.WithinDuration(t, started.Add(150*time.Millisecond), deadline, 50*time.Millisecond)
	require.EqualValues(t, 1, closes.Load())
}

func TestConfigurePoolAppliesSettings(t *testing.T) {
	pool := &recordingPool{}
	configurePool(pool)
	require.Equal(t, maxIdleConnections, pool.maxIdleConns)
	require.Equal(t, maxOpenConnections, pool.maxOpenConns)
	require.Equal(t, connectionLifetime, pool.connMaxLifetime)
	require.Equal(t, connectionIdleTime, pool.connMaxIdleTime)
}

func TestClientCloseIsExactlyOnceWhenConcurrent(t *testing.T) {
	var preparedCloses atomic.Int32
	var sqlCloses atomic.Int32
	client := &Client{
		prepared: preparedCloserFunc(func() { preparedCloses.Add(1) }),
		sqlDB:    &recordingDatabase{close: func() error { sqlCloses.Add(1); return nil }},
	}
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
		require.NoError(t, err)
	}
	require.EqualValues(t, 1, preparedCloses.Load())
	require.EqualValues(t, 1, sqlCloses.Load())
}

type recordingPool struct {
	connMaxIdleTime time.Duration
	connMaxLifetime time.Duration
	maxIdleConns    int
	maxOpenConns    int
}

func (p *recordingPool) SetMaxIdleConns(value int)              { p.maxIdleConns = value }
func (p *recordingPool) SetMaxOpenConns(value int)              { p.maxOpenConns = value }
func (p *recordingPool) SetConnMaxLifetime(value time.Duration) { p.connMaxLifetime = value }
func (p *recordingPool) SetConnMaxIdleTime(value time.Duration) { p.connMaxIdleTime = value }

type recordingDatabase struct {
	ping  func(context.Context) error
	close func() error
}

func (d *recordingDatabase) PingContext(ctx context.Context) error {
	if d.ping != nil {
		return d.ping(ctx)
	}
	return nil
}

func (d *recordingDatabase) Close() error {
	if d.close != nil {
		return d.close()
	}
	return nil
}
func (*recordingDatabase) Stats() sql.DBStats { return sql.DBStats{} }

type preparedCloserFunc func()

func (close preparedCloserFunc) Close() { close() }

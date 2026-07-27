package msqldb

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	stdlog "log"
	"strings"
	"sync"
	"time"

	httplog "http-services/utils/log"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

var (
	ErrMissingDSN = errors.New("database.mysql_dsn is empty")
	ErrOpen       = errors.New("open mysql")
	ErrPing       = errors.New("ping mysql")
)

const (
	maxIdleConnections = 25
	maxOpenConnections = 100
	connectionTimeout  = 5 * time.Second
	connectionLifetime = time.Hour
	connectionIdleTime = 30 * time.Minute
)

type Client struct {
	database *gorm.DB
	prepared preparedStatementCloser
	sqlDB    sqlDatabase
	closeErr error
	closeMu  sync.Mutex
	closed   bool
}

func New(ctx context.Context, dsn string) (*Client, error) {
	if strings.TrimSpace(dsn) == "" {
		return nil, fmt.Errorf("new mysql client: %w", ErrMissingDSN)
	}
	database, err := gorm.Open(mysql.New(mysql.Config{DSN: dsn, SkipInitializeWithVersion: true}), &gorm.Config{
		Logger: newGORMLogger(), PrepareStmt: true, DisableAutomaticPing: true,
		DisableForeignKeyConstraintWhenMigrating: true, DisableNestedTransaction: true,
		CreateBatchSize: 1000, QueryFields: true,
	})
	if err != nil {
		return nil, errors.Join(fmt.Errorf("%w: initialize driver: %w", ErrOpen, err), closeOpenedDatabase(database))
	}
	sqlDB, err := database.DB()
	if err != nil {
		prepared := preparedStatementManager(database)
		if prepared != nil {
			prepared.Close()
		}
		return nil, fmt.Errorf("%w: acquire SQL pool: %w", ErrOpen, err)
	}
	configurePool(sqlDB)
	client := &Client{database: database, prepared: preparedStatementManager(database), sqlDB: sqlDB}
	return verifyConnection(ctx, connectionTimeout, client)
}

func (c *Client) Database() *gorm.DB { return c.database }

func verifyConnection(ctx context.Context, timeout time.Duration, client *Client) (*Client, error) {
	pingCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	if err := client.Ping(pingCtx); err != nil {
		return nil, errors.Join(err, client.Close())
	}
	return client, nil
}

func (c *Client) Ping(ctx context.Context) error {
	if err := c.sqlDB.PingContext(ctx); err != nil {
		return fmt.Errorf("%w: check connection: %w", ErrPing, err)
	}
	return nil
}

func (c *Client) Stats() sql.DBStats { return c.sqlDB.Stats() }

func (c *Client) Close() error {
	c.closeMu.Lock()
	defer c.closeMu.Unlock()
	if c.closed {
		return c.closeErr
	}
	c.closed = true
	if c.prepared != nil {
		c.prepared.Close()
	}
	if err := c.sqlDB.Close(); err != nil {
		c.closeErr = fmt.Errorf("close mysql SQL pool: %w", err)
	}
	return c.closeErr
}

func newGORMLogger() gormlogger.Interface {
	writer := httplog.NewZapWriter(zap.L().With(zap.String("component", "gorm")), zapcore.InfoLevel)
	return gormlogger.New(stdlog.New(writer, "\r\n", stdlog.LstdFlags), gormlogger.Config{
		SlowThreshold: 200 * time.Millisecond, LogLevel: gormlogger.Warn,
		IgnoreRecordNotFoundError: true, Colorful: false, ParameterizedQueries: true,
	})
}

func preparedStatementManager(database *gorm.DB) preparedStatementCloser {
	prepared, ok := database.ConnPool.(*gorm.PreparedStmtDB)
	if !ok {
		return nil
	}
	return prepared
}

func closeOpenedDatabase(database *gorm.DB) error {
	if database == nil || database.Config == nil {
		return nil
	}
	if prepared := preparedStatementManager(database); prepared != nil {
		prepared.Close()
	}
	sqlDB, err := database.DB()
	if err != nil {
		return fmt.Errorf("acquire mysql SQL pool for cleanup: %w", err)
	}
	if err := sqlDB.Close(); err != nil {
		return fmt.Errorf("close mysql SQL pool after open failure: %w", err)
	}
	return nil
}

func configurePool(pool poolConfigurer) {
	pool.SetMaxIdleConns(maxIdleConnections)
	pool.SetMaxOpenConns(maxOpenConnections)
	pool.SetConnMaxLifetime(connectionLifetime)
	pool.SetConnMaxIdleTime(connectionIdleTime)
}

type (
	preparedStatementCloser interface{ Close() }
	sqlDatabase             interface {
		PingContext(context.Context) error
		Close() error
		Stats() sql.DBStats
	}
)

type poolConfigurer interface {
	SetMaxIdleConns(int)
	SetMaxOpenConns(int)
	SetConnMaxLifetime(time.Duration)
	SetConnMaxIdleTime(time.Duration)
}

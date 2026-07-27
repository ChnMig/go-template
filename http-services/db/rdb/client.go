package rdb

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"http-services/config"

	"github.com/redis/go-redis/v9"
)

var (
	ErrMissingHost = errors.New("redis.host is empty")
	ErrPing        = errors.New("ping redis")
	ErrClose       = errors.New("close redis")
)

const (
	defaultDatabase     = 0
	defaultPoolSize     = 100
	defaultMinIdleConns = 10
	defaultDialTimeout  = 5 * time.Second
	defaultReadTimeout  = 3 * time.Second
	defaultWriteTimeout = 3 * time.Second
	connectionTimeout   = 5 * time.Second
)

type Settings struct {
	KeyPrefix    string
	DialTimeout  time.Duration
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	DB           int
	PoolSize     int
	MinIdleConns int
}

type Client struct {
	database redisDatabase
	closeErr error
	settings Settings
	closeMu  sync.Mutex
	closed   bool
}

func New(ctx context.Context, cfg config.RedisConfig) (*Client, error) {
	return newClient(ctx, cfg, newRedisDatabase)
}

func newClient(ctx context.Context, cfg config.RedisConfig, open func(clientOptions) redisDatabase) (*Client, error) {
	if strings.TrimSpace(cfg.Host) == "" {
		return nil, fmt.Errorf("new redis client: %w", ErrMissingHost)
	}
	options := optionsFromConfig(cfg)
	client := &Client{database: open(options), settings: options.settings()}
	pingCtx, cancel := context.WithTimeout(ctx, connectionTimeout)
	defer cancel()
	if err := client.Ping(pingCtx); err != nil {
		return nil, errors.Join(err, client.Close())
	}
	return client, nil
}

func (c *Client) Ping(ctx context.Context) error {
	if err := c.database.Ping(ctx); err != nil {
		return fmt.Errorf("%w: check connection: %w", ErrPing, err)
	}
	return nil
}

func (c *Client) Settings() Settings { return c.settings }

func (c *Client) Close() error {
	c.closeMu.Lock()
	defer c.closeMu.Unlock()
	if c.closed {
		return c.closeErr
	}
	c.closed = true
	if err := c.database.Close(); err != nil {
		c.closeErr = fmt.Errorf("%w: release client: %w", ErrClose, err)
	}
	return c.closeErr
}

type clientOptions struct {
	address, password, keyPrefix           string
	dialTimeout, readTimeout, writeTimeout time.Duration
	database, poolSize, minIdleConns       int
}

func optionsFromConfig(cfg config.RedisConfig) clientOptions {
	return clientOptions{
		address: strings.TrimSpace(cfg.Host), password: cfg.Password,
		keyPrefix: strings.TrimSpace(cfg.KeyPrefix), database: defaultDatabase,
		poolSize: defaultPoolSize, minIdleConns: defaultMinIdleConns,
		dialTimeout: defaultDialTimeout, readTimeout: defaultReadTimeout, writeTimeout: defaultWriteTimeout,
	}
}

func (o clientOptions) settings() Settings {
	return Settings{
		KeyPrefix: o.keyPrefix, DB: o.database, PoolSize: o.poolSize,
		MinIdleConns: o.minIdleConns, DialTimeout: o.dialTimeout,
		ReadTimeout: o.readTimeout, WriteTimeout: o.writeTimeout,
	}
}

func newRedisDatabase(options clientOptions) redisDatabase {
	client := redis.NewClient(&redis.Options{
		Addr: options.address, Password: options.password, DB: options.database,
		PoolSize: options.poolSize, MinIdleConns: options.minIdleConns,
		DialTimeout: options.dialTimeout, ReadTimeout: options.readTimeout, WriteTimeout: options.writeTimeout,
	})
	addRedisKeyPrefixHook(client, options.keyPrefix)
	return &redisClient{client: client}
}

type redisDatabase interface {
	Ping(context.Context) error
	Close() error
}

type redisClient struct{ client *redis.Client }

func (c *redisClient) Ping(ctx context.Context) error { return c.client.Ping(ctx).Err() }
func (c *redisClient) Close() error                   { return c.client.Close() }

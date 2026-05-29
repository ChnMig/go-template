package rdb

import (
	"context"
	"errors"
	"strings"
	"sync"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"http-services/config"
)

var (
	client   *redis.Client
	clientMu sync.Mutex
)

var ErrMissingRedisHost = errors.New("redis.host is empty")

func Client() (*redis.Client, error) {
	if err := Init(); err != nil {
		return nil, err
	}
	if client == nil {
		return nil, errors.New("redis client not initialized")
	}
	return client, nil
}

func Init() error {
	clientMu.Lock()
	defer clientMu.Unlock()

	if client != nil {
		return nil
	}
	if strings.TrimSpace(config.RedisHost) == "" {
		return ErrMissingRedisHost
	}

	redisClient := redis.NewClient(&redis.Options{
		Addr:         config.RedisHost,
		Password:     config.RedisPassword,
		PoolSize:     100,
		MinIdleConns: 10,
	})

	if err := redisClient.Ping(context.Background()).Err(); err != nil {
		zap.L().Error("redis connection failed", zap.Error(err))
		return err
	}

	client = redisClient
	return nil
}

func GetClient() *redis.Client {
	redisClient, err := Client()
	if err != nil {
		zap.L().Error("redis client not initialized", zap.Error(err))
		return nil
	}
	return redisClient
}

func CloseClient() {
	clientMu.Lock()
	defer clientMu.Unlock()

	if client == nil {
		return
	}
	if err := client.Close(); err != nil {
		zap.L().Warn("close redis client failed", zap.Error(err))
	}
	client = nil
}

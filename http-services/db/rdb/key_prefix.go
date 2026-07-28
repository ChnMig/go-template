package rdb

import (
	"context"
	"net"
	"strconv"
	"strings"

	"github.com/redis/go-redis/v9"
)

type redisKeyPrefixHook struct {
	prefix string
}

func addRedisKeyPrefixHook(client *redis.Client, prefix string) {
	prefix = strings.TrimSpace(prefix)
	if client == nil || prefix == "" {
		return
	}
	client.AddHook(redisKeyPrefixHook{prefix: prefix})
}

func (h redisKeyPrefixHook) DialHook(next redis.DialHook) redis.DialHook {
	return func(ctx context.Context, network, address string) (net.Conn, error) {
		return next(ctx, network, address)
	}
}

func (h redisKeyPrefixHook) ProcessHook(next redis.ProcessHook) redis.ProcessHook {
	return func(ctx context.Context, command redis.Cmder) error {
		prefixRedisCommandArgs(h.prefix, command.Args())
		return next(ctx, command)
	}
}

func (h redisKeyPrefixHook) ProcessPipelineHook(next redis.ProcessPipelineHook) redis.ProcessPipelineHook {
	return func(ctx context.Context, commands []redis.Cmder) error {
		for _, command := range commands {
			prefixRedisCommandArgs(h.prefix, command.Args())
		}
		return next(ctx, commands)
	}
}

func prefixRedisCommandArgs(prefix string, arguments []any) {
	prefix = strings.TrimSpace(prefix)
	if prefix == "" || len(arguments) < 2 {
		return
	}
	switch redisCommandName(arguments[0]) {
	case "get", "set", "setnx", "getdel", "hget", "hset", "expire", "ttl", "sadd", "srem", "smembers", "incr", "decr", "incrby":
		prefixRedisKeyArg(prefix, arguments, 1)
	case "del", "exists", "mget":
		prefixRedisKeyArgRange(prefix, arguments, 1, len(arguments))
	case "rename":
		prefixRedisKeyArgRange(prefix, arguments, 1, 3)
	case "eval", "evalsha", "eval_ro", "evalsha_ro":
		prefixRedisScriptKeys(prefix, arguments)
	case "scan":
		prefixRedisScanMatch(prefix, arguments)
	case "keys":
		prefixRedisKeyArg(prefix, arguments, 1)
	}
}

func prefixRedisKeyArgRange(prefix string, arguments []any, start, end int) {
	if end > len(arguments) {
		end = len(arguments)
	}
	for index := start; index < end; index++ {
		prefixRedisKeyArg(prefix, arguments, index)
	}
}

func prefixRedisScriptKeys(prefix string, arguments []any) {
	if len(arguments) < 4 {
		return
	}
	keyCount, ok := redisScriptKeyCount(arguments[2])
	if !ok || keyCount <= 0 {
		return
	}
	prefixRedisKeyArgRange(prefix, arguments, 3, 3+keyCount)
}

func prefixRedisScanMatch(prefix string, arguments []any) {
	for index := 2; index < len(arguments)-1; index++ {
		if strings.EqualFold(redisArgString(arguments[index]), "match") {
			prefixRedisKeyArg(prefix, arguments, index+1)
			return
		}
	}
}

func prefixRedisKeyArg(prefix string, arguments []any, index int) {
	if index < 0 || index >= len(arguments) {
		return
	}
	switch key := arguments[index].(type) {
	case string:
		if key == "" || strings.HasPrefix(key, prefix) {
			return
		}
		arguments[index] = prefix + key
	case []byte:
		keyText := string(key)
		if keyText == "" || strings.HasPrefix(keyText, prefix) {
			return
		}
		arguments[index] = []byte(prefix + keyText)
	}
}

func redisCommandName(argument any) string {
	return strings.ToLower(redisArgString(argument))
}

func redisArgString(argument any) string {
	switch value := argument.(type) {
	case string:
		return value
	case []byte:
		return string(value)
	default:
		return ""
	}
}

func redisScriptKeyCount(argument any) (int, bool) {
	switch value := argument.(type) {
	case int:
		return value, true
	case int8:
		return int(value), true
	case int16:
		return int(value), true
	case int32:
		return int(value), true
	case int64:
		return int(value), true
	case uint:
		return int(value), true
	case uint8:
		return int(value), true
	case uint16:
		return int(value), true
	case uint32:
		return int(value), true
	case uint64:
		return int(value), true
	case string:
		count, err := strconv.Atoi(value)
		return count, err == nil
	case []byte:
		count, err := strconv.Atoi(string(value))
		return count, err == nil
	default:
		return 0, false
	}
}

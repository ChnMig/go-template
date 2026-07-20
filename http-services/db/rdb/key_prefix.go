package rdb

import (
	"context"
	"errors"
	"net"
	"strconv"
	"strings"

	"github.com/redis/go-redis/v9"
)

var ErrRedisScanRequiresMatch = errors.New("redis SCAN requires a non-empty MATCH when key prefix is active")

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
	return func(ctx context.Context, network string, addr string) (net.Conn, error) {
		return next(ctx, network, addr)
	}
}

func (h redisKeyPrefixHook) ProcessHook(next redis.ProcessHook) redis.ProcessHook {
	prefix := strings.TrimSpace(h.prefix)
	return func(ctx context.Context, cmd redis.Cmder) error {
		if err := validateRedisCommandArgs(prefix, cmd.Args()); err != nil {
			cmd.SetErr(err)
			return err
		}
		prefixRedisCommandArgs(prefix, cmd.Args())
		return next(ctx, cmd)
	}
}

func (h redisKeyPrefixHook) ProcessPipelineHook(next redis.ProcessPipelineHook) redis.ProcessPipelineHook {
	prefix := strings.TrimSpace(h.prefix)
	return func(ctx context.Context, cmds []redis.Cmder) error {
		for _, cmd := range cmds {
			if err := validateRedisCommandArgs(prefix, cmd.Args()); err != nil {
				for _, pipelineCmd := range cmds {
					pipelineCmd.SetErr(err)
				}
				return err
			}
		}
		for _, cmd := range cmds {
			prefixRedisCommandArgs(prefix, cmd.Args())
		}
		return next(ctx, cmds)
	}
}

func validateRedisCommandArgs(prefix string, args []any) error {
	if prefix == "" || len(args) == 0 || redisCommandName(args[0]) != "scan" {
		return nil
	}
	if _, ok := redisScanMatchArgIndex(args); !ok {
		return ErrRedisScanRequiresMatch
	}
	return nil
}

func prefixRedisCommandArgs(prefix string, args []any) {
	prefix = strings.TrimSpace(prefix)
	if prefix == "" || len(args) < 2 {
		return
	}

	switch redisCommandName(args[0]) {
	case "get", "set", "setnx", "setex", "getex", "getdel", "append", "incr", "decr", "incrby", "decrby", "incrbyfloat",
		"expire", "ttl", "persist", "pexpire", "pttl", "expireat", "pexpireat",
		"hget", "hset", "hdel", "hexists", "hgetall", "hmget", "hsetnx", "hincrby", "hincrbyfloat", "hlen", "hkeys", "hvals",
		"lpush", "rpush", "lpop", "rpop", "lrange", "llen", "lrem", "lset", "ltrim",
		"sadd", "srem", "smembers", "scard", "sismember", "smismember", "spop", "srandmember",
		"zadd", "zrem", "zrange", "zcard", "zscore", "zincrby", "hscan", "sscan", "zscan":
		prefixRedisKeyArg(prefix, args, 1)
	case "del", "exists", "mget", "unlink", "touch", "watch":
		prefixRedisKeyArgs(prefix, args[1:])
	case "mset", "msetnx":
		prefixRedisAlternatingKeys(prefix, args)
	case "rename", "renamenx", "copy":
		prefixRedisKeyArgs(prefix, args[1:min(3, len(args))])
	case "eval", "evalsha", "eval_ro", "evalsha_ro":
		prefixRedisScriptKeys(prefix, args)
	case "scan":
		prefixRedisScanMatch(prefix, args)
	case "keys":
		prefixRedisKeyArg(prefix, args, 1)
	}
}

func prefixRedisKeyArgs(prefix string, args []any) {
	for index := range args {
		prefixRedisKeyArg(prefix, args, index)
	}
}

func prefixRedisAlternatingKeys(prefix string, args []any) {
	for index := 1; index < len(args); index += 2 {
		prefixRedisKeyArg(prefix, args, index)
	}
}

func prefixRedisScriptKeys(prefix string, args []any) {
	if len(args) < 4 {
		return
	}
	keyCount, ok := redisScriptKeyCount(args[2])
	if !ok || keyCount <= 0 {
		return
	}
	keyCount = min(keyCount, len(args)-3)
	prefixRedisKeyArgs(prefix, args[3:3+keyCount])
}

func prefixRedisScanMatch(prefix string, args []any) {
	if index, ok := redisScanMatchArgIndex(args); ok {
		prefixRedisKeyArg(prefix, args, index)
	}
}

func redisScanMatchArgIndex(args []any) (int, bool) {
	for index := 2; index < len(args); index++ {
		if !strings.EqualFold(redisArgString(args[index]), "match") {
			continue
		}
		if index+1 >= len(args) || redisArgString(args[index+1]) == "" {
			return 0, false
		}
		return index + 1, true
	}
	return 0, false
}

func prefixRedisKeyArg(prefix string, args []any, index int) {
	if index < 0 || index >= len(args) {
		return
	}

	switch key := args[index].(type) {
	case string:
		if strings.HasPrefix(key, prefix) {
			return
		}
		args[index] = prefix + key
	case []byte:
		keyText := string(key)
		if strings.HasPrefix(keyText, prefix) {
			return
		}
		args[index] = []byte(prefix + keyText)
	}
}

func redisCommandName(arg any) string {
	return strings.ToLower(redisArgString(arg))
}

func redisArgString(arg any) string {
	switch value := arg.(type) {
	case string:
		return value
	case []byte:
		return string(value)
	default:
		return ""
	}
}

func redisScriptKeyCount(arg any) (int, bool) {
	switch value := arg.(type) {
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

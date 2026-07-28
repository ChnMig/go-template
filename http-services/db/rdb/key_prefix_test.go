package rdb

import (
	"context"
	"reflect"
	"testing"

	"github.com/redis/go-redis/v9"
)

func TestPrefixRedisCommandArgsPrefixesSimpleAndMultipleKeys(t *testing.T) {
	tests := []struct {
		args []any
		want []any
	}{
		{[]any{"hset", "hash:key", "field", "value"}, []any{"hset", "app:hash:key", "field", "value"}},
		{[]any{"del", "key:a", "key:b"}, []any{"del", "app:key:a", "app:key:b"}},
		{[]any{"rename", "live:key", "snapshot:key"}, []any{"rename", "app:live:key", "app:snapshot:key"}},
		{[]any{"evalsha", "sha", 2, "key:a", "key:b", "argv:a"}, []any{"evalsha", "sha", 2, "app:key:a", "app:key:b", "argv:a"}},
		{[]any{"scan", uint64(0), "match", "cache:*"}, []any{"scan", uint64(0), "match", "app:cache:*"}},
	}
	for _, testCase := range tests {
		command := redis.NewCmd(context.Background(), testCase.args...)
		prefixRedisCommandArgs("app:", command.Args())
		if !reflect.DeepEqual(command.Args(), testCase.want) {
			t.Fatalf("args = %#v, want %#v", command.Args(), testCase.want)
		}
	}
}

func TestPrefixRedisCommandArgsDoesNotDoublePrefix(t *testing.T) {
	command := redis.NewCmd(context.Background(), "get", "app:cache:key")
	prefixRedisCommandArgs("app:", command.Args())
	if !reflect.DeepEqual(command.Args(), []any{"get", "app:cache:key"}) {
		t.Fatalf("args = %#v", command.Args())
	}
}

func TestRedisKeyPrefixHookPrefixesPipelineCommands(t *testing.T) {
	hook := redisKeyPrefixHook{prefix: "app:"}
	commands := []redis.Cmder{
		redis.NewCmd(context.Background(), "get", "cache:a"),
		redis.NewCmd(context.Background(), "sadd", "index:a", "cache:a"),
	}
	called := false
	err := hook.ProcessPipelineHook(func(_ context.Context, got []redis.Cmder) error {
		called = true
		if !reflect.DeepEqual(got[0].Args(), []any{"get", "app:cache:a"}) {
			t.Fatalf("args = %#v", got[0].Args())
		}
		return nil
	})(context.Background(), commands)
	if err != nil || !called {
		t.Fatalf("error = %v, called = %v", err, called)
	}
}

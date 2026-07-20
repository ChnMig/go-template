package rdb

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
)

var _ redis.Hook = redisKeyPrefixHook{}

func TestPrefixRedisCommandArgsUsesGoRedisCommandLayouts(t *testing.T) {
	ctx := context.Background()
	pipe := newRedisTestPipeline(t)
	singleKeyCommands := []redis.Cmder{
		pipe.Get(ctx, "key"),
		pipe.Set(ctx, "key", "value", 0),
		pipe.SetNX(ctx, "key", "value", time.Minute),
		redis.NewCmd(ctx, "setnx", "key", "value"),
		pipe.GetDel(ctx, "key"),
		pipe.HGet(ctx, "hash", "field"),
		pipe.HSet(ctx, "hash", "field", "value"),
		pipe.Expire(ctx, "key", time.Minute),
		pipe.TTL(ctx, "key"),
		pipe.SAdd(ctx, "set", "member"),
		pipe.SRem(ctx, "set", "member"),
		pipe.SMembers(ctx, "set"),
		pipe.Incr(ctx, "counter"),
		pipe.Decr(ctx, "counter"),
		pipe.IncrBy(ctx, "counter", 2),
		pipe.SetEx(ctx, "key", "value", time.Minute),
		pipe.GetEx(ctx, "key", time.Minute),
		pipe.Append(ctx, "key", "value"),
		pipe.DecrBy(ctx, "counter", 2),
		pipe.IncrByFloat(ctx, "counter", 1.5),
		pipe.Persist(ctx, "key"),
		pipe.PExpire(ctx, "key", time.Minute),
		pipe.PTTL(ctx, "key"),
		pipe.ExpireAt(ctx, "key", time.Unix(100, 0)),
		pipe.PExpireAt(ctx, "key", time.Unix(100, 0)),
		pipe.HDel(ctx, "hash", "field"),
		pipe.HExists(ctx, "hash", "field"),
		pipe.HGetAll(ctx, "hash"),
		pipe.HMGet(ctx, "hash", "field:a", "field:b"),
		pipe.HSetNX(ctx, "hash", "field", "value"),
		pipe.HIncrBy(ctx, "hash", "field", 2),
		pipe.HIncrByFloat(ctx, "hash", "field", 1.5),
		pipe.HLen(ctx, "hash"),
		pipe.HKeys(ctx, "hash"),
		pipe.HVals(ctx, "hash"),
		pipe.LPush(ctx, "list", "value:a", "value:b"),
		pipe.RPush(ctx, "list", "value:a", "value:b"),
		pipe.LPop(ctx, "list"),
		pipe.RPop(ctx, "list"),
		pipe.LRange(ctx, "list", 0, 10),
		pipe.LLen(ctx, "list"),
		pipe.LRem(ctx, "list", 1, "value"),
		pipe.LSet(ctx, "list", 0, "value"),
		pipe.LTrim(ctx, "list", 0, 10),
		pipe.SCard(ctx, "set"),
		pipe.SIsMember(ctx, "set", "member"),
		pipe.SMIsMember(ctx, "set", "member:a", "member:b"),
		pipe.SPop(ctx, "set"),
		pipe.SRandMember(ctx, "set"),
		pipe.ZAdd(ctx, "sorted", redis.Z{Score: 1, Member: "member"}),
		pipe.ZRem(ctx, "sorted", "member"),
		pipe.ZRange(ctx, "sorted", 0, 10),
		pipe.ZCard(ctx, "sorted"),
		pipe.ZScore(ctx, "sorted", "member"),
		pipe.ZIncrBy(ctx, "sorted", 1.5, "member"),
		pipe.HScan(ctx, "hash", 0, "field:*", 10),
		pipe.SScan(ctx, "set", 0, "member:*", 10),
		pipe.ZScan(ctx, "sorted", 0, "member:*", 10),
		pipe.Keys(ctx, "cache:*"),
	}
	for _, cmd := range singleKeyCommands {
		assertOnlyRedisKeyArgsPrefixed(t, cmd, 1)
	}

	multiKeyCommands := []struct {
		cmd     redis.Cmder
		indexes []int
	}{
		{cmd: pipe.MSet(ctx, "key:a", "value:a", "key:b", "value:b"), indexes: []int{1, 3}},
		{cmd: pipe.MSetNX(ctx, "key:a", "value:a", "key:b", "value:b"), indexes: []int{1, 3}},
		{cmd: pipe.Del(ctx, "key:a", "key:b"), indexes: []int{1, 2}},
		{cmd: pipe.Exists(ctx, "key:a", "key:b"), indexes: []int{1, 2}},
		{cmd: pipe.MGet(ctx, "key:a", "key:b"), indexes: []int{1, 2}},
		{cmd: pipe.Unlink(ctx, "key:a", "key:b"), indexes: []int{1, 2}},
		{cmd: pipe.Touch(ctx, "key:a", "key:b"), indexes: []int{1, 2}},
		{cmd: pipe.Rename(ctx, "source", "destination"), indexes: []int{1, 2}},
		{cmd: pipe.RenameNX(ctx, "source", "destination"), indexes: []int{1, 2}},
		{cmd: pipe.Copy(ctx, "source", "destination", 1, true), indexes: []int{1, 2}},
		{cmd: redis.NewCmd(ctx, "watch", "key:a", "key:b"), indexes: []int{1, 2}},
		{cmd: pipe.Eval(ctx, "script", []string{"key:a", "key:b"}, "arg"), indexes: []int{3, 4}},
		{cmd: pipe.EvalSha(ctx, "sha", []string{"key:a", "key:b"}, "arg"), indexes: []int{3, 4}},
		{cmd: pipe.EvalRO(ctx, "script", []string{"key:a", "key:b"}, "arg"), indexes: []int{3, 4}},
		{cmd: pipe.EvalShaRO(ctx, "sha", []string{"key:a", "key:b"}, "arg"), indexes: []int{3, 4}},
	}
	for _, test := range multiKeyCommands {
		assertOnlyRedisKeyArgsPrefixed(t, test.cmd, test.indexes...)
	}
}

func TestPrefixRedisCommandArgsHandlesBoundaryInputs(t *testing.T) {
	tests := []struct {
		name   string
		prefix string
		args   []any
		want   []any
	}{
		{name: "empty string key", prefix: "app:", args: []any{"get", ""}, want: []any{"get", "app:"}},
		{name: "empty byte key", prefix: "app:", args: []any{"get", []byte{}}, want: []any{"get", []byte("app:")}},
		{name: "prefixed string key", prefix: "app:", args: []any{"get", "app:key"}, want: []any{"get", "app:key"}},
		{name: "prefixed byte key", prefix: "app:", args: []any{"get", []byte("app:key")}, want: []any{"get", []byte("app:key")}},
		{name: "empty prefix", prefix: "", args: []any{"get", "key"}, want: []any{"get", "key"}},
		{name: "whitespace prefix", prefix: " \t ", args: []any{"get", "key"}, want: []any{"get", "key"}},
		{name: "unknown command", prefix: "app:", args: []any{"publish", "channel", "payload"}, want: []any{"publish", "channel", "payload"}},
		{name: "invalid lua count", prefix: "app:", args: []any{"eval", "script", "two", "key", "arg"}, want: []any{"eval", "script", "two", "key", "arg"}},
		{name: "oversized lua count", prefix: "app:", args: []any{"eval", "script", 3, "key"}, want: []any{"eval", "script", 3, "app:key"}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cmd := redis.NewCmd(context.Background(), test.args...)
			prefixRedisCommandArgs(test.prefix, cmd.Args())
			assertRedisCommandArgs(t, cmd, test.want)
		})
	}
}

func TestRedisKeyPrefixHookRejectsScanWithoutNonEmptyMatch(t *testing.T) {
	ctx := context.Background()
	pipe := newRedisTestPipeline(t)
	tests := []struct {
		name string
		cmd  redis.Cmder
	}{
		{name: "builder omits empty match", cmd: pipe.Scan(ctx, 0, "", 10)},
		{name: "empty match operand", cmd: redis.NewCmd(ctx, "scan", uint64(0), "match", "")},
		{name: "missing match operand", cmd: redis.NewCmd(ctx, "scan", uint64(0), "match")},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			before := append([]any(nil), test.cmd.Args()...)
			nextCalled := false
			err := (redisKeyPrefixHook{prefix: "app:"}).ProcessHook(func(context.Context, redis.Cmder) error {
				nextCalled = true
				return nil
			})(ctx, test.cmd)
			if !errors.Is(err, ErrRedisScanRequiresMatch) || test.cmd.Err() != err {
				t.Fatalf("ProcessHook() error = %v, cmd error = %v", err, test.cmd.Err())
			}
			if nextCalled || !reflect.DeepEqual(test.cmd.Args(), before) {
				t.Fatalf("ProcessHook() called next or mutated args: %#v", test.cmd.Args())
			}
		})
	}
}

func TestRedisKeyPrefixHookRejectsPipelineBeforeAnyMutation(t *testing.T) {
	ctx := context.Background()
	pipe := newRedisTestPipeline(t)
	cmds := []redis.Cmder{pipe.Get(ctx, "key:a"), pipe.HSet(ctx, "hash", "field", "value"), pipe.Scan(ctx, 0, "", 10)}
	before := make([][]any, len(cmds))
	for index, cmd := range cmds {
		before[index] = append([]any(nil), cmd.Args()...)
	}
	nextCalled := false
	err := (redisKeyPrefixHook{prefix: "app:"}).ProcessPipelineHook(func(context.Context, []redis.Cmder) error {
		nextCalled = true
		return nil
	})(ctx, cmds)
	if !errors.Is(err, ErrRedisScanRequiresMatch) || nextCalled {
		t.Fatalf("ProcessPipelineHook() error = %v, nextCalled = %v", err, nextCalled)
	}
	for index, cmd := range cmds {
		if cmd.Err() != err || !reflect.DeepEqual(cmd.Args(), before[index]) {
			t.Fatalf("command %d error = %v, args = %#v", index, cmd.Err(), cmd.Args())
		}
	}
}

func TestRedisKeyPrefixHookProcessesScopedAndUnknownCommands(t *testing.T) {
	ctx := context.Background()
	pipe := newRedisTestPipeline(t)
	scan := pipe.Scan(ctx, 0, "cache:*", 10)
	unknown := redis.NewCmd(ctx, "publish", "channel", "payload")
	cmds := []redis.Cmder{scan, pipe.Get(ctx, "key"), unknown}
	nextCalled := false
	err := (redisKeyPrefixHook{prefix: " app: "}).ProcessPipelineHook(func(_ context.Context, got []redis.Cmder) error {
		nextCalled = true
		assertRedisCommandArgs(t, got[0], []any{"scan", uint64(0), "match", "app:cache:*", "count", int64(10)})
		assertRedisCommandArgs(t, got[1], []any{"get", "app:key"})
		assertRedisCommandArgs(t, got[2], []any{"publish", "channel", "payload"})
		return nil
	})(ctx, cmds)
	if err != nil || !nextCalled {
		t.Fatalf("ProcessPipelineHook() error = %v, nextCalled = %v", err, nextCalled)
	}
	direct := pipe.Get(ctx, "direct")
	err = (redisKeyPrefixHook{prefix: "app:"}).ProcessHook(func(_ context.Context, got redis.Cmder) error {
		assertRedisCommandArgs(t, got, []any{"get", "app:direct"})
		return nil
	})(ctx, direct)
	if err != nil {
		t.Fatalf("ProcessHook() error = %v", err)
	}

	unscopedScan := pipe.Scan(ctx, 0, "", 10)
	err = (redisKeyPrefixHook{prefix: " \t "}).ProcessHook(func(context.Context, redis.Cmder) error { return nil })(ctx, unscopedScan)
	if err != nil {
		t.Fatalf("ProcessHook() with inactive prefix error = %v", err)
	}
}

func newRedisTestPipeline(t *testing.T) redis.Pipeliner {
	t.Helper()
	client := redis.NewClient(&redis.Options{})
	pipe := client.Pipeline()
	t.Cleanup(func() {
		if err := client.Close(); err != nil {
			t.Errorf("close redis client: %v", err)
		}
	})
	return pipe
}

func assertOnlyRedisKeyArgsPrefixed(t *testing.T, cmd redis.Cmder, indexes ...int) {
	t.Helper()
	want := append([]any(nil), cmd.Args()...)
	for _, index := range indexes {
		want[index] = "app:" + want[index].(string)
	}
	prefixRedisCommandArgs("app:", cmd.Args())
	assertRedisCommandArgs(t, cmd, want)
}

func assertRedisCommandArgs(t *testing.T, cmd redis.Cmder, want []any) {
	t.Helper()
	if !reflect.DeepEqual(cmd.Args(), want) {
		t.Fatalf("cmd.Args() = %#v, want %#v", cmd.Args(), want)
	}
}

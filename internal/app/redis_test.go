package app

import (
	"testing"

	"nextopsx/internal/dbexec"
	hoststore "nextopsx/internal/store/hosts"
)

func TestRedisQuote(t *testing.T) {
	if got := RedisQuote(`a"b\c`); got != `"a\"b\\c"` {
		t.Fatalf("quote = %q", got)
	}
}

func TestRedisScanResult(t *testing.T) {
	res := &dbexec.Result{Rows: [][]any{
		{int64(1), "42"},
		{int64(2), []any{"user:1", "user:2"}},
	}}
	cursor, keys := redisScanResult(res)
	if cursor != "42" || len(keys) != 2 || keys[0] != "user:1" {
		t.Fatalf("scan = %q %v", cursor, keys)
	}
}

func TestRedisParseInfo(t *testing.T) {
	metrics := redisParseInfo("# Server\nredis_version:7.2.0\r\nused_memory_human:1.2M\nunknown:1\n")
	if len(metrics) != 2 {
		t.Fatalf("metrics = %+v", metrics)
	}
	if metrics[0].Name != "redis_version" || metrics[0].Value != "7.2.0" {
		t.Fatalf("first = %+v", metrics[0])
	}
}

func TestRedisTTLTextAndHistory(t *testing.T) {
	if RedisTTLText(-1) != "永久" || RedisTTLText(-2) != "不存在" || RedisTTLText(60) != "60s" {
		t.Fatalf("ttl labels wrong")
	}
	a := &App{}
	a.Redis.CommandHistory = []string{"PING", "GET x"}
	a.Redis.HistoryOpen = true
	a.RedisPickHistory("GET x")
	if a.Redis.Command != "GET x" || a.Redis.HistoryOpen {
		t.Fatalf("history pick = %q open=%v", a.Redis.Command, a.Redis.HistoryOpen)
	}
}

func TestAttachRedisWithoutBackend(t *testing.T) {
	a := &App{}
	n := &hoststore.TreeNode{
		ID: "r1", Name: "redis", NodeType: "host", Protocol: "redis",
		Params: map[string]any{"db": "2"},
	}
	a.AttachRedis(n)
	if a.Redis.HostID != "r1" || a.Redis.DBLabel != "redis · db2" {
		t.Fatalf("state = %+v", a.Redis)
	}
	if a.Redis.KeysErr != "后端未初始化" {
		t.Fatalf("keys err = %q", a.Redis.KeysErr)
	}
	a.RedisLoadInfo()
	if a.Redis.InfoErr != "后端未初始化" {
		t.Fatalf("info err = %q", a.Redis.InfoErr)
	}

	a.Redis.KeysErr = ""
	a.RedisSelectKey("user:1")
	if a.Redis.Detail.Loading || a.Redis.Detail.Err != "后端未初始化" {
		t.Fatalf("detail = %+v", a.Redis.Detail)
	}

	a.Redis.Command = "PING"
	a.RedisRunCommand()
	if a.Redis.CommandErr != "后端未初始化" || a.Redis.Command != "" {
		t.Fatalf("command = %q err=%q", a.Redis.Command, a.Redis.CommandErr)
	}
}

func TestRedisEditStringValue(t *testing.T) {
	a := &App{}
	a.Redis.Detail = RedisDetail{
		Key: "user:1", Type: "string",
		Result: &dbexec.Result{Columns: []string{"value"}, Rows: [][]any{{"alice"}}},
	}
	a.RedisOpenEditValue()
	if !a.Redis.EditOpen || a.Redis.EditValue != "alice" || a.Redis.EditKey != "user:1" {
		t.Fatalf("edit state = %+v", a.Redis)
	}

	a.Redis.TTLValue = "not-a-number"
	a.RedisSaveTTL()
	if a.Redis.Err != "TTL 需要是整数秒" {
		t.Fatalf("ttl error = %q", a.Redis.Err)
	}
}

func TestRedisTTLCommand(t *testing.T) {
	if got := redisTTLCommand("user:1", 60); got != `EXPIRE "user:1" 60` {
		t.Fatalf("expire = %q", got)
	}
	if got := redisTTLCommand("user:1", -1); got != `PERSIST "user:1"` {
		t.Fatalf("persist = %q", got)
	}
}

func TestRedisSetValueKeepsTTL(t *testing.T) {
	if got := redisSetValueCommand("user:1", "alice"); got != `SET "user:1" "alice" KEEPTTL` {
		t.Fatalf("set = %q", got)
	}
}

func TestRedisScanCollection(t *testing.T) {
	res := &dbexec.Result{Rows: [][]any{
		{int64(1), "7"},
		{int64(2), []any{"a", "1", "b", "2"}},
	}}
	parsed, truncated, ok := redisScanCollection("hash", res)
	if !ok || !truncated || len(parsed.Rows) != 2 ||
		parsed.Rows[0][0] != "a" || parsed.Rows[0][1] != "1" {
		t.Fatalf("hash parse = %+v truncated=%v ok=%v", parsed, truncated, ok)
	}

	res = &dbexec.Result{Rows: [][]any{
		{int64(1), "0"},
		{int64(2), []any{"x", "y"}},
	}}
	parsed, truncated, ok = redisScanCollection("set", res)
	if !ok || truncated || len(parsed.Rows) != 2 || parsed.Rows[1][1] != "y" {
		t.Fatalf("set parse = %+v truncated=%v ok=%v", parsed, truncated, ok)
	}
}

func TestRedisTTLGuards(t *testing.T) {
	a := &App{}
	a.Redis.Detail = RedisDetail{Key: "k", Type: "string", Loading: true}
	a.RedisOpenTTL()
	if a.Redis.TTLOpen || a.Redis.Err != "键信息尚未加载完成" {
		t.Fatalf("loading guard failed: open=%v err=%q", a.Redis.TTLOpen, a.Redis.Err)
	}

	a.Redis.Detail = RedisDetail{Key: "k", Type: "string"}
	a.RedisOpenTTL()
	if a.Redis.TTLValue != "" {
		t.Fatalf("unknown TTL should not prefill 0: %q", a.Redis.TTLValue)
	}
	a.Redis.TTLValue = "0"
	a.RedisSaveTTL()
	if a.Redis.Err == "" {
		t.Fatal("TTL 0 should be rejected")
	}
}

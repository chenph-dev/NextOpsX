package workspace

import (
	"slices"
	"testing"

	"nextopsx/domain/dbexec"
	hoststore "nextopsx/store/hosts"
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

type fakeRedisScan struct {
	results []*dbexec.Result
	queries []string
}

func (f *fakeRedisScan) ExecuteSQL(_, sqlText string) (*dbexec.Result, error) {
	f.queries = append(f.queries, sqlText)
	if len(f.results) == 0 {
		return nil, nil
	}
	res := f.results[0]
	f.results = f.results[1:]
	return res, nil
}

func redisScanRows(cursor string, keys ...string) *dbexec.Result {
	items := make([]any, len(keys))
	for i, key := range keys {
		items[i] = key
	}
	return &dbexec.Result{Rows: [][]any{
		{int64(1), cursor},
		{int64(2), items},
	}}
}

func TestRedisScanKeysPage(t *testing.T) {
	exec := &fakeRedisScan{results: []*dbexec.Result{
		redisScanRows("7", "a", "b", "c"),
		redisScanRows("9", "d", "a", "e", "f", "g"),
	}}
	keys, next, pending, err := scanRedisKeysPage(exec, "h1",
		redisPageStart{cursor: "0"}, "user:*", 5)
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(keys, []string{"a", "b", "c", "d", "e"}) {
		t.Fatalf("keys = %v", keys)
	}
	if next != "9" {
		t.Fatalf("next = %q", next)
	}
	if !slices.Equal(pending, []string{"f", "g"}) {
		t.Fatalf("pending = %v", pending)
	}
	wantQueries := []string{
		`SCAN 0 MATCH "user:*" COUNT 5`,
		`SCAN 7 MATCH "user:*" COUNT 2`,
	}
	if !slices.Equal(exec.queries, wantQueries) {
		t.Fatalf("queries = %q", exec.queries)
	}

	exec = &fakeRedisScan{results: []*dbexec.Result{redisScanRows("0", "h")}}
	keys, next, pending, err = scanRedisKeysPage(exec, "h1",
		redisPageStart{cursor: "9", pending: []string{"f", "g"}}, "", 5)
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(keys, []string{"f", "g", "h"}) || next != "0" || len(pending) != 0 {
		t.Fatalf("page 2: keys=%v next=%q pending=%v", keys, next, pending)
	}
	if !slices.Equal(exec.queries, []string{"SCAN 9 COUNT 3"}) {
		t.Fatalf("page 2 queries = %q", exec.queries)
	}
}

func TestRecordRedisPageStart(t *testing.T) {
	a := &State{}
	a.Redis.pageStarts = []redisPageStart{{cursor: "0"}}
	a.recordRedisPageStart(0, "9", []string{"x"})
	if len(a.Redis.pageStarts) != 2 || a.Redis.pageStarts[1].cursor != "9" ||
		len(a.Redis.pageStarts[1].pending) != 1 {
		t.Fatalf("page starts = %+v", a.Redis.pageStarts)
	}
	a.recordRedisPageStart(0, "0", nil)
	if len(a.Redis.pageStarts) != 1 {
		t.Fatalf("stale page kept: %+v", a.Redis.pageStarts)
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
	a := &State{}
	a.Redis.CommandHistory = []string{"PING", "GET x"}
	a.Redis.HistoryOpen = true
	a.RedisPickHistory("GET x")
	if a.Redis.Command != "GET x" || a.Redis.HistoryOpen {
		t.Fatalf("history pick = %q open=%v", a.Redis.Command, a.Redis.HistoryOpen)
	}
}

func TestAttachRedisWithoutBackend(t *testing.T) {
	a := &State{}
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
	a := &State{}
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
	a := &State{}
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

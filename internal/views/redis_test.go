package views

import (
	"testing"

	"github.com/egoist/mygo/ui"

	"nextopsx/internal/app"
	"nextopsx/internal/dbexec"
	hoststore "nextopsx/internal/store/hosts"
)

func redisTestApp() *app.App {
	a := &app.App{RightTab: 0}
	a.Selected = &hoststore.TreeNode{
		ID: "r1", Name: "redis-01", NodeType: "host", Protocol: "redis",
		Addr: "10.0.0.10", Port: 6379, User: "default",
		Params: map[string]any{"db": "2"},
	}
	a.Redis = app.RedisState{
		HostID: "r1", HostName: "redis-01", DBLabel: "redis-01 · db2",
		Keys:        []string{"user:1", "session:2"},
		Cursor:      "13",
		Pattern:     "user:*",
		SelectedKey: "user:1",
		Detail: app.RedisDetail{
			Key: "user:1", Type: "string", TTL: 60, TTLKnown: true,
			Result: &dbexec.Result{Columns: []string{"value"}, Rows: [][]any{{"alice"}}},
		},
		Info: []app.RedisMetric{
			{Name: "redis_version", Value: "7.2.0"},
			{Name: "used_memory_human", Value: "1.2M"},
		},
		CommandResult:  &dbexec.Result{Columns: []string{"value"}, Rows: [][]any{{"PONG"}}},
		CommandHistory: []string{"PING", "GET user:1"},
	}
	return a
}

func TestRedisPanelRenders(t *testing.T) {
	a := redisTestApp()
	tt := ui.NewTester(Shell(a), 1300, 900)
	for _, want := range []string{
		"Redis", "redis-01 · db2", "键 (2)", "游标 13", "搜索", "下一页",
		"user:1", "session:2", "键详情", "服务器信息", "类型 string", "TTL 60s",
		"编辑值", "设置 TTL", "删除", "alice", "命令终端", "执行", "PONG",
	} {
		if !tt.HasText(want) {
			t.Errorf("missing %q in %q", want, tt.Texts())
		}
	}
}

func TestRedisInfoTab(t *testing.T) {
	a := redisTestApp()
	a.Redis.MainTab = 1
	tt := ui.NewTester(Shell(a), 1300, 900)
	for _, want := range []string{"服务器信息", "刷新", "redis_version", "7.2.0", "used_memory_human", "1.2M"} {
		if !tt.HasText(want) {
			t.Errorf("missing %q in %q", want, tt.Texts())
		}
	}
}

func TestRedisDialogs(t *testing.T) {
	a := redisTestApp()
	a.Redis.EditOpen = true
	a.Redis.EditKey = "user:1"
	a.Redis.EditValue = "alice"
	tt := ui.NewTester(Shell(a), 1300, 900)
	if !tt.HasText("编辑字符串值") || !tt.HasText("user:1") {
		t.Fatalf("edit dialog missing: %q", tt.Texts())
	}
	a.Redis.EditOpen = false
	a.Redis.TTLOpen = true
	a.Redis.TTLKey = "user:1"
	a.Redis.TTLValue = "60"
	tt.Frame()
	if !tt.HasText("设置 TTL") || !tt.HasText("-1 表示永久") {
		t.Fatalf("ttl dialog missing: %q", tt.Texts())
	}
	a.Redis.TTLOpen = false
	a.Redis.ConfirmDeleteKey = "user:1"
	tt.Frame()
	if !tt.HasText("删除键") || !tt.HasText("确定删除「user:1」吗？") {
		t.Fatalf("delete dialog missing: %q", tt.Texts())
	}
}

func TestRedisEmptySelection(t *testing.T) {
	a := redisTestApp()
	a.Redis.SelectedKey = ""
	a.Redis.Detail = app.RedisDetail{}
	tt := ui.NewTester(Shell(a), 1300, 900)
	if !tt.HasText("选择左侧的键查看详情") {
		t.Fatalf("selection hint missing: %q", tt.Texts())
	}
}

func TestRedisTruncatedPreview(t *testing.T) {
	a := redisTestApp()
	a.Redis.Detail.Truncated = true
	tt := ui.NewTester(Shell(a), 1300, 900)
	if !tt.HasText("仅显示前 200 项") {
		t.Fatalf("truncated hint missing: %q", tt.Texts())
	}
}

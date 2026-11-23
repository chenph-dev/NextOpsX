package redis_test

import (
	"fmt"
	"testing"

	"github.com/egoist/mygo/ui"

	"nextopsx/domain/dbexec"
	hoststore "nextopsx/store/hosts"
	"nextopsx/ui/views"
	"nextopsx/ui/workspace"
)

func redisTestApp() *workspace.State {
	a := &workspace.State{}
	a.Selected = &hoststore.TreeNode{
		ID: "r1", Name: "redis-01", NodeType: "host", Protocol: "redis",
		Addr: "10.0.0.10", Port: 6379, User: "default",
		Params: map[string]any{"db": "2"},
	}
	a.Redis = workspace.RedisState{
		HostID: "r1", HostName: "redis-01", DBLabel: "redis-01 · db2",
		Keys:        []string{"user:1", "session:2"},
		Cursor:      "13",
		Pattern:     "user:*",
		SelectedKey: "user:1",
		Detail: workspace.RedisDetail{
			Key: "user:1", Type: "string", TTL: 60, TTLKnown: true,
			Result: &dbexec.Result{Columns: []string{"value"}, Rows: [][]any{{"alice"}}},
		},
		Info: []workspace.RedisMetric{
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
	tt := ui.NewTester(views.Shell(a), 1300, 900)
	for _, want := range []string{
		"键 (2)", "游标 13", "刷新键", "搜索", "上一页", "第 1 页", "下一页", "每页 20",
		"user:1", "session:2", "键详情", "服务器信息", "类型 string", "TTL 60s",
		"编辑值", "设置 TTL", "删除", "alice", "命令终端", "执行", "PONG",
	} {
		if !tt.HasText(want) {
			t.Errorf("missing %q in %q", want, tt.Texts())
		}
	}
}

func TestRedisRefreshIconInKeysHeader(t *testing.T) {
	a := redisTestApp()
	tt := ui.NewTester(views.Shell(a), 1300, 900)
	title, ok := tt.Find("键 (2)")
	if !ok {
		t.Fatalf("keys header not found: %q", tt.Texts())
	}
	r, ok := tt.Find("刷新键")
	if !ok {
		t.Fatalf("refresh icon not found: %q", tt.Texts())
	}
	if r.X < title.X || r.X+r.W > title.X+280 {
		t.Fatalf("refresh icon outside the keys pane: x=%.1f w=%.1f pane=%.1f", r.X, r.W, title.X)
	}
	if cy := r.Y + r.H/2; cy < title.Y || cy > title.Y+title.H {
		t.Fatalf("refresh icon not on the keys header row: y=%.1f header=%.1f..%.1f",
			cy, title.Y, title.Y+title.H)
	}
	if tt.HasText("redis-01 · db2") {
		t.Fatal("top header and status label should be gone")
	}
	if tt.HasText("就绪") {
		t.Fatal("empty status bar should be gone")
	}
}

func TestRedisPagerBelowKeys(t *testing.T) {
	a := redisTestApp()
	tt := ui.NewTester(views.Shell(a), 1300, 900)
	pane, ok := tt.Find("键 (2)")
	if !ok {
		t.Fatalf("keys header not found: %q", tt.Texts())
	}
	last, ok := tt.Find("session:2")
	if !ok {
		t.Fatalf("last key not found: %q", tt.Texts())
	}
	for _, label := range []string{"上一页", "下一页"} {
		r, ok := tt.Find(label)
		if !ok {
			t.Fatalf("%q not found: %q", label, tt.Texts())
		}
		if r.Y < last.Y+last.H {
			t.Fatalf("%q is not below the key list: y=%.1f list bottom=%.1f",
				label, r.Y, last.Y+last.H)
		}
		if r.X < pane.X-8 || r.X+r.W > pane.X+272 {
			t.Fatalf("%q outside the keys pane: x=%.1f w=%.1f pane=%.1f",
				label, r.X, r.W, pane.X)
		}
	}
}

func TestRedisKeysFillPane(t *testing.T) {
	a := redisTestApp()
	a.Redis.Keys = make([]string, 20)
	for i := range a.Redis.Keys {
		a.Redis.Keys[i] = fmt.Sprintf("device:%02d", i)
	}
	tt := ui.NewTester(views.Shell(a), 1300, 900)
	last, ok := tt.Find("device:19")
	if !ok {
		t.Fatalf("last key not laid out: %q", tt.Texts())
	}
	prev, ok := tt.Find("上一页")
	if !ok {
		t.Fatalf("pager not found: %q", tt.Texts())
	}
	if last.Y+last.H > prev.Y {
		t.Fatalf("last key clipped by pager: key bottom=%.1f pager=%.1f",
			last.Y+last.H, prev.Y)
	}
}

func TestRedisInfoTab(t *testing.T) {
	a := redisTestApp()
	a.Redis.MainTab = 1
	tt := ui.NewTester(views.Shell(a), 1300, 900)
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
	tt := ui.NewTester(views.Shell(a), 1300, 900)
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
	a.Redis.Detail = workspace.RedisDetail{}
	tt := ui.NewTester(views.Shell(a), 1300, 900)
	if !tt.HasText("选择左侧的键查看详情") {
		t.Fatalf("selection hint missing: %q", tt.Texts())
	}
}

func TestRedisTruncatedPreview(t *testing.T) {
	a := redisTestApp()
	a.Redis.Detail.Truncated = true
	tt := ui.NewTester(views.Shell(a), 1300, 900)
	if !tt.HasText("仅显示前 200 项") {
		t.Fatalf("truncated hint missing: %q", tt.Texts())
	}
}

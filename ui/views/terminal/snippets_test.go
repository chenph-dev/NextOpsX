package terminal_test

import (
	"testing"

	"github.com/egoist/mygo/ui"

	"nextopsx/i18n"
	historystore "nextopsx/store/history"
	hoststore "nextopsx/store/hosts"
	snippetsstore "nextopsx/store/snippets"
	"nextopsx/ui/views/terminal"
	"nextopsx/ui/workspace"
)

func snippetsTestApp() *workspace.State {
	a := &workspace.State{View: workspace.ViewSettings}
	a.Settings.Tab = workspace.SettingsSnippets
	a.Selected = &hoststore.TreeNode{
		ID: "h1", NodeType: "host", Name: "web-01", Addr: "10.0.0.5", User: "ops", Port: 22, Protocol: "ssh",
	}
	a.Snippets.Items = []snippetsstore.Snippet{
		{ID: "s1", Name: "磁盘占用", Group: "排障", Command: "df -h"},
		{ID: "s2", Name: "当前用户", Group: "排障", Command: "whoami -- {{user}}"},
		{ID: "s3", Name: "系统日志", Group: "", Command: "tail -n 50 /var/log/syslog"},
	}
	a.Snippets.Selected = "s1"
	a.Snippets.Editing = true
	a.Snippets.Draft = a.Snippets.Items[0]
	return a
}

func TestCommandBubbleRendersAndFilters(t *testing.T) {
	i18n.SetLanguage(i18n.ZhCN)
	t.Cleanup(func() { i18n.SetLanguage(i18n.ZhCN) })

	a := snippetsTestApp()
	code := 0
	a.Snippets.History = []historystore.Entry{
		{ID: "h1", Source: "ai", Command: "free -h", ExitCode: &code, Output: "total used free", Ts: 1700000000},
		{ID: "h2", Source: "snippet", Command: "df -h", Ts: 1700000001},
	}
	tt := ui.NewTester(func(c *ui.Context) { terminal.SnippetPicker(c, a, c.Theme()) }, 600, 780)
	for _, want := range []string{
		"命令历史", "管理收藏", "free -h", "df -h", "磁盘占用",
		"全部", "历史", "收藏", "输出",
		"插入", "执行", "请先连接 SSH 终端",
	} {
		if !tt.HasText(want) {
			t.Errorf("missing %q in %q", want, tt.Texts())
		}
	}
	// df -h exists as both a favorite and history; the merged bubble shows it once.
	n := 0
	for _, text := range tt.Texts() {
		if text == "df -h" {
			n++
		}
	}
	if n != 1 {
		t.Fatalf("df -h rendered %d times, want once: %q", n, tt.Texts())
	}

	if err := tt.Click("输出"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if !tt.HasText("total used free") {
		t.Fatalf("history output did not expand: %q", tt.Texts())
	}

	a.Snippets.CommandQuery = "free -h"
	tt.Frame()
	if !tt.HasText("free -h") || tt.HasText("df -h") {
		t.Fatalf("command search filter failed: %q", tt.Texts())
	}
}

func TestCommandBubbleFilterTabs(t *testing.T) {
	i18n.SetLanguage(i18n.ZhCN)
	t.Cleanup(func() { i18n.SetLanguage(i18n.ZhCN) })

	a := snippetsTestApp()
	a.Snippets.History = []historystore.Entry{
		{ID: "h1", Source: "ai", Command: "free -h", Ts: 1700000000},
		{ID: "h2", Source: "snippet", Command: "df -h", Ts: 1700000001},
	}
	tt := ui.NewTester(func(c *ui.Context) { terminal.SnippetPicker(c, a, c.Theme()) }, 600, 780)
	tt.Frame()

	a.Snippets.PickerTab = workspace.PickerTabFavorites
	tt.Frame()
	if !tt.HasText("df -h") || !tt.HasText("当前用户") || tt.HasText("free -h") {
		t.Fatalf("favorites tab = %q", tt.Texts())
	}

	a.Snippets.PickerTab = workspace.PickerTabHistory
	tt.Frame()
	if !tt.HasText("free -h") || !tt.HasText("df -h") || tt.HasText("当前用户") {
		t.Fatalf("history tab = %q", tt.Texts())
	}
}

func TestCompletionBubbleRendersMergedItems(t *testing.T) {
	i18n.SetLanguage(i18n.ZhCN)
	t.Cleanup(func() { i18n.SetLanguage(i18n.ZhCN) })

	a := snippetsTestApp()
	code := 0
	a.Snippets.History = []historystore.Entry{
		{ID: "h1", Command: "df -h", ExitCode: &code, Output: "ok", Ts: 1700000000},
		{ID: "h2", Command: "free -h", Ts: 1700000001},
	}
	a.Completion = workspace.CompletionState{
		Open: true, Items: a.CommandItems(), Selected: 0,
	}
	tt := ui.NewTester(func(c *ui.Context) {
		ui.Box(c).Fill().Children(func() {
			terminal.CompletionBubble(c, a, c.Theme())
		})
	}, 700, 520)
	for _, want := range []string{
		"命令建议", "df -h", "磁盘占用", "收藏 · 历史", "free -h",
		"↑↓ 选择 · Tab 填入 · Esc 关闭",
	} {
		if !tt.HasText(want) {
			t.Errorf("missing %q in %q", want, tt.Texts())
		}
	}
}

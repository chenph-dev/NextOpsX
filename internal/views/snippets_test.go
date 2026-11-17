package views

import (
	"testing"

	"github.com/egoist/mygo/ui"

	"nextopsx/internal/app"
	"nextopsx/internal/i18n"
	historystore "nextopsx/internal/store/history"
	hoststore "nextopsx/internal/store/hosts"
	snippetsstore "nextopsx/internal/store/snippets"
)

func snippetsTestApp() *app.App {
	a := &app.App{View: app.ViewSettings}
	a.Settings.Tab = app.SettingsSnippets
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

func TestSnippetsSettingsRenders(t *testing.T) {
	i18n.SetLanguage(i18n.ZhCN)
	t.Cleanup(func() { i18n.SetLanguage(i18n.ZhCN) })

	a := snippetsTestApp()
	tt := ui.NewTester(Shell(a), 1300, 900)
	for _, want := range []string{
		"收藏命令", "共 3 条收藏", "新建收藏",
		"磁盘占用", "当前用户", "系统日志", "排障", "未分组",
		"编辑收藏", "名称", "分组", "命令", "说明",
		"可用变量", "预览：df -h", "保存", "删除",
	} {
		if !tt.HasText(want) {
			t.Errorf("missing %q in %q", want, tt.Texts())
		}
	}
}

func TestSnippetsSearchAndNew(t *testing.T) {
	i18n.SetLanguage(i18n.ZhCN)
	t.Cleanup(func() { i18n.SetLanguage(i18n.ZhCN) })

	a := snippetsTestApp()
	tt := ui.NewTester(Shell(a), 1300, 900)
	a.Snippets.Query = "whoami"
	tt.Frame()
	if !tt.HasText("当前用户") || tt.HasText("磁盘占用") {
		t.Fatalf("search filter failed: %q", tt.Texts())
	}
	a.Snippets.Query = ""
	tt.Frame()
	if err := tt.Click("新建收藏"); err != nil {
		t.Fatal(err)
	}
	if !a.Snippets.Editing || a.Snippets.Draft.ID != "" || a.Snippets.Selected != "" {
		t.Fatalf("new favorite state = %+v", a.Snippets)
	}
}

func TestSnippetsEmptyAndDeleteDialog(t *testing.T) {
	i18n.SetLanguage(i18n.ZhCN)
	t.Cleanup(func() { i18n.SetLanguage(i18n.ZhCN) })

	a := &app.App{View: app.ViewSettings}
	a.Settings.Tab = app.SettingsSnippets
	tt := ui.NewTester(Shell(a), 1300, 900)
	if !tt.HasText("还没有收藏命令") || !tt.HasText("从命令历史收藏") {
		t.Fatalf("empty state missing: %q", tt.Texts())
	}

	a = snippetsTestApp()
	a.Snippets.ConfirmDelete = "s1"
	tt = ui.NewTester(Shell(a), 1300, 900)
	if !tt.HasText("删除收藏命令") || !tt.HasText("确定删除收藏「磁盘占用」吗？") {
		t.Fatalf("delete dialog missing: %q", tt.Texts())
	}
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
	tt := ui.NewTester(func(c *ui.Context) { snippetPicker(c, a, c.Theme()) }, 600, 780)
	for _, want := range []string{
		"命令历史", "管理收藏", "free -h", "df -h", "磁盘占用",
		"收藏 · 历史", "历史", "收藏", "exit 0", "输出",
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

func TestCompletionBubbleRendersMergedItems(t *testing.T) {
	i18n.SetLanguage(i18n.ZhCN)
	t.Cleanup(func() { i18n.SetLanguage(i18n.ZhCN) })

	a := snippetsTestApp()
	code := 0
	a.Snippets.History = []historystore.Entry{
		{ID: "h1", Command: "df -h", ExitCode: &code, Output: "ok", Ts: 1700000000},
		{ID: "h2", Command: "free -h", Ts: 1700000001},
	}
	a.Completion = app.CompletionState{
		Open: true, Items: a.CommandItems(), Selected: 0,
	}
	tt := ui.NewTester(func(c *ui.Context) {
		ui.Box(c).Fill().Children(func() {
			completionBubble(c, a, c.Theme())
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

func TestSettingsEntryOpensSnippets(t *testing.T) {
	i18n.SetLanguage(i18n.ZhCN)
	t.Cleanup(func() { i18n.SetLanguage(i18n.ZhCN) })

	a := &app.App{}
	tt := ui.NewTester(Shell(a), 1300, 900)
	if err := tt.Click("设置"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if err := tt.Click("收藏"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if a.Settings.Tab != app.SettingsSnippets {
		t.Fatalf("tab = %v, want favorites", a.Settings.Tab)
	}
	if !tt.HasText("新建收藏") {
		t.Fatalf("favorites section missing: %q", tt.Texts())
	}
}

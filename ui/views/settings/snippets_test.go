package settings_test

import (
	"testing"

	"github.com/egoist/mygo/ui"

	"nextopsx/i18n"
	hoststore "nextopsx/store/hosts"
	snippetsstore "nextopsx/store/snippets"
	"nextopsx/ui/views"
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

func TestSnippetsSettingsRenders(t *testing.T) {
	i18n.SetLanguage(i18n.ZhCN)
	t.Cleanup(func() { i18n.SetLanguage(i18n.ZhCN) })

	a := snippetsTestApp()
	tt := ui.NewTester(views.Shell(a), 1300, 900)
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
	tt := ui.NewTester(views.Shell(a), 1300, 900)
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

	a := &workspace.State{View: workspace.ViewSettings}
	a.Settings.Tab = workspace.SettingsSnippets
	tt := ui.NewTester(views.Shell(a), 1300, 900)
	if !tt.HasText("还没有收藏命令") || !tt.HasText("从命令历史收藏") {
		t.Fatalf("empty state missing: %q", tt.Texts())
	}

	a = snippetsTestApp()
	a.Snippets.ConfirmDelete = "s1"
	tt = ui.NewTester(views.Shell(a), 1300, 900)
	if !tt.HasText("删除收藏命令") || !tt.HasText("确定删除收藏「磁盘占用」吗？") {
		t.Fatalf("delete dialog missing: %q", tt.Texts())
	}
}

func TestSettingsEntryOpensSnippets(t *testing.T) {
	i18n.SetLanguage(i18n.ZhCN)
	t.Cleanup(func() { i18n.SetLanguage(i18n.ZhCN) })

	a := &workspace.State{}
	tt := ui.NewTester(views.Shell(a), 1300, 900)
	if err := tt.Click("设置"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if err := tt.Click("收藏"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if a.Settings.Tab != workspace.SettingsSnippets {
		t.Fatalf("tab = %v, want favorites", a.Settings.Tab)
	}
	if !tt.HasText("新建收藏") {
		t.Fatalf("favorites section missing: %q", tt.Texts())
	}
}

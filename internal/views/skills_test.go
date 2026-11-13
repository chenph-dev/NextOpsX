package views

import (
	"strings"
	"testing"

	"github.com/egoist/mygo/ui"

	"nextopsx/internal/app"
	"nextopsx/internal/skill"
)

func skillsTestApp() *app.App {
	a := &app.App{View: app.ViewSettings}
	a.Settings.Tab = app.SettingsSkills
	a.Skills = app.SkillsState{
		Items: []*skill.Skill{
			{Name: "disk-analysis", Title: "disk-analysis", Description: "分析磁盘占用", Enabled: true},
			{Name: "log-cleanup", Title: "log-cleanup", Description: "清理旧日志", Enabled: false},
		},
		Selected: "disk-analysis",
		Markdown: "---\nname: disk-analysis\ndescription: 分析磁盘占用\n---\n# 磁盘分析\n\n运行 df -h。\n",
	}
	return a
}

func TestSkillsPanelRenders(t *testing.T) {
	a := skillsTestApp()
	tt := ui.NewTester(Shell(a), 1300, 900)
	for _, want := range []string{
		"设置", "技能", "安装技能 ZIP", "返回", "已安装 (2)",
		"disk-analysis", "分析磁盘占用", "log-cleanup", "清理旧日志",
		"已启用", "停用", "删除", "磁盘分析", "运行 df -h。",
	} {
		if !tt.HasText(want) {
			t.Errorf("missing %q in %q", want, tt.Texts())
		}
	}
	if tt.HasText("name: disk-analysis") {
		t.Errorf("frontmatter should be stripped: %q", tt.Texts())
	}
}

func TestSkillsEmptyAndError(t *testing.T) {
	a := &app.App{View: app.ViewSettings}
	a.Settings.Tab = app.SettingsSkills
	tt := ui.NewTester(Shell(a), 1300, 900)
	if !tt.HasText("还没有安装技能") {
		t.Fatalf("empty hint missing: %q", tt.Texts())
	}
	a.Skills.Err = "读取技能失败：磁盘错误"
	tt.Frame()
	if !tt.HasText("读取技能失败：磁盘错误") || !tt.HasText("重试") {
		t.Fatalf("error state missing: %q", tt.Texts())
	}
}

func TestSkillsDeleteDialog(t *testing.T) {
	a := skillsTestApp()
	a.Skills.ConfirmDelete = "disk-analysis"
	tt := ui.NewTester(Shell(a), 1300, 900)
	if !tt.HasText("删除技能") || !tt.HasText("确定删除技能「disk-analysis」吗？") {
		t.Fatalf("delete dialog missing: %q", tt.Texts())
	}
}

func TestSettingsEntryOpensSkills(t *testing.T) {
	a := &app.App{}
	tt := ui.NewTester(Shell(a), 1300, 900)
	if err := tt.Click("设置"); err != nil {
		t.Fatal(err)
	}
	if a.View != app.ViewSettings {
		t.Fatal("settings did not open")
	}
	if err := tt.Click("技能"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if a.Settings.Tab != app.SettingsSkills || a.View != app.ViewSettings {
		t.Fatalf("tab = %v, view = %v", a.Settings.Tab, a.View)
	}
	if !tt.HasText("安装技能 ZIP") {
		t.Fatalf("skills section missing: %q", tt.Texts())
	}
}

func TestSkillMarkdownBody(t *testing.T) {
	md := "---\nname: x\n---\n# Title\nbody"
	if got := skillMarkdownBody(md); strings.Contains(got, "name: x") || !strings.Contains(got, "# Title") {
		t.Fatalf("body = %q", got)
	}
}

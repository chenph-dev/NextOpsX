package views

import (
	"testing"

	"github.com/egoist/mygo/ui"

	"nextopsx/internal/app"
	configstore "nextopsx/internal/store/config"
	logsstore "nextopsx/internal/store/logs"
)

func TestSettingsModelShowsProviders(t *testing.T) {
	a := &app.App{View: app.ViewSettings}
	a.Settings.Tab = app.SettingsModel
	a.Settings.Providers = []configstore.AIProvider{
		{ID: "p1", Name: "默认供应商", Protocol: "openai", Model: "deepseek-chat", IsDefault: true},
		{ID: "p2", Name: "备用 Claude", Protocol: "claude", Model: "claude-sonnet-5"},
	}
	a.Settings.SelectedProvider = "p1"
	a.Settings.Draft = a.Settings.Providers[0]
	a.Settings.Draft.APIKey = "sk-x"

	tt := ui.NewTester(Shell(a), 1300, 900)
	for _, want := range []string{
		"设置", "模型", "新建供应商", "供应商 (2)",
		"默认供应商", "openai · deepseek-chat", "备用 Claude", "claude · claude-sonnet-5",
		"名称", "协议", "Base URL", "API Key", "删除", "保存",
	} {
		if !tt.HasText(want) {
			t.Errorf("missing %q in %q", want, tt.Texts())
		}
	}
	if tt.HasText("设为默认") {
		t.Errorf("已是默认供应商时不应显示「设为默认」: %q", tt.Texts())
	}
}

func TestSettingsGeneralSection(t *testing.T) {
	a := &app.App{View: app.ViewSettings}
	a.Settings.Theme = "dark"
	a.Settings.ThemeIndex = app.ThemeIndex("dark")
	a.Settings.Scale = 100
	a.Settings.ScaleIndex = app.ScaleIndex(100)
	a.Settings.DataDir = "/tmp/nextopsx-data"

	tt := ui.NewTester(Shell(a), 1300, 900)
	for _, want := range []string{"通用", "主题", "深色", "界面缩放", "125%", "数据目录", "/tmp/nextopsx-data"} {
		if !tt.HasText(want) {
			t.Errorf("missing %q in %q", want, tt.Texts())
		}
	}

	if err := tt.Click("深色"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if err := tt.Click("浅色"); err != nil {
		t.Fatal(err)
	}
	if a.Settings.Theme != "light" {
		t.Fatalf("theme = %q, want light", a.Settings.Theme)
	}
}

func TestSettingsAuditSection(t *testing.T) {
	a := &app.App{View: app.ViewSettings}
	a.Settings.Tab = app.SettingsAudit
	a.Audit = app.AuditState{
		Logs: []logsstore.CallLog{
			{
				ID: "l1", SessionID: "s1", Ts: 1700000000, Component: "model",
				Name: "llm", Provider: "OpenAI", TokensIn: 10, TokensOut: 20,
				TokensTotal: 30, DurationMS: 120, OK: true,
			},
			{
				ID: "l2", SessionID: "s1", Ts: 1700000001, Component: "tool",
				Name: "execute_command", Provider: "SSH", OK: false, Error: "boom",
			},
		},
		Summary: logsstore.TokenSummary{
			TotalCalls: 2, ModelCalls: 1, ToolCalls: 1,
			TokensIn: 10, TokensOut: 20, TokensTotal: 30,
		},
		Selected: 1,
	}

	tt := ui.NewTester(Shell(a), 1300, 900)
	for _, want := range []string{
		"审计", "刷新", "调用", "模型", "工具",
		"llm", "OpenAI", "execute_command", "SSH", "成功", "失败",
		"错误：boom",
	} {
		if !tt.HasText(want) {
			t.Errorf("missing %q in %q", want, tt.Texts())
		}
	}
}

func TestSettingsPolicySection(t *testing.T) {
	a := &app.App{View: app.ViewSettings}
	a.Settings.Tab = app.SettingsPolicy
	a.Policy.Loaded = true
	a.Policy.EnableAuto = true
	a.Policy.ReadOnly = "df -h\nfree -m"

	tt := ui.NewTester(Shell(a), 1300, 900)
	for _, want := range []string{
		"审批", "自动放行只读命令", "只读白名单",
		"清空（使用内置默认）", "保存",
	} {
		if !tt.HasText(want) {
			t.Errorf("missing %q in %q", want, tt.Texts())
		}
	}
}

func TestSettingsNewProviderDraft(t *testing.T) {
	a := &app.App{View: app.ViewSettings}
	a.Settings.Tab = app.SettingsModel
	a.NewAIProvider()

	tt := ui.NewTester(Shell(a), 1300, 900)
	for _, want := range []string{"新建供应商", "新供应商", "openai", "保存", "取消"} {
		if !tt.HasText(want) {
			t.Errorf("missing %q in %q", want, tt.Texts())
		}
	}
	if tt.HasText("删除") {
		t.Errorf("新建草稿不应显示「删除」: %q", tt.Texts())
	}
}

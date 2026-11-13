package views

import (
	"testing"

	"github.com/egoist/mygo/ui"

	"nextopsx/internal/app"
	configstore "nextopsx/internal/store/config"
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

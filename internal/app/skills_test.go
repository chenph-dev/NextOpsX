package app

import "testing"

func TestSkillsWithoutBackend(t *testing.T) {
	a := &App{}
	a.OpenSettings(SettingsSkills)
	if a.View != ViewSettings || a.Settings.Tab != SettingsSkills ||
		a.Skills.Loading || a.Skills.Err != "后端未初始化" {
		t.Fatalf("view = %v, state = %+v", a.View, a.Skills)
	}
	a.SelectSkill("x")
	if a.Skills.MarkdownErr != "后端未初始化" {
		t.Fatalf("markdown err = %q", a.Skills.MarkdownErr)
	}
	a.CloseSettings()
	if a.View != ViewAssets {
		t.Fatal("settings did not close")
	}
}

package configstore

import (
	"testing"

	"nextopsx/internal/store"
)

func TestUISettings_DefaultAndSave(t *testing.T) {
	t.Setenv("NEXTOPSX_DATA_DIR", t.TempDir())
	t.Setenv("NEXTOPSX_TEST_NO_KEYRING", "1")
	app, err := store.Open()
	if err != nil {
		t.Fatalf("store.Open: %v", err)
	}
	defer closeDB(app)

	s := NewUISettingsStore(app)
	got, err := s.GetUISettings()
	if err != nil {
		t.Fatalf("GetUISettings: %v", err)
	}
	if got.Theme != "system" || got.Scale != ScaleDefault || got.Language != "zh-CN" {
		t.Fatalf("默认界面设置 = %+v", got)
	}
	if got.AIHeight != DefaultAIHeight || got.SFTPWidth != DefaultSFTPWidth {
		t.Fatalf("默认抽屉尺寸 = %+v", got)
	}

	if err := s.SaveUISettings(UISettings{
		Theme: "dark", Scale: 125, Language: "en-US",
		AIOpen: true, AIHeight: 380, SFTPOpen: true, SFTPWidth: 400,
		LastAssetID: "h1",
	}); err != nil {
		t.Fatalf("SaveUISettings: %v", err)
	}
	got, _ = s.GetUISettings()
	if got.Theme != "dark" || got.Scale != 125 || got.Language != "en-US" {
		t.Fatalf("回读 = %+v", got)
	}
	if !got.AIOpen || got.AIHeight != 380 || !got.SFTPOpen || got.SFTPWidth != 400 ||
		got.LastAssetID != "h1" {
		t.Fatalf("抽屉回读 = %+v", got)
	}

	// 非法值回落到默认或边界。
	if err := s.SaveUISettings(UISettings{
		Theme: "neon", Scale: 999, Language: "fr-FR",
		AIOpen: false, AIHeight: 9999, SFTPOpen: false, SFTPWidth: 1,
	}); err != nil {
		t.Fatalf("SaveUISettings: %v", err)
	}
	got, _ = s.GetUISettings()
	if got.Theme != "system" || got.Scale != ScaleDefault || got.Language != "zh-CN" {
		t.Fatalf("非法值应回落默认，得到 %+v", got)
	}
	if got.AIHeight != MaxAIHeight || got.SFTPWidth != MinSFTPWidth {
		t.Fatalf("非法抽屉尺寸应钳制，得到 %+v", got)
	}
}

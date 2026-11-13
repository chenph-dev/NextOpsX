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
	if got.Theme != "system" {
		t.Fatalf("默认主题 = %q", got.Theme)
	}

	if err := s.SaveUISettings(UISettings{Theme: "dark"}); err != nil {
		t.Fatalf("SaveUISettings: %v", err)
	}
	got, _ = s.GetUISettings()
	if got.Theme != "dark" {
		t.Fatalf("回读主题 = %q", got.Theme)
	}

	// 非法值回落为跟随系统。
	if err := s.SaveUISettings(UISettings{Theme: "neon"}); err != nil {
		t.Fatalf("SaveUISettings: %v", err)
	}
	got, _ = s.GetUISettings()
	if got.Theme != "system" {
		t.Fatalf("非法主题应回落 system，得到 %q", got.Theme)
	}
}

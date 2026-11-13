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
	if got.Theme != "system" || got.Scale != ScaleDefault {
		t.Fatalf("默认界面设置 = %+v", got)
	}

	if err := s.SaveUISettings(UISettings{Theme: "dark", Scale: 125}); err != nil {
		t.Fatalf("SaveUISettings: %v", err)
	}
	got, _ = s.GetUISettings()
	if got.Theme != "dark" || got.Scale != 125 {
		t.Fatalf("回读 = %+v", got)
	}

	// 非法值回落到默认。
	if err := s.SaveUISettings(UISettings{Theme: "neon", Scale: 999}); err != nil {
		t.Fatalf("SaveUISettings: %v", err)
	}
	got, _ = s.GetUISettings()
	if got.Theme != "system" || got.Scale != ScaleDefault {
		t.Fatalf("非法值应回落默认，得到 %+v", got)
	}
}

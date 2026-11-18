package settings

import (
	"testing"

	"nextopsx/store"
	configstore "nextopsx/store/config"
)

func TestUISettingsFacade(t *testing.T) {
	t.Setenv("NEXTOPSX_DATA_DIR", t.TempDir())
	t.Setenv("NEXTOPSX_TEST_NO_KEYRING", "1")
	data, err := store.Open()
	if err != nil {
		t.Fatalf("store.Open: %v", err)
	}
	t.Cleanup(func() {
		if sqlDB, err := data.GORM().DB(); err == nil {
			_ = sqlDB.Close()
		}
	})
	s := New(configstore.NewUISettingsStore(data))

	got, err := s.GetUISettings()
	if err != nil || got.Theme != "system" || got.Scale != configstore.ScaleDefault {
		t.Fatalf("默认 UI 设置 = %+v err=%v", got, err)
	}
	if err := s.SaveUISettings(configstore.UISettings{
		Theme: "light", Scale: 110,
		AIOpen: true, AIHeight: 360, SFTPOpen: true, SFTPWidth: 380,
		LastAssetID: "h1",
	}); err != nil {
		t.Fatalf("SaveUISettings: %v", err)
	}
	got, _ = s.GetUISettings()
	if got.Theme != "light" || got.Scale != 110 {
		t.Fatalf("回读 = %+v", got)
	}
	if !got.AIOpen || got.AIHeight != 360 || !got.SFTPOpen || got.SFTPWidth != 380 ||
		got.LastAssetID != "h1" {
		t.Fatalf("抽屉回读 = %+v", got)
	}
	if s.DataDir() == "" {
		t.Fatal("DataDir 不应为空")
	}
}

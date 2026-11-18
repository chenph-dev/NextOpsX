package history

import (
	"testing"

	"nextopsx/store"
	historystore "nextopsx/store/history"
)

func TestCommandHistoryFacade(t *testing.T) {
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
	b := New(historystore.NewStore(data))
	code := 0
	if err := b.SaveCommandHistory("h1", "snippet", "df -h", &code, "ok"); err != nil {
		t.Fatalf("SaveCommandHistory: %v", err)
	}
	list, err := b.ListCommandHistory("h1", 10)
	if err != nil {
		t.Fatalf("ListCommandHistory: %v", err)
	}
	if len(list) != 1 || list[0].Command != "df -h" || list[0].Source != "snippet" {
		t.Fatalf("history = %+v", list)
	}
}

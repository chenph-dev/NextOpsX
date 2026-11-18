package historystore

import (
	"testing"

	"nextopsx/store"
)

func newTestStore(t *testing.T) *Store {
	t.Helper()
	t.Setenv("NEXTOPSX_DATA_DIR", t.TempDir())
	t.Setenv("NEXTOPSX_TEST_NO_KEYRING", "1")
	app, err := store.Open()
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() {
		if sqlDB, _ := app.GORM().DB(); sqlDB != nil {
			sqlDB.Close()
		}
	})
	return NewStore(app)
}

func TestHistoryStoreListAndSave(t *testing.T) {
	s := newTestStore(t)
	code := 0
	for _, e := range []Entry{
		{HostID: "h1", Source: "ai", Command: "ls", Output: "a", Ts: 100},
		{HostID: "h1", Source: "snippet", Command: "df -h", ExitCode: &code, Ts: 200},
		{HostID: "h2", Source: "ai", Command: "whoami", Ts: 300},
	} {
		if err := s.Save(e); err != nil {
			t.Fatalf("Save %q: %v", e.Command, err)
		}
	}

	list, err := s.List("h1", 10)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("list = %+v", list)
	}
	if list[0].Command != "df -h" || list[1].Command != "ls" {
		t.Fatalf("order = %+v", list)
	}
	if list[0].ID == "" || list[0].ExitCode == nil || *list[0].ExitCode != 0 {
		t.Fatalf("saved fields = %+v", list[0])
	}
}

func TestHistoryStoreAssignsIDAndTime(t *testing.T) {
	s := newTestStore(t)
	if err := s.Save(Entry{HostID: "h1", Source: "app", Command: "uptime"}); err != nil {
		t.Fatalf("Save: %v", err)
	}
	list, err := s.List("h1", 10)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(list) != 1 || list[0].ID == "" || list[0].Ts == 0 {
		t.Fatalf("entry = %+v", list)
	}
}

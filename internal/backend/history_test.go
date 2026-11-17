package backend

import "testing"

func TestCommandHistoryFacade(t *testing.T) {
	b := openTestBackend(t)
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

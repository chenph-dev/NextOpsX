package workspace

import (
	"strings"
	"testing"

	"nextopsx/i18n"
	historystore "nextopsx/store/history"
	hoststore "nextopsx/store/hosts"
	snippetsstore "nextopsx/store/snippets"
)

func TestResolveSnippetCommand(t *testing.T) {
	a := &State{Selected: &hoststore.TreeNode{
		Name: "web-01", Addr: "10.0.0.5", User: "ops", Port: 22, Protocol: "ssh",
	}}
	got := a.ResolveSnippetCommand("ssh {{user}}@{{host}} -p {{port}} # {{name}} {{protocol}} {{addr}}")
	want := "ssh ops@10.0.0.5 -p 22 # web-01 ssh 10.0.0.5"
	if got != want {
		t.Fatalf("resolved = %q, want %q", got, want)
	}
}

func TestSnippetActionsWithoutBackendOrTerminal(t *testing.T) {
	i18n.SetLanguage(i18n.ZhCN)
	t.Cleanup(func() { i18n.SetLanguage(i18n.ZhCN) })

	a := &State{}
	a.Snippets.Items = []snippetsstore.Snippet{{ID: "s1", Name: "ls", Command: "ls"}}
	a.InsertSnippet("s1")
	if a.Snippets.Err != i18n.T("snippets.err.notConnected") {
		t.Fatalf("insert error = %q", a.Snippets.Err)
	}
	a.Snippets.Err = ""
	a.RunSnippet("s1")
	if a.Snippets.Err != i18n.T("snippets.err.notConnected") {
		t.Fatalf("run error = %q", a.Snippets.Err)
	}

	a.LoadSnippets()
	if !strings.Contains(a.Snippets.Err, i18n.T("settings.err.backendNotInitialized")) {
		t.Fatalf("load error = %q", a.Snippets.Err)
	}

	a.Snippets.Draft = snippetsstore.Snippet{Name: " ", Command: ""}
	a.SaveSnippet()
	if a.Snippets.DraftErr != i18n.T("snippets.err.required") {
		t.Fatalf("save error = %q", a.Snippets.DraftErr)
	}
}

func TestToggleSnippetPicker(t *testing.T) {
	a := &State{}
	a.ToggleSnippetPicker()
	if !a.Snippets.PickerOpen {
		t.Fatal("picker should open")
	}
	if a.Snippets.Err == "" {
		t.Fatal("opening the picker without a backend should surface the load error")
	}
	a.ToggleSnippetPicker()
	if a.Snippets.PickerOpen {
		t.Fatal("picker should close")
	}
}

func TestLoadCommandHistory(t *testing.T) {
	i18n.SetLanguage(i18n.ZhCN)
	t.Cleanup(func() { i18n.SetLanguage(i18n.ZhCN) })

	a := &State{}
	a.Snippets.History = []historystore.Entry{{ID: "old", Command: "ls"}}
	a.LoadCommandHistory()
	if len(a.Snippets.History) != 0 || a.Snippets.HistoryErr != "" {
		t.Fatalf("history without a host = %+v", a.Snippets)
	}

	a.Selected = &hoststore.TreeNode{ID: "h1", NodeType: "host"}
	a.LoadCommandHistory()
	if !strings.Contains(a.Snippets.HistoryErr, i18n.T("settings.err.backendNotInitialized")) {
		t.Fatalf("history error = %q", a.Snippets.HistoryErr)
	}
}

func TestFavoriteHistoryWithoutBackend(t *testing.T) {
	i18n.SetLanguage(i18n.ZhCN)
	t.Cleanup(func() { i18n.SetLanguage(i18n.ZhCN) })

	a := &State{}
	a.Snippets.History = []historystore.Entry{{ID: "h1", Command: "df -h"}}
	a.FavoriteHistory("h1")
	if a.Snippets.Err != i18n.T("settings.err.backendNotInitialized") {
		t.Fatalf("favorite error = %q", a.Snippets.Err)
	}
}

func TestSnippetName(t *testing.T) {
	if got := snippetName("  df -h\nsecond line"); got != "df -h" {
		t.Fatalf("name = %q", got)
	}
	if got := snippetName(""); got != i18n.T("history.favoriteFallback") {
		t.Fatalf("empty name = %q", got)
	}
	long := strings.Repeat("x", 60)
	got := snippetName(long)
	if n := len([]rune(got)); n != 40 || !strings.HasSuffix(got, "…") {
		t.Fatalf("long name = %q (%d runes)", got, n)
	}
}

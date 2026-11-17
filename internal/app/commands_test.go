package app

import (
	"testing"

	historystore "nextopsx/internal/store/history"
	hoststore "nextopsx/internal/store/hosts"
	snippetsstore "nextopsx/internal/store/snippets"
)

func TestCommandItemsDedupAndMerge(t *testing.T) {
	a := &App{}
	code := 0
	a.Snippets.Items = []snippetsstore.Snippet{
		{ID: "s1", Name: "磁盘", Command: "df -h"},
		{ID: "s2", Name: "当前用户", Command: "whoami -- {{user}}"},
	}
	a.Snippets.History = []historystore.Entry{
		{ID: "h1", Command: "df -h", ExitCode: &code, Output: "old", Ts: 100},
		{ID: "h2", Command: "free -h", Ts: 200},
		{ID: "h3", Command: "df -h", Output: "newer", Ts: 300},
	}
	items := a.CommandItems()
	if len(items) != 3 {
		t.Fatalf("items = %+v", items)
	}
	var df, free, who *CommandItem
	for i := range items {
		switch items[i].Command {
		case "df -h":
			df = &items[i]
		case "free -h":
			free = &items[i]
		case "whoami -- {{user}}":
			who = &items[i]
		}
	}
	if df == nil || df.Source != "both" || df.Title != "磁盘" ||
		df.FavoriteID != "s1" || df.HistoryID != "h3" || df.Output != "newer" {
		t.Fatalf("merged df = %+v", df)
	}
	if free == nil || free.Source != "history" || free.HistoryID != "h2" {
		t.Fatalf("free = %+v", free)
	}
	if who == nil || who.Source != "favorite" || who.FavoriteID != "s2" {
		t.Fatalf("who = %+v", who)
	}
	if items[0].Command != "df -h" {
		t.Fatalf("favorites/both should sort first: %+v", items)
	}
}

func TestCommandItemText(t *testing.T) {
	a := &App{Selected: &hoststore.TreeNode{
		ID: "h1", NodeType: "host", User: "ops",
	}}
	if got := a.CommandItemText(CommandItem{Command: "whoami -- {{user}}", FavoriteID: "s1"}); got != "whoami -- ops" {
		t.Fatalf("favorite command = %q", got)
	}
	if got := a.CommandItemText(CommandItem{Command: "df -h", HistoryID: "h1"}); got != "df -h" {
		t.Fatalf("history command = %q", got)
	}
}

package app

import (
	"testing"

	"github.com/egoist/mygo/plugins/terminal"
	"github.com/egoist/mygo/ui"

	historystore "nextopsx/internal/store/history"
	hoststore "nextopsx/internal/store/hosts"
	snippetsstore "nextopsx/internal/store/snippets"
)

func completionTestApp() *App {
	a := &App{Term: &terminal.Terminal{}}
	a.Selected = &hoststore.TreeNode{ID: "h1", NodeType: "host", User: "ops"}
	a.Snippets.Items = []snippetsstore.Snippet{
		{ID: "s1", Name: "磁盘", Command: "df -h"},
	}
	a.Snippets.History = []historystore.Entry{
		{ID: "h1", Command: "df -h", Ts: 100},
		{ID: "h2", Command: "free -h", Ts: 200},
	}
	return a
}

func TestUpdateCompletionMergesAndFilters(t *testing.T) {
	a := completionTestApp()
	a.UpdateCompletion("d")
	if a.Completion.Open {
		t.Fatal("one-character prefix should not open the bubble")
	}
	a.UpdateCompletion("df")
	if !a.Completion.Open || len(a.Completion.Items) != 1 {
		t.Fatalf("completion = %+v", a.Completion)
	}
	if a.Completion.Items[0].Source != "both" {
		t.Fatalf("dedup source = %q", a.Completion.Items[0].Source)
	}
	a.UpdateCompletion("df -h")
	if !a.Completion.Open || len(a.Completion.Items) != 1 {
		t.Fatalf("completion = %+v", a.Completion)
	}

	a.UpdateCompletion("free")
	if !a.Completion.Open || a.Completion.Items[0].Command != "free -h" {
		t.Fatalf("completion = %+v", a.Completion)
	}
	a.UpdateCompletion("")
	if a.Completion.Open {
		t.Fatal("empty input should close the bubble")
	}
	a.UpdateCompletion("a|b")
	if a.Completion.Open {
		t.Fatal("shell metacharacters should close the bubble")
	}
}

func TestHandleTermKeyNavigation(t *testing.T) {
	a := &App{Completion: CompletionState{
		Open: true, Selected: 0,
		Items: []CommandItem{{Command: "a"}, {Command: "b"}},
	}}
	if !a.HandleTermKey(0, ui.KeyDown) || a.Completion.Selected != 1 {
		t.Fatalf("down: %+v", a.Completion)
	}
	if !a.HandleTermKey(0, ui.KeyUp) || a.Completion.Selected != 0 {
		t.Fatalf("up: %+v", a.Completion)
	}
	if a.HandleTermKey(0, ui.KeyEnter) {
		t.Fatal("Enter must pass through to the shell")
	}
	if !a.HandleTermKey(0, ui.KeyEscape) || a.Completion.Open {
		t.Fatalf("escape: %+v", a.Completion)
	}
}

func TestInputTrackerLineEditing(t *testing.T) {
	var line string
	tr := newInputTracker(nil, func(s string) { line = s })
	tr.observe([]byte("df -h"))
	if line != "df -h" {
		t.Fatalf("line = %q", line)
	}
	tr.observe([]byte{0x7f})
	if line != "df -" {
		t.Fatalf("backspace = %q", line)
	}
	tr.observe([]byte{0x1b, '[', 'A'})
	if line != "" {
		t.Fatalf("arrow should clear the line, got %q", line)
	}
	tr.observe([]byte("ls"))
	tr.observe([]byte{0x15})
	if line != "" {
		t.Fatalf("ctrl-u should clear the line, got %q", line)
	}
	tr.observe([]byte("pwd\r"))
	if line != "" {
		t.Fatalf("enter should clear the line, got %q", line)
	}
}

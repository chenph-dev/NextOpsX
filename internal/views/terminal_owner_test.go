package views

import (
	"testing"

	"github.com/egoist/mygo/plugins/terminal"
	"github.com/egoist/mygo/ui"

	"nextopsx/internal/app"
	hoststore "nextopsx/internal/store/hosts"
)

func TestShellHidesTerminalOfAnotherHost(t *testing.T) {
	term, err := terminal.New(terminal.Options{Command: []string{"/bin/sh"}})
	if err != nil {
		t.Fatalf("terminal.New: %v", err)
	}
	defer term.Close()

	a := &app.App{RightTab: 0, Term: term, TermHostID: "h1"}
	a.Roots = []hoststore.TreeNode{{
		ID: "h1", Name: "first", NodeType: "host", Protocol: "ssh",
		Addr: "10.0.0.5", Port: 22, User: "ops",
	}}
	a.Selected = &hoststore.TreeNode{
		ID: "h2", Name: "other", NodeType: "host", Protocol: "ssh",
		Addr: "10.0.0.6", Port: 22, User: "ops",
	}

	tt := ui.NewTester(Shell(a), 1000, 700)
	if !tt.HasText("关闭后台终端") || !tt.HasText("后台仍保持到「first」的终端会话。") {
		t.Fatalf("background terminal notice missing: %q", tt.Texts())
	}
	if !tt.HasText("连接") {
		t.Fatalf("selected host pane is not shown: %q", tt.Texts())
	}
}

package terminal_test

import (
	"testing"

	mygoterm "github.com/egoist/mygo/plugins/terminal"
	"github.com/egoist/mygo/ui"

	hoststore "nextopsx/store/hosts"
	"nextopsx/ui/views"
	"nextopsx/ui/workspace"
)

func TestShellHidesTerminalOfAnotherHost(t *testing.T) {
	term, err := mygoterm.New(mygoterm.Options{Command: []string{"/bin/sh"}})
	if err != nil {
		t.Fatalf("terminal.New: %v", err)
	}
	defer term.Close()

	a := &workspace.State{Term: term, TermHostID: "h1"}
	a.Roots = []hoststore.TreeNode{{
		ID: "h1", Name: "first", NodeType: "host", Protocol: "ssh",
		Addr: "10.0.0.5", Port: 22, User: "ops",
	}}
	a.Selected = &hoststore.TreeNode{
		ID: "h2", Name: "other", NodeType: "host", Protocol: "ssh",
		Addr: "10.0.0.6", Port: 22, User: "ops",
	}

	tt := ui.NewTester(views.Shell(a), 1000, 700)
	if !tt.HasText("关闭后台终端") || !tt.HasText("后台仍保持到「first」的终端会话。") {
		t.Fatalf("background terminal notice missing: %q", tt.Texts())
	}
	if !tt.HasText("打开终端") {
		t.Fatalf("selected host pane is not shown: %q", tt.Texts())
	}
}

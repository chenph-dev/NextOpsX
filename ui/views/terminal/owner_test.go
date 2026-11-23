package terminal_test

import (
	"testing"

	mygoterm "github.com/egoist/mygo/plugins/terminal"
	"github.com/egoist/mygo/ui"

	hoststore "nextopsx/store/hosts"
	"nextopsx/ui/views"
	"nextopsx/ui/workspace"
)

func TestSessionTabsSurviveAssetSwitch(t *testing.T) {
	term, err := mygoterm.New(mygoterm.Options{Command: []string{"/bin/sh"}})
	if err != nil {
		t.Fatalf("terminal.New: %v", err)
	}
	defer term.Close()

	a := &workspace.State{}
	a.Roots = []hoststore.TreeNode{
		{ID: "h1", Name: "first", NodeType: "host", Protocol: "ssh",
			Addr: "10.0.0.5", Port: 22, User: "ops"},
		{ID: "h2", Name: "other", NodeType: "host", Protocol: "ssh",
			Addr: "10.0.0.6", Port: 22, User: "ops"},
	}
	a.Sessions = []*workspace.TerminalSession{{
		HostID: "h1", Protocol: "ssh", Status: workspace.TerminalLive, Term: term,
	}}
	a.ActiveSessionID = "h1"
	a.Term = term
	a.TermHostID = "h1"
	a.Selected = &a.Roots[1]
	a.Tabs = []string{"h1"}

	tt := ui.NewTester(views.Shell(a), 1000, 700)
	if !tt.HasText("first") {
		t.Fatalf("session tab missing: %q", tt.Texts())
	}
	if !tt.HasText("打开终端") && !tt.HasText("双击资产打开终端，或右键选择「打开终端」。") {
		t.Fatalf("selected host welcome missing: %q", tt.Texts())
	}
	if len(a.Sessions) != 1 || a.ActiveSessionID != "h1" {
		t.Fatalf("switching assets closed the session: sessions=%d active=%q", len(a.Sessions), a.ActiveSessionID)
	}

	a.ActivateSession("h1")
	tt.Frame()
	if a.Selected == nil || a.Selected.ID != "h1" || a.ActiveSessionID != "h1" {
		t.Fatalf("activating tab: selected=%v active=%q", a.Selected, a.ActiveSessionID)
	}

	a.Sessions = append(a.Sessions, &workspace.TerminalSession{
		HostID: "h2", Protocol: "ssh", Status: workspace.TerminalConnecting,
	})
	a.Tabs = append(a.Tabs, "h2")
	tt.Frame()
	tt.Key(ui.Ctrl, ui.KeyTab)
	tt.Frame()
	if a.ActiveSessionID != "h2" || a.Selected == nil || a.Selected.ID != "h2" {
		t.Fatalf("Ctrl+Tab: selected=%v active=%q", a.Selected, a.ActiveSessionID)
	}
	tt.Key(ui.Ctrl|ui.Shift, ui.KeyTab)
	tt.Frame()
	if a.ActiveSessionID != "h1" {
		t.Fatalf("Ctrl+Shift+Tab: active=%q", a.ActiveSessionID)
	}
}

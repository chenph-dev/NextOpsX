package views

import (
	"strings"
	"testing"
	"time"

	"github.com/egoist/mygo/plugins/terminal"
	"github.com/egoist/mygo/ui"
)

// TestTerminalShell proves the Ghostty-backed terminal loads and runs a
// local shell in a headless view, before we rely on it for SSH.
func TestTerminalShell(t *testing.T) {
	term, err := terminal.New(terminal.Options{Command: []string{"/bin/sh"}})
	if err != nil {
		t.Fatalf("terminal.New: %v", err)
	}
	defer term.Close()

	tt := ui.NewTester(func(c *ui.Context) {
		terminal.View(c, term).Fill().AutoFocus()
	}, 640, 400)
	tt.Type("echo hello")
	tt.Key(0, ui.KeyEnter)

	deadline := time.Now().Add(5 * time.Second)
	for !strings.Contains(term.Text(), "\nhello") {
		if time.Now().After(deadline) {
			t.Fatalf("terminal output: %q", term.Text())
		}
		time.Sleep(10 * time.Millisecond)
		tt.Frame()
	}
}

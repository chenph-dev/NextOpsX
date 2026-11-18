package views

import (
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/egoist/mygo/plugins/terminal"
	"github.com/egoist/mygo/ui"
)

// ansiSeq strips color and cursor sequences so prompt characters do not hide
// the command output.
var ansiSeq = regexp.MustCompile(`\x1b\[[0-9;]*[A-Za-z]`)

// terminalShowsHello reports whether the terminal printed "hello" as a line
// of its own. The echoed input line ("echo hello") is ignored, and a shell
// prompt before the output ("$ hello", "user@host:~$ hello") is allowed.
func terminalShowsHello(text string) bool {
	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(ansiSeq.ReplaceAllString(line, ""))
		if line == "" || strings.HasPrefix(line, "echo ") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) > 0 && fields[len(fields)-1] == "hello" {
			return true
		}
	}
	return false
}

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
	for !terminalShowsHello(term.Text()) {
		if time.Now().After(deadline) {
			t.Fatalf("terminal output: %q", term.Text())
		}
		time.Sleep(10 * time.Millisecond)
		tt.Frame()
	}
}

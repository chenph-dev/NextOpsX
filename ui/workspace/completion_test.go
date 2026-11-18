package workspace

import (
	"io"
	"testing"

	"github.com/egoist/mygo/plugins/terminal"
	"github.com/egoist/mygo/ui"

	historystore "nextopsx/store/history"
	hoststore "nextopsx/store/hosts"
	snippetsstore "nextopsx/store/snippets"
)

func completionTestApp() *State {
	a := &State{Term: &terminal.Terminal{}}
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
	a := &State{Completion: CompletionState{
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
	var stale bool
	tr := newInputTracker(nil, func(s string, s2 bool) { line, stale = s, s2 })
	tr.observe([]byte("df -h"))
	if line != "df -h" || stale {
		t.Fatalf("line = %q stale = %v", line, stale)
	}
	tr.observe([]byte{0x7f})
	if line != "df -" || stale {
		t.Fatalf("backspace = %q stale = %v", line, stale)
	}
	tr.observe([]byte{0x1b, '[', 'A'})
	if line != "" || !stale {
		t.Fatalf("arrow should clear and mark the line, got %q stale = %v", line, stale)
	}
	tr.observe([]byte("ls"))
	if !stale {
		t.Fatalf("typing after an arrow must keep the line unknown, got %q", line)
	}
	tr.observe([]byte{0x15})
	if line != "" || stale {
		t.Fatalf("ctrl-u should reset the line, got %q stale = %v", line, stale)
	}
	tr.observe([]byte("pwd\r"))
	if line != "" || stale {
		t.Fatalf("enter should reset the line, got %q stale = %v", line, stale)
	}
}

func TestInputTrackerTabMarksLineUnknown(t *testing.T) {
	var line string
	var stale bool
	tr := newInputTracker(nil, func(s string, s2 bool) { line, stale = s, s2 })
	tr.observe([]byte("cd /var/lo\t"))
	if line != "" || !stale {
		t.Fatalf("tab completion = %q stale = %v", line, stale)
	}
	tr.observe([]byte("df"))
	if !stale {
		t.Fatalf("typing after tab completion must stay unknown: %q", line)
	}
	tr.observe([]byte("\r"))
	if stale {
		t.Fatal("enter should start a fresh line")
	}
	tr.observe([]byte("free"))
	if line != "free" || stale {
		t.Fatalf("fresh line = %q stale = %v", line, stale)
	}
}

func TestInputTrackerInteractiveState(t *testing.T) {
	tr := newInputTracker(nil, nil)
	tr.observeOutput([]byte("user@host:~$ "))
	if tr.Interactive() {
		t.Fatal("a shell prompt is not interactive")
	}
	tr.observeOutput([]byte("[sudo] password for root: "))
	if !tr.Interactive() {
		t.Fatal("password prompt should own the keyboard")
	}
	tr.observe([]byte("\r"))
	if tr.Interactive() {
		t.Fatal("enter should leave the password prompt")
	}
	tr.observeOutput([]byte("\x1b[?1049h"))
	if !tr.Interactive() {
		t.Fatal("alternate screen should own the keyboard")
	}
	tr.observeOutput([]byte("\x1b[?1049l"))
	if tr.Interactive() {
		t.Fatal("leaving the alternate screen should restore suggestions")
	}
}

func TestHandleTermKeyLeavesFullScreenAlone(t *testing.T) {
	tr := newInputTracker(nil, nil)
	a := &State{termTracker: tr}
	a.Completion = CompletionState{
		Open: true, Selected: 0,
		Items: []CommandItem{{Command: "df -h", HistoryID: "h1"}},
	}
	tr.observeOutput([]byte("\x1b[?1049h"))
	if a.HandleTermKey(0, ui.KeyDown) {
		t.Fatal("down arrow must go to the full-screen program")
	}
	if a.Completion.Open {
		t.Fatal("suggestions should close in a full-screen program")
	}

	a.Completion = CompletionState{
		Open: true, Selected: 0,
		Items: []CommandItem{{Command: "df -h", HistoryID: "h1"}},
	}
	tr.observeOutput([]byte("Password: "))
	if a.HandleTermKey(0, ui.KeyTab) {
		t.Fatal("tab must go to the password prompt")
	}
	if a.Completion.Open {
		t.Fatal("suggestions should close at a password prompt")
	}
}

type resizeTestConn struct {
	cols, rows int
}

func (c *resizeTestConn) Read([]byte) (int, error)    { return 0, io.EOF }
func (c *resizeTestConn) Write(p []byte) (int, error) { return len(p), nil }
func (c *resizeTestConn) Close() error                { return nil }
func (c *resizeTestConn) Resize(cols, rows int) error {
	c.cols, c.rows = cols, rows
	return nil
}

func TestInputTrackerForwardsResize(t *testing.T) {
	conn := &resizeTestConn{}
	tracker := newInputTracker(conn, nil)
	var _ interface{ Resize(cols, rows int) error } = tracker
	if err := tracker.Resize(120, 40); err != nil {
		t.Fatalf("Resize: %v", err)
	}
	if conn.cols != 120 || conn.rows != 40 {
		t.Fatalf("resize = %dx%d, want 120x40", conn.cols, conn.rows)
	}
}

func TestHandleTermKeyTogglesDrawers(t *testing.T) {
	a := &State{}
	a.Selected = &hoststore.TreeNode{ID: "h1", NodeType: "host", Protocol: "ssh"}
	if !a.HandleTermKey(ui.Ctrl|ui.Shift, ui.KeyJ) || !a.Layout.AIOpen {
		t.Fatalf("ops drawer did not open: %+v", a.Layout)
	}
	if !a.HandleTermKey(ui.Ctrl|ui.Shift, ui.KeyJ) || a.Layout.AIOpen {
		t.Fatalf("ops drawer did not close: %+v", a.Layout)
	}
	a.SFTP = SFTPState{HostID: "h1"}
	if !a.HandleTermKey(ui.Ctrl|ui.Shift, ui.KeyE) || !a.Layout.SFTPOpen {
		t.Fatalf("SFTP drawer did not open: %+v", a.Layout)
	}
}

func TestUpdateCompletionRespectsUnknownAndInteractiveLines(t *testing.T) {
	a := &State{Term: &terminal.Terminal{}}
	a.Selected = &hoststore.TreeNode{ID: "h1", NodeType: "host", User: "ops"}
	a.Snippets.History = []historystore.Entry{{ID: "h1", Command: "df -h", Ts: 10}}
	a.Completion = CompletionState{Open: true, Selected: 0, Buffer: "df"}
	a.UpdateCompletionLine("df", true)
	if a.Completion.Open {
		t.Fatal("an unknown shell line must not keep suggestions open")
	}

	tr := newInputTracker(nil, nil)
	tr.observeOutput([]byte("Password: "))
	a.termTracker = tr
	if !a.TerminalInteractive() {
		t.Fatal("password prompt should be interactive")
	}
	a.Completion = CompletionState{Open: true, Selected: 0, Buffer: "df"}
	a.UpdateCompletionLine("df", false)
	if a.Completion.Open {
		t.Fatal("suggestions must stay closed at a password prompt")
	}
}

func TestPasswordPromptDetection(t *testing.T) {
	if !passwordPrompt([]byte("\x1b[33mPassword:\x1b[0m ")) {
		t.Fatal("colored password prompt not detected")
	}
	if !passwordPrompt([]byte("Enter passphrase for key '/home/ops/.ssh/id_ed25519': ")) {
		t.Fatal("passphrase prompt not detected")
	}
	if passwordPrompt([]byte("user@host:~$ ")) {
		t.Fatal("a normal shell prompt was mistaken for a password prompt")
	}
}

type screenReadConn struct{ data []byte }

func (c *screenReadConn) Read(p []byte) (int, error) {
	if len(c.data) == 0 {
		return 0, io.EOF
	}
	n := copy(p, c.data)
	c.data = c.data[n:]
	return n, nil
}
func (c *screenReadConn) Write(p []byte) (int, error) { return len(p), nil }
func (c *screenReadConn) Close() error                { return nil }

func TestInputTrackerReadsScreenState(t *testing.T) {
	conn := &screenReadConn{data: []byte("Password: ")}
	tr := newInputTracker(conn, nil)
	buf := make([]byte, 64)
	if _, err := tr.Read(buf); err != nil {
		t.Fatalf("Read: %v", err)
	}
	if !tr.Interactive() {
		t.Fatal("output read through the tracker should mark the prompt")
	}
}

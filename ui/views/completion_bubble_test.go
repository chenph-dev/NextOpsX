package views

import (
	"bytes"
	"io"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/egoist/mygo/plugins/terminal"
	"github.com/egoist/mygo/ui"

	hoststore "nextopsx/store/hosts"
	"nextopsx/ui/app"
)

type completionTestConn struct {
	mu   sync.Mutex
	out  bytes.Buffer
	done chan struct{}
	once sync.Once
}

func newCompletionTestConn() *completionTestConn {
	return &completionTestConn{done: make(chan struct{})}
}

func (c *completionTestConn) Read([]byte) (int, error) {
	<-c.done
	return 0, io.EOF
}

func (c *completionTestConn) Write(p []byte) (int, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.out.Write(p)
}

func (c *completionTestConn) Close() error {
	c.once.Do(func() { close(c.done) })
	return nil
}

func (c *completionTestConn) String() string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.out.String()
}

func TestTerminalWithCompletionKeepsFocus(t *testing.T) {
	conn := newCompletionTestConn()
	a := &app.App{}
	term, err := terminal.New(terminal.Options{
		Conn: conn, OnKey: a.HandleTermKey,
	})
	if err != nil {
		t.Fatalf("terminal.New: %v", err)
	}
	defer term.Close()
	a.Term = term
	a.TermHostID = "h1"
	a.Selected = &hoststore.TreeNode{ID: "h1", NodeType: "host"}
	a.Completion = app.CompletionState{
		Open: true, Buffer: "free", Selected: 0,
		Items: []app.CommandItem{
			{Command: "free -h", HistoryID: "h1"},
			{Command: "df -h", HistoryID: "h2"},
		},
	}
	tt := ui.NewTester(func(c *ui.Context) {
		terminalWithCompletion(c, a, c.Theme())
	}, 800, 500)
	tt.Frame()
	tt.Frame() // the Focus requested on the first frame takes effect
	tt.Key(0, ui.KeyDown)
	if a.Completion.Selected != 1 {
		t.Fatalf("bubble stole focus from the terminal: %+v", a.Completion)
	}
}

func TestTerminalOnKeyDrivesCompletion(t *testing.T) {
	conn := newCompletionTestConn()
	a := &app.App{}
	term, err := terminal.New(terminal.Options{
		Conn: conn, OnKey: a.HandleTermKey,
	})
	if err != nil {
		t.Fatalf("terminal.New: %v", err)
	}
	defer term.Close()
	a.Term = term
	a.TermHostID = "h1"
	a.Selected = &hoststore.TreeNode{ID: "h1", NodeType: "host"}
	a.Completion = app.CompletionState{
		Open: true, Buffer: "free", Selected: 0,
		Items: []app.CommandItem{
			{Command: "free -h", HistoryID: "h1"},
			{Command: "df -h", HistoryID: "h2"},
		},
	}
	tt := ui.NewTester(func(c *ui.Context) {
		terminal.View(c, term).Fill().AutoFocus()
	}, 600, 400)
	tt.Frame()

	tt.Key(0, ui.KeyDown)
	if a.Completion.Selected != 1 {
		t.Fatalf("down did not reach OnKey: %+v", a.Completion)
	}
	tt.Key(0, ui.KeyUp)
	if a.Completion.Selected != 0 {
		t.Fatalf("up did not reach OnKey: %+v", a.Completion)
	}
	tt.Key(0, ui.KeyEscape)
	if a.Completion.Open {
		t.Fatalf("escape did not close: %+v", a.Completion)
	}

	a.Completion = app.CompletionState{
		Open: true, Buffer: "free", Selected: 0,
		Items: []app.CommandItem{{Command: "free -h", HistoryID: "h1"}},
	}
	tt.Key(0, ui.KeyTab)
	deadline := time.Now().Add(2 * time.Second)
	for !strings.Contains(conn.String(), "free -h") && time.Now().Before(deadline) {
		time.Sleep(5 * time.Millisecond)
	}
	if got := conn.String(); !strings.Contains(got, "free -h") {
		t.Fatalf("Tab did not insert the command: %q", got)
	}
	if a.Completion.Open {
		t.Fatal("Tab should close the bubble")
	}
}

func TestDrawerShortcutsDoNotDoubleToggle(t *testing.T) {
	conn := newCompletionTestConn()
	a := &app.App{}
	term, err := terminal.New(terminal.Options{
		Conn: conn, OnKey: a.HandleTermKey,
	})
	if err != nil {
		t.Fatalf("terminal.New: %v", err)
	}
	defer term.Close()
	a.Term = term
	a.TermHostID = "h1"
	a.Selected = &hoststore.TreeNode{ID: "h1", NodeType: "host", Protocol: "ssh"}
	a.SFTP = app.SFTPState{HostID: "h1"}
	a.Completion = app.CompletionState{
		Open: true, Buffer: "free", Selected: 0,
		Items: []app.CommandItem{{Command: "free -h", HistoryID: "h1"}},
	}
	tt := ui.NewTester(Shell(a), 1000, 700)
	tt.Frame()
	tt.Frame() // the terminal takes the keyboard

	tt.Key(ui.Ctrl|ui.Shift, ui.KeyJ)
	if !a.Layout.AIOpen {
		t.Fatal("terminal shortcut should open the ops drawer once")
	}
	tt.Key(ui.Ctrl|ui.Shift, ui.KeyE)
	if !a.Layout.SFTPOpen {
		t.Fatal("terminal shortcut should open the SFTP drawer once")
	}
}

func TestCompletionPastesMultiLine(t *testing.T) {
	conn := newCompletionTestConn()
	a := &app.App{}
	term, err := terminal.New(terminal.Options{
		Conn: conn, OnKey: a.HandleTermKey,
	})
	if err != nil {
		t.Fatalf("terminal.New: %v", err)
	}
	defer term.Close()
	term.Feed([]byte("\x1b[?2004h")) // the shell asked for bracketed paste
	a.Term = term
	a.TermHostID = "h1"
	a.Selected = &hoststore.TreeNode{ID: "h1", NodeType: "host", Protocol: "ssh"}
	a.Completion = app.CompletionState{
		Open: true, Buffer: "run", Selected: 0,
		Items: []app.CommandItem{{
			Command: "echo one\necho two", HistoryID: "h1",
		}},
	}
	tt := ui.NewTester(func(c *ui.Context) {
		terminal.View(c, term).Fill().AutoFocus()
	}, 600, 400)
	tt.Frame()
	tt.Key(0, ui.KeyTab)

	deadline := time.Now().Add(2 * time.Second)
	for !strings.Contains(conn.String(), "echo two") && time.Now().Before(deadline) {
		time.Sleep(5 * time.Millisecond)
	}
	got := conn.String()
	if !strings.Contains(got, "\x1b[200~echo one\necho two\x1b[201~") {
		t.Fatalf("command was not pasted as one block: %q", got)
	}
}

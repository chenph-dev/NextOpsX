package app

import (
	"io"
	"sync"
	"testing"
	"time"

	"github.com/egoist/mygo/plugins/terminal"

	hoststore "nextopsx/internal/store/hosts"
)

func TestOpenAssetSelects(t *testing.T) {
	a := &App{}
	ssh := &hoststore.TreeNode{ID: "h1", NodeType: "host", Protocol: "ssh"}
	a.OpenAsset(ssh)
	if a.Selected != ssh {
		t.Fatalf("selected = %+v", a.Selected)
	}

	db := &hoststore.TreeNode{ID: "h2", NodeType: "host", Protocol: "mysql"}
	a.OpenAsset(db)
	if a.Selected != db {
		t.Fatalf("selected = %+v", a.Selected)
	}

	win := &hoststore.TreeNode{ID: "h3", NodeType: "host", Protocol: "winrm"}
	a.OpenAsset(win)
	if a.Selected != win {
		t.Fatalf("selected = %+v", a.Selected)
	}
}

type liveConn struct {
	done chan struct{}
	once sync.Once
}

func newLiveConn() *liveConn { return &liveConn{done: make(chan struct{})} }

func (c *liveConn) Read([]byte) (int, error)    { <-c.done; return 0, io.EOF }
func (c *liveConn) Write(p []byte) (int, error) { return len(p), nil }
func (c *liveConn) Close() error {
	c.once.Do(func() { close(c.done) })
	return nil
}

func TestTerminalLiveAfterProcessEnds(t *testing.T) {
	conn := newLiveConn()
	term, err := terminal.New(terminal.Options{Conn: conn})
	if err != nil {
		t.Fatalf("terminal.New: %v", err)
	}
	a := &App{Term: term, TermHostID: "h1"}
	if !a.terminalLive("h1") {
		t.Fatal("a running terminal should be live")
	}
	term.Close()
	select {
	case <-term.Done():
	case <-time.After(2 * time.Second):
		t.Fatal("terminal did not finish")
	}
	if a.terminalLive("h1") {
		t.Fatal("a finished terminal must not be live")
	}
}

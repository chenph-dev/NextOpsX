package workspace

import (
	"io"
	"sync"
	"testing"
	"time"

	"github.com/egoist/mygo/plugins/terminal"

	hoststore "nextopsx/store/hosts"
)

func TestSelectDoesNotAttachWorkbench(t *testing.T) {
	cases := []struct {
		protocol string
		attached func(*State) bool
	}{
		{"mysql", func(a *State) bool { return a.DB.HostID != "" }},
		{"redis", func(a *State) bool { return a.Redis.HostID != "" }},
		{"elasticsearch", func(a *State) bool { return a.ES.HostID != "" }},
	}
	for _, tc := range cases {
		t.Run(tc.protocol, func(t *testing.T) {
			a := &State{}
			n := &hoststore.TreeNode{
				ID: "h1", Name: "svc", NodeType: "host", Protocol: tc.protocol,
				Addr: "10.0.0.9", Port: 3306, User: "root",
			}
			a.Select(n)
			if tc.attached(a) {
				t.Fatal("single click attached the workbench")
			}
			a.OpenAsset(n)
			if !tc.attached(a) {
				t.Fatal("double click did not attach the workbench")
			}
		})
	}
}

func TestSelectSyncsOpenWorkbench(t *testing.T) {
	a := &State{}
	a.Redis = RedisState{HostID: "r1", HostName: "old", DBLabel: "old · db0"}
	a.Select(&hoststore.TreeNode{
		ID: "r1", Name: "new", NodeType: "host", Protocol: "redis",
		Params: map[string]any{"db": "2"},
	})
	if a.Redis.DBLabel != "new · db2" || a.Redis.HostName != "new" {
		t.Fatalf("redis label = %q host = %q", a.Redis.DBLabel, a.Redis.HostName)
	}

	a.ES = ESState{HostID: "e1", Label: "old · 10.0.0.9:9200"}
	a.Select(&hoststore.TreeNode{
		ID: "e1", Name: "new-es", NodeType: "host", Protocol: "elasticsearch",
		Addr: "10.0.0.10", Port: 9200,
	})
	if a.ES.Label != "new-es · 10.0.0.10:9200" {
		t.Fatalf("es label = %q", a.ES.Label)
	}
}

func TestOpenAssetSelects(t *testing.T) {
	a := &State{}
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

func TestSessionLiveAfterProcessEnds(t *testing.T) {
	conn := newLiveConn()
	term, err := terminal.New(terminal.Options{Conn: conn})
	if err != nil {
		t.Fatalf("terminal.New: %v", err)
	}
	session := &TerminalSession{HostID: "h1", Protocol: "ssh", Status: TerminalLive, Term: term}
	a := &State{Sessions: []*TerminalSession{session}}
	if !a.SessionLive(session) {
		t.Fatal("a running terminal should be live")
	}
	term.Close()
	select {
	case <-term.Done():
	case <-time.After(2 * time.Second):
		t.Fatal("terminal did not finish")
	}
	if a.SessionLive(session) {
		t.Fatal("a finished terminal must not be live")
	}
	if session.Status != TerminalClosed {
		t.Fatalf("finished session status = %v", session.Status)
	}
}

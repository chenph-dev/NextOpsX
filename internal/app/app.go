// Package app holds the UI state and the actions the views call. It does
// not render anything; package views does.
package app

import (
	"io"

	"github.com/egoist/mygo/plugins/terminal"

	"nextopsx/internal/backend"
	hoststore "nextopsx/internal/store/hosts"
)

// Updater runs a function on the UI thread. *mygo.Window implements it.
type Updater interface {
	Update(fn func())
}

// App is the state a window shows.
type App struct {
	Backend *backend.Backend

	Roots    []hoststore.TreeNode
	Open     map[string]*bool
	Selected *hoststore.TreeNode

	Term   *terminal.Terminal
	Busy   bool
	ErrMsg string

	conn io.Closer
	up   Updater
}

// New loads the asset tree and creates the app state.
func New(b *backend.Backend) (*App, error) {
	roots, err := b.ListHosts()
	if err != nil {
		return nil, err
	}
	a := &App{Backend: b, Roots: roots, Open: map[string]*bool{}}
	a.initOpen(roots)
	return a, nil
}

func (a *App) initOpen(nodes []hoststore.TreeNode) {
	for i := range nodes {
		n := &nodes[i]
		if n.NodeType != "folder" {
			continue
		}
		open := true
		a.Open[n.ID] = &open
		a.initOpen(n.Children)
	}
}

// Attach gives the app a way to post state changes to the UI thread.
func (a *App) Attach(up Updater) {
	a.up = up
}

// Post runs fn on the UI thread. Without a window (tests) it runs inline.
func (a *App) Post(fn func()) {
	if a.up != nil {
		a.up.Update(fn)
		return
	}
	fn()
}

// Select marks an asset as selected.
func (a *App) Select(n *hoststore.TreeNode) {
	a.Selected = n
}

// Connect opens an SSH terminal for n and closes the previous session.
func (a *App) Connect(n *hoststore.TreeNode) {
	if a.Busy || a.Backend == nil {
		return
	}
	a.Busy = true
	a.ErrMsg = ""
	a.Disconnect()
	a.Selected = n

	go func() {
		conn, err := a.Backend.OpenTerminalConn(n.ID, 80, 24)
		if err != nil {
			a.Fail("连接失败: " + err.Error())
			return
		}
		a.Post(func() {
			term, err := terminal.New(terminal.Options{Conn: conn})
			if err != nil {
				_ = conn.Close()
				a.Busy = false
				a.ErrMsg = "创建终端失败: " + err.Error()
				return
			}
			a.Term = term
			a.conn = conn
			a.Busy = false
		})
	}()
}

// Fail records a connection error and clears the busy flag.
func (a *App) Fail(msg string) {
	a.Post(func() {
		a.Busy = false
		a.ErrMsg = msg
	})
}

// Disconnect closes the terminal and its SSH session.
func (a *App) Disconnect() {
	if a.Term != nil {
		a.Term.Close()
		a.Term = nil
	}
	if a.conn != nil {
		_ = a.conn.Close()
		a.conn = nil
	}
}

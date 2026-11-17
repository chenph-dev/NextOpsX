// Package app holds the UI state and the actions the views call. It does
// not render anything; package views does.
package app

import (
	"io"
	"sync"

	"github.com/egoist/mygo/plugins/terminal"

	"nextopsx/internal/backend"
	"nextopsx/internal/i18n"
	configstore "nextopsx/internal/store/config"
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

	// View is the window's top-level view: assets or settings.
	View     MainView
	Settings SettingsState

	// Layout holds the workbench panes around the main view.
	Layout   LayoutState
	AI       AIState
	Hosts    HostsState
	DB       DBState
	Redis    RedisState
	ES       ESState
	SFTP     SFTPState
	Skills   SkillsState
	Policy   PolicyState
	Audit    AuditState
	Snippets SnippetState

	Term        *terminal.Terminal
	TermHostID  string
	termTracker *inputTracker
	Completion  CompletionState
	Busy        bool
	ErrMsg      string

	conn   io.Closer
	up     Updater
	writes sync.WaitGroup           // 退出前等待短小的异步写（主题等）落库
	notify func(title, body string) // 桌面通知（由 main 注入；可空）

	// rebuildChrome rebuilds native chrome whose labels depend on the
	// language (menu bar, tray). main installs it; nil in tests.
	rebuildChrome func()
}

// LayoutState holds the workbench's auxiliary panes. The main view is chosen
// from the selected asset's protocol; the drawers keep their own state.
type LayoutState struct {
	// AIOpen shows the intelligent ops drawer at the bottom.
	AIOpen   bool
	AIHeight float32
	// SFTPOpen shows the file drawer on the right, for SSH assets only.
	SFTPOpen  bool
	SFTPWidth float32
}

// Default drawer sizes.
const (
	defaultAIHeight  float32 = 320
	defaultSFTPWidth float32 = 330
	minAIHeight      float32 = 200
	maxAIHeight      float32 = 480
	minSFTPWidth     float32 = 260
	maxSFTPWidth     float32 = 420
)

// AIDrawerHeight returns the drawer height within its limits.
func (a *App) AIDrawerHeight() float32 {
	return clampFloat(a.Layout.AIHeight, defaultAIHeight, minAIHeight, maxAIHeight)
}

// SFTPDrawerWidth returns the drawer width within its limits.
func (a *App) SFTPDrawerWidth() float32 {
	return clampFloat(a.Layout.SFTPWidth, defaultSFTPWidth, minSFTPWidth, maxSFTPWidth)
}

// SetAIDrawerHeight clamps and stores the intelligent ops drawer height.
func (a *App) SetAIDrawerHeight(v float32) {
	a.Layout.AIHeight = clampFloat(v, defaultAIHeight, minAIHeight, maxAIHeight)
}

// SetSFTPDrawerWidth clamps and stores the file drawer width.
func (a *App) SetSFTPDrawerWidth(v float32) {
	a.Layout.SFTPWidth = clampFloat(v, defaultSFTPWidth, minSFTPWidth, maxSFTPWidth)
}

// ToggleAIDrawer opens or collapses the intelligent ops drawer.
func (a *App) ToggleAIDrawer() {
	if a.Selected == nil || a.Selected.NodeType != "host" {
		return
	}
	a.Layout.AIOpen = !a.Layout.AIOpen
	if a.Layout.AIOpen && a.AI.HostID != a.Selected.ID {
		a.attachAI(a.Selected.ID)
	}
}

// ToggleSFTPDrawer opens or closes the file drawer for an SSH asset.
func (a *App) ToggleSFTPDrawer() {
	if a.Selected == nil || a.Selected.NodeType != "host" ||
		!a.IsSSHProtocol(a.Selected.Protocol) {
		return
	}
	a.Layout.SFTPOpen = !a.Layout.SFTPOpen
	if a.Layout.SFTPOpen {
		a.EnsureSFTP()
	}
}

func clampFloat(v, fallback, min, max float32) float32 {
	if v <= 0 {
		v = fallback
	}
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}

// New loads the asset tree and creates the app state.
func New(b *backend.Backend) (*App, error) {
	roots, err := b.ListHosts()
	if err != nil {
		return nil, err
	}
	a := &App{Backend: b, Roots: roots, Open: map[string]*bool{}}
	a.initOpen(roots)
	a.Audit.Selected = -1
	a.Settings.DataDir = b.DataDir()
	if ui, err := b.GetUISettings(); err == nil {
		a.Settings.Theme = ui.Theme
		a.Settings.Scale = ui.Scale
		a.Settings.Language = ui.Language
	} else {
		a.Settings.UIErr = i18n.Tf("app.readUISettingsFailed", err.Error())
	}
	if a.Settings.Theme == "" {
		a.Settings.Theme = "system"
	}
	if a.Settings.Language == "" {
		a.Settings.Language = string(i18n.Default)
	}
	a.Settings.Scale = configstore.NormalizeScale(a.Settings.Scale)
	a.Settings.ThemeIndex = ThemeIndex(a.Settings.Theme)
	a.Settings.ScaleIndex = ScaleIndex(a.Settings.Scale)
	a.Settings.Language = configstore.NormalizeLanguage(a.Settings.Language)
	a.Settings.LanguageIndex = i18n.LangIndex(i18n.Lang(a.Settings.Language))
	i18n.SetLanguage(i18n.Lang(a.Settings.Language))
	a.ensureHostConnectors()
	b.SetAIEventSink(a.HandleAIEvent)
	b.SetSFTPEventSink(a.HandleSFTPEvent)
	return a, nil
}

func (a *App) initOpen(nodes []hoststore.TreeNode) {
	for i := range nodes {
		n := &nodes[i]
		if n.NodeType != "folder" {
			continue
		}
		if _, ok := a.Open[n.ID]; !ok {
			open := true
			a.Open[n.ID] = &open
		}
		a.initOpen(n.Children)
	}
}

// Attach gives the app a way to post state changes to the UI thread.
func (a *App) Attach(up Updater) {
	a.up = up
}

// SetNotifier installs the desktop-notification sink; nil disables
// notifications.
func (a *App) SetNotifier(fn func(title, body string)) {
	a.notify = fn
}

// SetChromeRebuilder installs the callback that rebuilds language-dependent
// native chrome (menu bar and tray). main sets it once the window exists.
func (a *App) SetChromeRebuilder(fn func()) {
	a.rebuildChrome = fn
}

func (a *App) notifyUser(title, body string) {
	if a.notify != nil {
		a.notify(title, body)
	}
}

// WaitForWrites blocks until short background writes (theme preferences)
// have reached the database. Call it before closing the backend on quit.
func (a *App) WaitForWrites() { a.writes.Wait() }

// Post runs fn on the UI thread. Without a window (tests) it runs inline.
func (a *App) Post(fn func()) {
	if a.up != nil {
		a.up.Update(fn)
		return
	}
	fn()
}

// Select marks an asset as selected and attaches its workbench and AI
// conversation. Passing nil clears the selection. Either way it leaves the
// settings page, so every selection path goes through here.
func (a *App) Select(n *hoststore.TreeNode) {
	a.View = ViewAssets
	a.Selected = n
	a.ErrMsg = ""
	a.Snippets.PickerOpen = false
	a.CloseCompletion()
	a.attachAsset(n)
}

// attachAsset initializes per-asset state for a selected asset. It is used
// by Select for tree clicks and by tree refreshes (create/edit) after the
// selection is re-pointed.
func (a *App) attachAsset(n *hoststore.TreeNode) {
	if n == nil || n.NodeType != "host" {
		return
	}
	if a.IsRedisProtocol(n.Protocol) {
		a.Layout.SFTPOpen = false
		a.AttachRedis(n)
	} else if a.IsESProtocol(n.Protocol) {
		a.Layout.SFTPOpen = false
		a.AttachES(n)
	} else if a.IsDBProtocol(n.Protocol) {
		a.Layout.SFTPOpen = false
		a.attachDB(n)
	} else if !a.IsSSHProtocol(n.Protocol) {
		a.Layout.SFTPOpen = false
	} else if a.Layout.SFTPOpen {
		a.AttachSFTPFor(n)
	}
	if n.ID != a.AI.HostID {
		a.attachAI(n.ID)
	}
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
			a.Fail(i18n.Tf("app.connectFailed", err.Error()))
			return
		}
		a.Post(func() {
			tracker := newInputTracker(conn, func(line string, stale bool) {
				a.Post(func() { a.UpdateCompletionLine(line, stale) })
			})
			term, err := terminal.New(terminal.Options{
				Conn: tracker, OnKey: a.HandleTermKey,
			})
			if err != nil {
				_ = tracker.Close()
				a.Busy = false
				a.ErrMsg = i18n.Tf("app.terminalFailed", err.Error())
				return
			}
			a.Term = term
			a.TermHostID = n.ID
			a.conn = tracker
			a.termTracker = tracker
			a.Busy = false
			a.CloseCompletion()
			a.LoadCommandHistory()
			if len(a.Snippets.Items) == 0 {
				a.LoadSnippets()
			}
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
	a.TermHostID = ""
	a.termTracker = nil
	a.Layout.SFTPOpen = false
	a.Snippets.PickerOpen = false
	a.CloseCompletion()
}

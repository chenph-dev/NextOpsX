// Package workspace holds one window's UI state and the actions the views call.
// It is the ViewModel between ui/views and service: views read this state
// and invoke its methods, and the methods call service/*, run blocking work
// off the UI thread, then post results back here.
//
// It is not the application entry point (cmd/nextopsx is), it renders
// nothing (ui/views does), and it never talks to the database or a remote
// host directly (service/* does).
//
// It is intentionally one package. The feature files (ai.go, db.go,
// sftp.go, settings.go, …) share one State and its selected asset, layout
// and terminal, and their actions call each other. Splitting them into
// subpackages would only turn the store into a web of callbacks.
package workspace

import (
	"sync"
	"time"

	"github.com/egoist/mygo/plugins/terminal"
	"github.com/egoist/mygo/ui"

	"nextopsx/i18n"
	"nextopsx/service/backend"
	configstore "nextopsx/store/config"
	hoststore "nextopsx/store/hosts"
)

// Updater runs a function on the UI thread. *mygo.Window implements it.
type Updater interface {
	Update(fn func())
}

// State is the workbench state one window shows.
type State struct {
	Backend *backend.Backend

	Roots       []hoststore.TreeNode
	Open        map[string]*bool
	Selected    *hoststore.TreeNode
	lastAssetID string

	// View is the window's top-level view: assets or settings.
	View     MainView
	Settings SettingsState

	// Layout holds the workbench panes around the main view.
	Layout LayoutState
	AI     AIState
	Hosts  HostsState
	DB     DBState
	Redis  RedisState
	ES     ESState
	WinRM  WinRMState
	SFTP   SFTPState

	// Tabs lists open asset IDs in tab order. Inactive workbench states
	// are kept per asset so switching tabs preserves their open tabs and
	// loaded data.
	Tabs        []string
	TabScroll   ui.ScrollState
	tabScrolled string
	dbStates    map[string]DBState
	redisStates map[string]RedisState
	esStates    map[string]ESState
	winrmStates map[string]WinRMState
	Skills      SkillsState
	Policy      PolicyState
	Audit       AuditState
	Snippets    SnippetState

	Sessions        []*TerminalSession
	ActiveSessionID string

	Term        *terminal.Terminal
	TermHostID  string
	termTracker *inputTracker
	Completion  CompletionState
	ErrMsg      string

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
	defaultAIHeight  float32 = configstore.DefaultAIHeight
	defaultSFTPWidth float32 = configstore.DefaultSFTPWidth
	minAIHeight      float32 = configstore.MinAIHeight
	maxAIHeight      float32 = configstore.MaxAIHeight
	minSFTPWidth     float32 = configstore.MinSFTPWidth
	maxSFTPWidth     float32 = configstore.MaxSFTPWidth
)

// AIDrawerHeight returns the drawer height within its limits.
func (a *State) AIDrawerHeight() float32 {
	return clampFloat(a.Layout.AIHeight, defaultAIHeight, minAIHeight, maxAIHeight)
}

// SFTPDrawerWidth returns the drawer width within its limits.
func (a *State) SFTPDrawerWidth() float32 {
	return clampFloat(a.Layout.SFTPWidth, defaultSFTPWidth, minSFTPWidth, maxSFTPWidth)
}

// SetAIDrawerHeight clamps and stores the intelligent ops drawer height.
func (a *State) SetAIDrawerHeight(v float32) {
	a.Layout.AIHeight = clampFloat(v, defaultAIHeight, minAIHeight, maxAIHeight)
}

// SetSFTPDrawerWidth clamps and stores the file drawer width.
func (a *State) SetSFTPDrawerWidth(v float32) {
	a.Layout.SFTPWidth = clampFloat(v, defaultSFTPWidth, minSFTPWidth, maxSFTPWidth)
}

// SaveLayout persists the drawer state after a drag.
func (a *State) SaveLayout() {
	a.saveUIPrefs()
}

// ToggleAIDrawer opens or collapses the intelligent ops drawer.
func (a *State) ToggleAIDrawer() {
	if a.Layout.AIOpen {
		a.CloseAIDrawer()
		return
	}
	a.OpenAIDrawer()
}

// OpenAIDrawer shows the intelligent ops drawer for the selected host.
func (a *State) OpenAIDrawer() {
	if a.Selected == nil || a.Selected.NodeType != "host" {
		return
	}
	a.Layout.AIOpen = true
	if a.AI.HostID != a.Selected.ID {
		a.attachAI(a.Selected.ID)
	}
	a.saveUIPrefs()
}

// CloseAIDrawer collapses the intelligent ops drawer.
func (a *State) CloseAIDrawer() {
	if !a.Layout.AIOpen {
		return
	}
	a.Layout.AIOpen = false
	a.saveUIPrefs()
}

// ToggleSFTPDrawer opens or closes the file drawer for an SSH asset.
func (a *State) ToggleSFTPDrawer() {
	if a.Layout.SFTPOpen {
		a.CloseSFTPDrawer()
		return
	}
	a.OpenSFTPDrawer()
}

// OpenSFTPDrawer shows the file drawer for the selected SSH host.
func (a *State) OpenSFTPDrawer() {
	if a.Selected == nil || a.Selected.NodeType != "host" ||
		!a.IsSSHProtocol(a.Selected.Protocol) {
		return
	}
	a.Layout.SFTPOpen = true
	a.EnsureSFTP()
	a.saveUIPrefs()
}

// CloseSFTPDrawer hides the file drawer; an active transfer keeps running.
func (a *State) CloseSFTPDrawer() {
	if !a.Layout.SFTPOpen {
		return
	}
	a.Layout.SFTPOpen = false
	a.saveUIPrefs()
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
func New(b *backend.Backend) (*State, error) {
	roots, err := b.Hosts.ListHosts()
	if err != nil {
		return nil, err
	}
	a := &State{Backend: b, Roots: roots, Open: map[string]*bool{}}
	a.initOpen(roots)
	a.Audit.Selected = -1
	a.Settings.DataDir = b.Settings.DataDir()
	if ui, err := b.Settings.GetUISettings(); err == nil {
		a.Settings.Theme = ui.Theme
		a.Settings.Scale = ui.Scale
		a.Settings.Language = ui.Language
		a.Layout.AIOpen = ui.AIOpen
		a.Layout.SFTPOpen = ui.SFTPOpen
		a.SetAIDrawerHeight(float32(ui.AIHeight))
		a.SetSFTPDrawerWidth(float32(ui.SFTPWidth))
		a.lastAssetID = ui.LastAssetID
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

// RestoreWorkspace selects the asset from the previous session. Call it after
// Attach: selecting an asset starts background work whose Post must land on
// the UI thread.
func (a *State) RestoreWorkspace() {
	if a.Selected != nil || a.lastAssetID == "" {
		return
	}
	if n := findTreeNode(a.Roots, a.lastAssetID); n != nil {
		a.Select(n)
	} else {
		a.lastAssetID = ""
		a.saveUIPrefs()
	}
}

func (a *State) initOpen(nodes []hoststore.TreeNode) {
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
func (a *State) Attach(up Updater) {
	a.up = up
}

// SetNotifier installs the desktop-notification sink; nil disables
// notifications.
func (a *State) SetNotifier(fn func(title, body string)) {
	a.notify = fn
}

// SetChromeRebuilder installs the callback that rebuilds language-dependent
// native chrome (menu bar and tray). main sets it once the window exists.
func (a *State) SetChromeRebuilder(fn func()) {
	a.rebuildChrome = fn
}

func (a *State) notifyUser(title, body string) {
	if a.notify != nil {
		a.notify(title, body)
	}
}

// WaitForWrites blocks until short background writes (theme preferences)
// have reached the database. Call it before closing the backend on quit.
func (a *State) WaitForWrites() { a.writes.Wait() }

// Post runs fn on the UI thread. Without a window (tests) it runs inline.
func (a *State) Post(fn func()) {
	if a.up != nil {
		a.up.Update(fn)
		return
	}
	fn()
}

// Select marks an asset as selected and attaches its workbench and AI
// conversation. Passing nil clears the selection. Either way it leaves the
// settings page, so every selection path goes through here.
func (a *State) Select(n *hoststore.TreeNode) {
	a.View = ViewAssets
	a.Selected = n
	a.ErrMsg = ""
	a.Snippets.PickerOpen = false
	a.CloseCompletion()
	a.attachAsset(n)
	if n != nil {
		if a.IsSSHProtocol(n.Protocol) {
			a.activateExistingSession(n.ID)
		}
		a.rememberAsset(n.ID)
	}
}

// rememberAsset persists the last non-nil selection for the next launch.
func (a *State) rememberAsset(id string) {
	if id == "" || id == a.lastAssetID {
		return
	}
	a.lastAssetID = id
	a.saveUIPrefs()
}

// OpenAsset selects n and opens its main view. SSH hosts connect (or
// reconnect when the previous process exited); WinRM opens the remote
// desktop; Redis, Elasticsearch and database assets attach their workbench.
func (a *State) OpenAsset(n *hoststore.TreeNode) {
	if n != nil && n.NodeType == "host" && !a.IsSSHProtocol(n.Protocol) {
		a.openAssetTab(n.ID)
	}
	a.Select(n)
	if n == nil || n.NodeType != "host" {
		return
	}
	switch {
	case a.IsSSHProtocol(n.Protocol):
		s := a.SessionFor(n.ID)
		if s == nil {
			a.Connect(n)
			return
		}
		if s.Status == TerminalConnecting || a.SessionLive(s) {
			return // Select already focused this session.
		}
		a.ActivateSession(n.ID)
	case a.IsWinRMProtocol(n.Protocol):
		a.attachWinRM(n)
	case a.IsRedisProtocol(n.Protocol):
		a.AttachRedis(n)
	case a.IsESProtocol(n.Protocol):
		a.AttachES(n)
	case a.IsSQLProtocol(n.Protocol):
		a.attachDB(n)
	}
}

// attachAsset initializes per-asset state for a selected asset. It is used
// by Select for tree clicks and by tree refreshes (create/edit) after the
// selection is re-pointed.
func (a *State) attachAsset(n *hoststore.TreeNode) {
	if n == nil || n.NodeType != "host" {
		return
	}
	a.saveActiveWorkbenches()
	// A single click only selects a non-SSH asset: workbench protocols
	// attach on double-click (OpenAsset), like an SSH host connects.
	if !a.IsSSHProtocol(n.Protocol) {
		a.Layout.SFTPOpen = false
	} else if a.Layout.SFTPOpen {
		a.AttachSFTPFor(n)
	}
	// An open tab shows its preserved workbench; an unopened asset stays a
	// selection panel until OpenAsset attaches it.
	if a.IsAssetTabOpen(n.ID) {
		switch {
		case a.IsRedisProtocol(n.Protocol):
			a.AttachRedis(n)
		case a.IsESProtocol(n.Protocol):
			a.AttachES(n)
		case a.IsWinRMProtocol(n.Protocol):
			a.attachWinRM(n)
		case a.IsSQLProtocol(n.Protocol):
			a.attachDB(n)
		}
	}
	if n.ID != a.AI.HostID {
		a.attachAI(n.ID)
	}
}

// Connect opens a terminal tab for n, or focuses its existing tab.
func (a *State) Connect(n *hoststore.TreeNode) {
	if n == nil || a.Backend == nil {
		return
	}
	if s := a.SessionFor(n.ID); s != nil {
		if s.Status == TerminalConnecting || a.SessionLive(s) {
			a.ActivateSession(n.ID)
			return
		}
		a.closeSessionTerm(s)
		a.removeSession(s)
	}
	if len(a.Sessions) >= MaxTerminalSessions {
		for _, s := range a.Sessions {
			if s.Status == TerminalClosed {
				a.closeSessionTerm(s)
				a.removeSession(s)
				break
			}
		}
	}
	if len(a.Sessions) >= MaxTerminalSessions {
		a.ErrMsg = i18n.Tf("terminal.limit", MaxTerminalSessions)
		return
	}
	a.openAssetTab(n.ID)

	session := &TerminalSession{HostID: n.ID, Protocol: n.Protocol, Status: TerminalConnecting, LastUsed: time.Now()}
	a.Sessions = append(a.Sessions, session)
	a.Selected = n
	a.View = ViewAssets
	a.ErrMsg = ""
	a.rememberAsset(n.ID)
	a.bindActiveSession(session)

	go func() {
		conn, err := a.Backend.Terminal.OpenTerminalConn(n.ID, 80, 24)
		if err != nil {
			a.Post(func() {
				if !a.hasSession(session) {
					return
				}
				session.Status = TerminalClosed
				session.Err = err.Error()
				if a.ActiveSessionID == session.HostID {
					a.ErrMsg = i18n.Tf("app.connectFailed", err.Error())
				}
			})
			return
		}
		a.Post(func() {
			if !a.hasSession(session) {
				_ = conn.Close()
				return
			}
			tracker := newInputTracker(conn, func(line string, stale bool) {
				a.Post(func() {
					if a.ActiveSessionID == session.HostID {
						a.UpdateCompletionLine(line, stale)
					}
				})
			})
			term, err := terminal.New(terminal.Options{
				Conn: tracker, OnKey: a.HandleTermKey,
			})
			if err != nil {
				_ = tracker.Close()
				session.Status = TerminalClosed
				session.Err = err.Error()
				if a.ActiveSessionID == session.HostID {
					a.ErrMsg = i18n.Tf("app.terminalFailed", err.Error())
				}
				return
			}
			session.Term = term
			session.conn = tracker
			session.tracker = tracker
			session.Status = TerminalLive
			session.Err = ""
			if a.ActiveSessionID == session.HostID {
				a.bindActiveSession(session)
				a.ErrMsg = ""
				if a.Selected != nil && a.Selected.ID == session.HostID {
					a.LoadCommandHistory()
				}
			}
			if len(a.Snippets.Items) == 0 {
				a.LoadSnippets()
			}
		})
	}()
}

// Disconnect closes every terminal tab, for app exit.
func (a *State) Disconnect() {
	a.CloseAllSessions()
}

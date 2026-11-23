package workspace

import (
	"io"
	"time"

	"github.com/egoist/mygo/plugins/terminal"

	"nextopsx/i18n"
)

// MaxTerminalSessions caps concurrently open terminal tabs.
const MaxTerminalSessions = 8

// TerminalStatus is the lifecycle state of one terminal tab.
type TerminalStatus uint8

const (
	TerminalConnecting TerminalStatus = iota
	TerminalLive
	TerminalClosed
)

// TerminalSession is one SSH terminal kept alive independently of the
// selected asset. Switching assets or tabs does not close it.
type TerminalSession struct {
	HostID   string
	Protocol string
	Status   TerminalStatus
	Err      string
	Term     *terminal.Terminal
	conn     io.Closer
	tracker  *inputTracker
	LastUsed time.Time
}

// SessionFor returns the session for hostID, or nil.
func (a *State) SessionFor(hostID string) *TerminalSession {
	for _, s := range a.Sessions {
		if s.HostID == hostID {
			return s
		}
	}
	return nil
}

// SessionLive reports whether s has a running terminal, marking a process
// that exited as closed.
func (a *State) SessionLive(s *TerminalSession) bool {
	if s == nil || s.Status != TerminalLive || s.Term == nil {
		return false
	}
	select {
	case <-s.Term.Done():
		s.Status = TerminalClosed
		s.Err = ""
		return false
	default:
		return true
	}
}

// HostLabel names a host in the session bar.
func (a *State) HostLabel(id string) string {
	if id == "" {
		return ""
	}
	if n := findTreeNode(a.Roots, id); n != nil {
		return n.Name
	}
	return i18n.T("hosts.unknown")
}

// bindActiveSession makes s the terminal that views and completion use.
func (a *State) bindActiveSession(s *TerminalSession) {
	a.ActiveSessionID = s.HostID
	a.TermHostID = s.HostID
	a.Term = s.Term
	a.termTracker = s.tracker
	a.Snippets.PickerOpen = false
	a.CloseCompletion()
}

// clearActiveSession drops the active terminal references.
func (a *State) clearActiveSession() {
	a.ActiveSessionID = ""
	a.TermHostID = ""
	a.Term = nil
	a.termTracker = nil
	a.Snippets.PickerOpen = false
	a.CloseCompletion()
}

// ActivateSession shows the terminal of hostID, reconnecting a closed one.
func (a *State) ActivateSession(hostID string) {
	s := a.SessionFor(hostID)
	if s == nil {
		return
	}
	a.openAssetTab(hostID)
	if s.Status != TerminalConnecting && !a.SessionLive(s) {
		n := findTreeNode(a.Roots, hostID)
		if n == nil {
			a.CloseSession(hostID)
			return
		}
		a.View = ViewAssets
		a.Selected = n
		a.ErrMsg = ""
		a.attachAsset(n)
		a.rememberAsset(hostID)
		a.reconnectSession(s)
		return
	}
	n := findTreeNode(a.Roots, hostID)
	if n == nil {
		a.CloseSession(hostID)
		return
	}
	a.View = ViewAssets
	a.Selected = n
	a.ErrMsg = ""
	a.Snippets.PickerOpen = false
	a.attachAsset(n)
	a.rememberAsset(hostID)
	a.bindActiveSession(s)
	s.LastUsed = time.Now()
	if s.Term != nil {
		a.LoadCommandHistory()
	}
}

// activateExistingSession binds the session of a newly selected SSH host
// without opening or reconnecting anything.
func (a *State) activateExistingSession(hostID string) {
	s := a.SessionFor(hostID)
	if s == nil {
		return
	}
	if s.Status == TerminalConnecting || a.SessionLive(s) {
		a.openAssetTab(hostID)
		a.bindActiveSession(s)
		s.LastUsed = time.Now()
		a.LoadCommandHistory()
	}
}

// NextSession activates the next terminal tab.
func (a *State) NextSession() { a.cycleSession(1) }

// PrevSession activates the previous terminal tab.
func (a *State) PrevSession() { a.cycleSession(-1) }

func (a *State) cycleSession(delta int) {
	if len(a.Sessions) == 0 {
		return
	}
	index := -1
	for i, s := range a.Sessions {
		if s.HostID == a.ActiveSessionID {
			index = i
			break
		}
	}
	if index < 0 {
		if s := a.nextOpenSession(0); s != nil {
			a.ActivateSession(s.HostID)
		}
		return
	}
	for step := 1; step <= len(a.Sessions); step++ {
		j := index + delta*step
		for j < 0 {
			j += len(a.Sessions)
		}
		s := a.Sessions[j%len(a.Sessions)]
		if s.Status == TerminalConnecting || a.SessionLive(s) {
			a.ActivateSession(s.HostID)
			return
		}
	}
}

// CloseSession closes one terminal tab and activates the next open one when
// it was active.
func (a *State) CloseSession(hostID string) {
	s := a.SessionFor(hostID)
	if s == nil {
		return
	}
	index := 0
	for i, x := range a.Sessions {
		if x == s {
			index = i
			break
		}
	}
	active := a.ActiveSessionID == hostID
	// Closing a background tab while viewing another asset should not pull
	// the user away from that asset; only follow the next tab when the
	// closed terminal was the selected host.
	sameAsset := a.Selected != nil && a.Selected.ID == hostID
	nextTab := a.tabAfter(hostID)
	a.closeSessionTerm(s)
	a.removeSession(s)
	a.removeAssetTab(hostID)
	a.closeWorkbenchState(hostID)
	if !active {
		return
	}
	a.clearActiveSession()
	a.Layout.SFTPOpen = false
	if sameAsset {
		if nextTab != "" {
			a.ActivateAssetTab(nextTab)
			return
		}
		if next := a.nextOpenSession(index); next != nil {
			a.ActivateSession(next.HostID)
		}
	}
}

// CloseSessionsUnder closes every session whose host is id or inside it.
func (a *State) CloseSessionsUnder(id string) {
	if id == "" {
		return
	}
	// Close open asset tabs under the node first: that closes their
	// sessions and drops their workbench state.
	for _, tabID := range append([]string(nil), a.Tabs...) {
		if tabID == id || treeContains(a.Roots, id, tabID) {
			a.CloseAssetTab(tabID)
		}
	}
	// Close any session that was not tied to a tab.
	kept := a.Sessions[:0]
	for _, s := range a.Sessions {
		if s.HostID == id || treeContains(a.Roots, id, s.HostID) {
			a.closeSessionTerm(s)
			continue
		}
		kept = append(kept, s)
	}
	a.Sessions = kept
	if a.ActiveSessionID != "" && a.SessionFor(a.ActiveSessionID) == nil {
		a.clearActiveSession()
		a.Layout.SFTPOpen = false
		if next := a.nextOpenSession(0); next != nil {
			a.ActivateSession(next.HostID)
		}
	}
}

// CloseAllSessions closes every terminal tab, for app exit.
func (a *State) CloseAllSessions() {
	for _, s := range a.Sessions {
		a.closeSessionTerm(s)
	}
	a.Sessions = nil
	a.Tabs = nil
	a.dbStates = nil
	a.redisStates = nil
	a.esStates = nil
	a.clearActiveSession()
	a.Layout.SFTPOpen = false
	a.Snippets.PickerOpen = false
}

func (a *State) reconnectSession(s *TerminalSession) {
	hostID := s.HostID
	if a.ActiveSessionID == hostID {
		a.clearActiveSession()
	}
	a.closeSessionTerm(s)
	a.removeSession(s)
	if n := findTreeNode(a.Roots, hostID); n != nil {
		a.Connect(n)
	}
}

func (a *State) closeSessionTerm(s *TerminalSession) {
	if s == nil {
		return
	}
	if s.Term != nil {
		s.Term.Close()
		s.Term = nil
	}
	if s.conn != nil {
		_ = s.conn.Close()
		s.conn = nil
	}
	s.tracker = nil
	s.Status = TerminalClosed
}

func (a *State) removeSession(s *TerminalSession) bool {
	for i, x := range a.Sessions {
		if x == s {
			a.Sessions = append(a.Sessions[:i], a.Sessions[i+1:]...)
			return true
		}
	}
	return false
}

func (a *State) hasSession(s *TerminalSession) bool {
	for _, x := range a.Sessions {
		if x == s {
			return true
		}
	}
	return false
}

// nextOpenSession returns the first connecting or live session at or after
// start, wrapping around.
func (a *State) nextOpenSession(start int) *TerminalSession {
	if len(a.Sessions) == 0 {
		return nil
	}
	if start < 0 {
		start = 0
	}
	for i := 0; i < len(a.Sessions); i++ {
		s := a.Sessions[(start+i)%len(a.Sessions)]
		if s.Status == TerminalConnecting || a.SessionLive(s) {
			return s
		}
	}
	return nil
}

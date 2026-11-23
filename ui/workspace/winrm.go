package workspace

import (
	"context"
	"math"
	"strings"

	"github.com/egoist/mygo/ui"

	"nextopsx/i18n"
	hoststore "nextopsx/store/hosts"
)

// winRMMaxLines caps the console scrollback.
const winRMMaxLines = 2000

// WinRM connection probe states.
const (
	WinRMStatusConnecting = "connecting"
	WinRMStatusReady      = "ready"
	WinRMStatusFailed     = "error"
)

// winRMProbeCommand asks PowerShell for the version, user, host and default
// working directory, shown as the console's connection banner.
const winRMProbeCommand = `"PowerShell " + $PSVersionTable.PSVersion.ToString() + " · " + $env:USERNAME + "@" + $env:COMPUTERNAME + " · " + (Get-Location).Path`

// WinRMLine is one command or output line in the WinRM console.
type WinRMLine struct {
	Stream string // "cmd" | "stdout" | "stderr" | "exit" | "error"
	Text   string
}

// WinRMState is one Windows asset's PowerShell console. Commands run one
// at a time through WinRM; the remote process does not persist between them.
type WinRMState struct {
	HostID   string
	HostName string
	Input    string
	Lines    []WinRMLine
	Running  bool
	Scroll   ui.ScrollState
	Status   string
	Info     string

	seq      uint64
	probeSeq uint64
}

// IsWinRMProtocol reports whether a protocol is WinRM.
func (a *State) IsWinRMProtocol(protocol string) bool {
	return strings.EqualFold(protocol, "winrm")
}

// attachWinRM shows the console of a Windows asset, creating it on first use.
func (a *State) attachWinRM(n *hoststore.TreeNode) {
	if n == nil || n.NodeType != "host" || !a.IsWinRMProtocol(n.Protocol) {
		return
	}
	if a.WinRM.HostID == n.ID {
		a.WinRM.HostName = n.Name
		return
	}
	a.saveWinRMActive()
	if st, ok := a.winrmStates[n.ID]; ok {
		a.WinRM = st
		a.WinRM.HostName = n.Name
		a.resumeWinRMState()
		return
	}
	a.WinRM = WinRMState{HostID: n.ID, HostName: n.Name}
	a.WinRMProbe()
}

// resumeWinRMState drops a command whose output was lost while the tab was
// inactive; the user can run it again. A connection probe interrupted by
// the switch runs again.
func (a *State) resumeWinRMState() {
	a.WinRM.Running = false
	if a.WinRM.Status == WinRMStatusConnecting {
		a.WinRM.Status = ""
		a.WinRMProbe()
	}
}

// WinRMProbe connects once and loads the PowerShell banner: version, user,
// host and default working directory.
func (a *State) WinRMProbe() {
	st := &a.WinRM
	if st.HostID == "" {
		return
	}
	st.probeSeq++
	seq := st.probeSeq
	hostID := st.HostID
	st.Status = WinRMStatusConnecting
	st.Info = ""
	if a.Backend == nil || a.Backend.Executor == nil {
		st.Status = WinRMStatusFailed
		st.Info = i18n.T("winrm.error.backend")
		return
	}
	ex := a.Backend.Executor.ExecFor(hostID)
	if ex == nil {
		st.Status = WinRMStatusFailed
		st.Info = i18n.T("winrm.error.executor")
		return
	}
	go func() {
		ch, err := ex.Exec(context.Background(), winRMProbeCommand)
		if err != nil {
			a.Post(func() {
				if a.WinRM.HostID != hostID || a.WinRM.probeSeq != seq {
					return
				}
				a.WinRM.Status = WinRMStatusFailed
				a.WinRM.Info = err.Error()
			})
			return
		}
		var out []string
		failure := ""
		for line := range ch {
			switch line.Stream {
			case "stdout":
				if text := strings.TrimSpace(line.Text); text != "" {
					out = append(out, text)
				}
			case "error", "exit":
				failure = line.Text
			}
		}
		a.Post(func() {
			if a.WinRM.HostID != hostID || a.WinRM.probeSeq != seq {
				return
			}
			switch {
			case len(out) > 0:
				a.WinRM.Status = WinRMStatusReady
				a.WinRM.Info = strings.Join(out, " ")
			case failure != "":
				a.WinRM.Status = WinRMStatusFailed
				a.WinRM.Info = failure
			default:
				a.WinRM.Status = WinRMStatusReady
			}
		})
	}()
}

// WinRMRunCommand runs the console input as a PowerShell command.
func (a *State) WinRMRunCommand() {
	st := &a.WinRM
	cmd := strings.TrimSpace(st.Input)
	if st.HostID == "" || st.Running || st.Status == WinRMStatusConnecting || cmd == "" {
		return
	}
	st.Input = ""
	st.Running = true
	st.seq++
	seq := st.seq
	hostID := st.HostID
	st.appendLine(WinRMLine{Stream: "cmd", Text: cmd})
	if a.Backend == nil || a.Backend.Executor == nil {
		st.Running = false
		st.appendLine(WinRMLine{Stream: "error", Text: i18n.T("winrm.error.backend")})
		return
	}
	ex := a.Backend.Executor.ExecFor(hostID)
	if ex == nil {
		st.Running = false
		st.appendLine(WinRMLine{Stream: "error", Text: i18n.T("winrm.error.executor")})
		return
	}
	go func() {
		ch, err := ex.Exec(context.Background(), cmd)
		if err != nil {
			a.Post(func() {
				if a.WinRM.HostID != hostID || a.WinRM.seq != seq {
					return
				}
				a.WinRM.Running = false
				a.WinRM.appendLine(WinRMLine{Stream: "error", Text: err.Error()})
			})
			return
		}
		for line := range ch {
			line := line
			a.Post(func() {
				if a.WinRM.HostID != hostID || a.WinRM.seq != seq {
					return
				}
				a.WinRM.appendLine(WinRMLine{Stream: line.Stream, Text: line.Text})
			})
		}
		a.Post(func() {
			if a.WinRM.HostID != hostID || a.WinRM.seq != seq {
				return
			}
			a.WinRM.Running = false
		})
	}()
}

func (s *WinRMState) appendLine(line WinRMLine) {
	s.Lines = append(s.Lines, line)
	if len(s.Lines) > winRMMaxLines {
		s.Lines = append([]WinRMLine(nil), s.Lines[len(s.Lines)-winRMMaxLines:]...)
	}
	s.Scroll.Y = math.MaxFloat32
}

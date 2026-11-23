package hosts

import (
	"fmt"

	"github.com/egoist/mygo/ui"

	"nextopsx/i18n"
	"nextopsx/ui/style"
	"nextopsx/ui/workspace"
)

// Welcome shows the empty state, a selected folder, or a selected host with
// its actions.
func Welcome(c *ui.Context, a *workspace.State) {
	t := c.Theme()
	ui.Column(c).Fill().Padding(32).Gap(12).AlignItems(ui.Start).Children(func() {
		if a.Hosts.Err != "" {
			ui.Text(c, a.Hosts.Err).FontSize(style.FontSize(13)).TextColor(t.Danger).Wrap()
		}
		n := a.Selected
		if n == nil {
			ui.Text(c, i18n.T("welcome.empty.select")).FontSize(style.FontSize(18)).TextColor(t.TextMuted)
			ui.Text(c, i18n.T("welcome.empty.hint")).
				FontSize(style.FontSize(13)).TextColor(t.TextMuted)
			ui.Row(c).Gap(8).Children(func() {
				ui.PrimaryButton(c, i18n.T("welcome.newAsset")).OnClick(a.OpenNewHostForm)
				ui.Button(c, i18n.T("welcome.newFolder")).OnClick(a.OpenNewFolderForm)
			})
			return
		}

		if n.NodeType == "folder" {
			ui.Text(c, n.Name).FontSize(style.FontSize(22)).Bold()
			ui.Text(c, i18n.Tf("welcome.folder.subtitle", len(n.Children))).
				TextColor(t.TextMuted)
			ui.Text(c, i18n.T("welcome.folder.hint")).
				FontSize(style.FontSize(13)).TextColor(t.TextMuted)
			return
		}

		if a.IsRedisProtocol(n.Protocol) || a.IsESProtocol(n.Protocol) {
			ui.Row(c).Grow(1).FillWidth().Center().Children(func() {
				ui.Column(c).Gap(6).AlignItems(ui.Center).Children(func() {
					ui.Text(c, i18n.T("welcome.workbench.openHint")).
						FontSize(style.FontSize(15)).TextColor(t.TextMuted)
					ui.Text(c, a.HostProtocolName(n.Protocol)).
						FontSize(style.FontSize(12)).TextColor(t.TextMuted)
				})
			})
			return
		}

		if a.IsSQLProtocol(n.Protocol) {
			meta := a.HostProtocolName(n.Protocol)
			if db := a.DBDatabaseLabel(); db != "" {
				meta += " · " + db
			}
			ui.Row(c).Grow(1).FillWidth().Center().Children(func() {
				ui.Column(c).Gap(6).AlignItems(ui.Center).Children(func() {
					ui.Text(c, i18n.T("welcome.db.openHint")).
						FontSize(style.FontSize(15)).TextColor(t.TextMuted)
					ui.Text(c, meta).FontSize(style.FontSize(12)).TextColor(t.TextMuted)
				})
			})
			return
		}

		if a.IsSSHProtocol(n.Protocol) {
			session := a.SessionFor(n.ID)
			ui.Row(c).Grow(1).FillWidth().Center().Children(func() {
				ui.Column(c).Gap(6).AlignItems(ui.Center).Children(func() {
					if session != nil && session.Status == workspace.TerminalConnecting {
						ui.Row(c).Gap(8).AlignItems(ui.Center).Children(func() {
							ui.Spinner(c)
							ui.Text(c, i18n.T("welcome.connecting")).TextColor(t.TextMuted)
						})
						return
					}
					if session != nil && session.Status == workspace.TerminalClosed {
						if session.Err != "" {
							ui.Text(c, session.Err).FontSize(style.FontSize(13)).TextColor(t.Danger).Wrap()
						} else {
							ui.Text(c, i18n.T("terminal.disconnected")).
								FontSize(style.FontSize(13)).TextColor(t.TextMuted)
						}
						ui.Button(c, i18n.T("terminal.tabs.reconnect")).OnClick(func() { a.Connect(n) })
						return
					}
					ui.Text(c, i18n.T("welcome.host.openHint")).
						FontSize(style.FontSize(15)).TextColor(t.TextMuted)
					ui.Text(c, i18n.Tf("welcome.host.meta",
						a.HostProtocolName(n.Protocol), a.AuthDisplay(n.AuthType), a.AutoApproveDisplay(n.AutoApprove))).
						FontSize(style.FontSize(12)).TextColor(t.TextMuted)
				})
			})
			return
		}

		if a.IsWinRMProtocol(n.Protocol) {
			ui.Row(c).Grow(1).FillWidth().Center().Children(func() {
				ui.Column(c).Gap(6).AlignItems(ui.Center).Children(func() {
					ui.Text(c, i18n.T("welcome.winrm.openHint")).
						FontSize(style.FontSize(15)).TextColor(t.TextMuted)
					ui.Text(c, a.HostProtocolName(n.Protocol)).
						FontSize(style.FontSize(12)).TextColor(t.TextMuted)
				})
			})
			return
		}

		ui.Text(c, n.Name).FontSize(style.FontSize(22)).Bold()
		ui.Text(c, fmt.Sprintf("%s@%s:%d", n.User, n.Addr, n.Port)).
			TextColor(t.TextMuted)
		ui.Text(c, i18n.Tf("welcome.host.meta",
			a.HostProtocolName(n.Protocol), a.AuthDisplay(n.AuthType), a.AutoApproveDisplay(n.AutoApprove))).
			FontSize(style.FontSize(13)).TextColor(t.TextMuted)
		// connection/RDP errors are shown once in Shell's header strip.
	})
}

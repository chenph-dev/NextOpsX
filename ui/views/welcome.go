package views

import (
	"fmt"

	"github.com/egoist/mygo/ui"

	"nextopsx/i18n"
	"nextopsx/ui/app"
)

// Welcome shows the empty state, a selected folder, or a selected host with
// its actions.
func Welcome(c *ui.Context, a *app.App) {
	t := c.Theme()
	ui.Column(c).Fill().Padding(32).Gap(12).AlignItems(ui.Start).Children(func() {
		if a.Hosts.Err != "" {
			ui.Text(c, a.Hosts.Err).FontSize(fontSize(13)).TextColor(t.Danger).Wrap()
		}
		n := a.Selected
		if n == nil {
			ui.Text(c, i18n.T("welcome.empty.select")).FontSize(fontSize(18)).TextColor(t.TextMuted)
			ui.Text(c, i18n.T("welcome.empty.hint")).
				FontSize(fontSize(13)).TextColor(t.TextMuted)
			ui.Row(c).Gap(8).Children(func() {
				ui.PrimaryButton(c, i18n.T("welcome.newAsset")).OnClick(a.OpenNewHostForm)
				ui.Button(c, i18n.T("welcome.newFolder")).OnClick(a.OpenNewFolderForm)
			})
			backgroundTerminal(c, a, t)
			return
		}

		if n.NodeType == "folder" {
			ui.Text(c, n.Name).FontSize(fontSize(22)).Bold()
			ui.Text(c, i18n.Tf("welcome.folder.subtitle", len(n.Children))).
				TextColor(t.TextMuted)
			ui.Text(c, i18n.T("welcome.folder.hint")).
				FontSize(fontSize(13)).TextColor(t.TextMuted)
			backgroundTerminal(c, a, t)
			return
		}

		if a.IsSSHProtocol(n.Protocol) {
			ui.Row(c).Grow(1).FillWidth().Center().Children(func() {
				ui.Column(c).Gap(4).AlignItems(ui.Center).Children(func() {
					if a.Busy {
						ui.Row(c).Gap(8).AlignItems(ui.Center).Children(func() {
							ui.Spinner(c)
							ui.Text(c, i18n.T("welcome.connecting")).TextColor(t.TextMuted)
						})
						return
					}
					ui.Text(c, i18n.T("welcome.host.openHint")).
						FontSize(fontSize(15)).TextColor(t.TextMuted)
					ui.Text(c, i18n.Tf("welcome.host.meta",
						a.HostProtocolName(n.Protocol), a.AuthDisplay(n.AuthType), a.AutoApproveDisplay(n.AutoApprove))).
						FontSize(fontSize(12)).TextColor(t.TextMuted)
				})
			})
			backgroundTerminal(c, a, t)
			return
		}

		ui.Text(c, n.Name).FontSize(fontSize(22)).Bold()
		ui.Text(c, fmt.Sprintf("%s@%s:%d", n.User, n.Addr, n.Port)).
			TextColor(t.TextMuted)
		ui.Text(c, i18n.Tf("welcome.host.meta",
			a.HostProtocolName(n.Protocol), a.AuthDisplay(n.AuthType), a.AutoApproveDisplay(n.AutoApprove))).
			FontSize(fontSize(13)).TextColor(t.TextMuted)
		// connection/RDP errors are shown once in Shell's header strip.
		backgroundTerminal(c, a, t)
	})
}

// backgroundTerminal reminds the user that another asset's terminal is still
// open when the current pane cannot show it.
func backgroundTerminal(c *ui.Context, a *app.App, t *ui.Theme) {
	if a.Term == nil {
		return
	}
	if a.Selected != nil && a.Selected.ID == a.TermHostID {
		return
	}
	ui.Row(c).Gap(8).AlignItems(ui.Center).Children(func() {
		ui.Text(c, i18n.Tf("welcome.backgroundTerminal", a.TermHostLabel())).
			FontSize(fontSize(12)).TextColor(t.TextMuted)
		ui.Button(c, i18n.T("welcome.closeBackground")).FontSize(fontSize(12)).OnClick(a.Disconnect)
	})
}

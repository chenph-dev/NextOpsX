package terminal

import (
	"github.com/egoist/mygo/ui"

	"nextopsx/i18n"
	"nextopsx/ui/assets"
	"nextopsx/ui/style"
	"nextopsx/ui/workspace"
)

// Tabs renders the session bar above the workbench. Each tab keeps its SSH
// connection alive independently of the selected asset.
func Tabs(c *ui.Context, a *workspace.State) {
	if len(a.Sessions) == 0 {
		return
	}
	t := c.Theme()
	ui.Row(c).Height(34).Shrink(0).FillWidth().Padding(0, 8).Gap(6).
		AlignItems(ui.Center).Background(t.Surface).
		BorderWidth(0, 0, 1, 0).BorderColor(t.Border).Children(func() {
		ui.ScrollHorizontal(c).Grow(1).Children(func() {
			ui.Row(c).Gap(4).AlignItems(ui.Center).Children(func() {
				for _, s := range a.Sessions {
					sessionTab(c, a, t, s)
				}
			})
		})
	})
}

func sessionTab(c *ui.Context, a *workspace.State, t *ui.Theme, s *workspace.TerminalSession) {
	active := a.ActiveSessionID == s.HostID
	live := a.SessionLive(s)

	row := ui.Row(c).Height(26).Radius(6).AlignItems(ui.Center)
	if active {
		row = row.Background(t.SurfaceHover).Border(1, t.Border)
	}
	row.Children(func() {
		open := ui.ButtonBase(c).Padding(3, 6).Radius(5).Gap(6).
			OnClick(func() { a.ActivateSession(s.HostID) })
		nameColor := t.Text
		switch {
		case s.Status == workspace.TerminalConnecting:
			open = open.Tooltip(i18n.T("terminal.tabs.connecting"))
		case !live:
			nameColor = t.TextMuted
			open = open.Tooltip(i18n.T("terminal.tabs.closed"))
		}
		open.Children(func() {
			if b := assets.ProtocolBitmap(s.Protocol); b != nil {
				ui.Image(c, b).Size(14, 14)
			}
			ui.Text(c, a.HostLabel(s.HostID)).FontSize(style.FontSize(12)).
				TextColor(nameColor).SingleLine().MaxWidth(180)
			dot := t.Success
			switch {
			case s.Status == workspace.TerminalConnecting:
				dot = t.Accent
			case !live:
				dot = t.TextMuted
			}
			ui.Box(c).Size(6, 6).Radius(3).Background(dot)
		})

		closeButton := ui.ButtonBase(c).Padding(3).Radius(4).
			Tooltip(i18n.T("terminal.tabs.close")).
			OnClick(func() { a.CloseSession(s.HostID) })
		closeButton.Children(func() {
			if svg := assets.SVG("close"); svg != nil {
				ui.Icon(c, svg).Size(10, 10).TextColor(t.TextMuted)
			}
		})
	})
}

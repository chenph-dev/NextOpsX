package views

import (
	"github.com/egoist/mygo/ui"

	"nextopsx/i18n"
	"nextopsx/ui/assets"
	"nextopsx/ui/style"
	"nextopsx/ui/workspace"
)

// Tabs renders the open asset tabs above the workbench. Each tab keeps its
// connection and workbench state independently of the selected asset.
func Tabs(c *ui.Context, a *workspace.State) {
	if len(a.Tabs) == 0 {
		return
	}
	t := c.Theme()
	ui.Row(c).Height(34).Shrink(0).FillWidth().Padding(0, 8).Gap(6).
		AlignItems(ui.Center).Background(t.Surface).
		BorderWidth(0, 0, 1, 0).BorderColor(t.Border).Children(func() {
		ui.ScrollHorizontal(c).TrackScroll(&a.TabScroll).Grow(1).Children(func() {
			ui.Row(c).Gap(4).AlignItems(ui.Center).Children(func() {
				for _, id := range a.Tabs {
					assetTab(c, a, t, id)
				}
			})
		})
		ui.Box(c).Width(1).Height(16).Background(t.Border)
		tabArrow(c, t, "chevron-left", i18n.T("terminal.tabs.scrollLeft"),
			a.TabScroll.X <= 0, a.ScrollTabsLeft)
		tabArrow(c, t, "chevron-right", i18n.T("terminal.tabs.scrollRight"),
			a.TabScroll.MaxX <= 0 || a.TabScroll.X >= a.TabScroll.MaxX, a.ScrollTabsRight)
	})
}

func tabArrow(c *ui.Context, t *ui.Theme, icon, label string, disabled bool, onClick func()) {
	color := t.TextMuted
	btn := ui.ButtonBase(c).Size(22, 22).Radius(5).
		Tooltip(label).Label(label).Cursor(ui.CursorPointer).OnClick(onClick)
	if disabled {
		btn = btn.Disabled(true).Cursor(ui.CursorDefault)
		color = color.Alpha(0.35)
	} else if btn.Hovered() {
		btn = btn.Background(t.SurfaceHover)
	}
	btn.Children(func() {
		if svg := assets.SVG(icon); svg != nil {
			ui.Icon(c, svg).Size(14, 14).TextColor(color)
		}
	})
}

func assetTab(c *ui.Context, a *workspace.State, t *ui.Theme, id string) {
	active := a.Selected != nil && a.Selected.ID == id
	protocol := a.HostProtocol(id)
	session := a.SessionFor(id)

	row := ui.Row(c).Height(26).Radius(6).AlignItems(ui.Center)
	if active {
		row = row.Background(t.SurfaceHover).Border(1, t.Border)
	}
	row.Children(func() {
		open := ui.ButtonBase(c).Padding(3, 6).Radius(5).Gap(6).
			OnClick(func() { a.ActivateAssetTab(id) })
		nameColor := t.Text
		switch {
		case session != nil && session.Status == workspace.TerminalConnecting:
			open = open.Tooltip(i18n.T("terminal.tabs.connecting"))
		case session != nil && !a.SessionLive(session):
			nameColor = t.TextMuted
			open = open.Tooltip(i18n.T("terminal.tabs.closed"))
		}
		open.Children(func() {
			if b := assets.ProtocolBitmap(protocol); b != nil {
				ui.Image(c, b).Size(14, 14)
			}
			ui.Text(c, a.HostLabel(id)).FontSize(style.FontSize(12)).
				TextColor(nameColor).SingleLine().MaxWidth(180)
			if session != nil {
				dot := t.Success
				switch {
				case session.Status == workspace.TerminalConnecting:
					dot = t.Accent
				case !a.SessionLive(session):
					dot = t.TextMuted
				}
				ui.Box(c).Size(6, 6).Radius(3).Background(dot)
			}
		})

		closeLabel := i18n.T("terminal.tabs.close")
		closeButton := ui.ButtonBase(c).Size(20, 20).Radius(4).
			Tooltip(closeLabel).Label(closeLabel).
			Cursor(ui.CursorPointer).OnClick(func() { a.CloseAssetTab(id) })
		closeColor := t.TextMuted
		if closeButton.Hovered() {
			closeButton = closeButton.Background(t.SurfacePressed)
			closeColor = t.Text
		}
		closeButton.Children(func() {
			if svg := assets.SVG("close"); svg != nil {
				ui.Icon(c, svg).Size(13, 13).TextColor(closeColor)
			}
		})
	})
	if active && a.TakeTabScroll(id) {
		row.ScrollIntoView()
	}
}

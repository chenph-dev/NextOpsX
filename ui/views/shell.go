// Package views renders the native UI from app state. It performs no I/O:
// actions go through *app.App.
package views

import (
	"fmt"

	"github.com/egoist/mygo/ui"

	"nextopsx/i18n"
	configstore "nextopsx/store/config"
	"nextopsx/ui/app"
	"nextopsx/ui/assets"
)

// Base themes are created once; each frame copies one and applies the
// desktop's preferences, so an explicit theme keeps accent, contrast and text
// scale like the system-following default does.
var (
	lightThemeBase = ui.LightTheme()
	darkThemeBase  = ui.DarkTheme()
	passTheme      ui.Theme
	uiScale        float32 = 1
)

// fontSize scales an explicit text size by the user's UI scale.
func fontSize(v float32) float32 { return v * uiScale }

// applyTheme applies the user's theme and UI scale. The root view rebuilds
// every frame, which is where SetTheme takes effect for the whole tree.
func applyTheme(c *ui.Context, a *app.App) {
	scale := float32(a.Settings.Scale) / 100
	if scale <= 0 {
		scale = 1
	}
	uiScale = scale

	prefs := c.Preferences()
	var t ui.Theme
	switch a.Settings.Theme {
	case configstore.ThemeLight:
		t = *lightThemeBase
		t.FontSize = 13
		followPreferences(&t, prefs)
	case configstore.ThemeDark:
		t = *darkThemeBase
		t.FontSize = 13
		followPreferences(&t, prefs)
	default:
		// 跟随系统：默认主题已应用外观与系统文字缩放，这里统一基准字号。
		t = *c.Theme()
		textScale := prefs.TextScale
		if textScale <= 0 {
			textScale = 1
		}
		t.FontSize = 13 * textScale
	}
	if scale != 1 {
		t.FontSize *= scale
		t.Spacing *= scale
		t.Radius *= scale
	}
	passTheme = t
	c.SetTheme(&passTheme)
}

// followPreferences mirrors MyGo's unexported Theme.follow (v0.4.8), which the
// default theme applies but an overriding theme does not.
func followPreferences(t *ui.Theme, p ui.Preferences) {
	if a := p.Accent; a.A > 0 {
		a.A = 255
		black := ui.Color{A: 255}
		white := ui.Color{R: 255, G: 255, B: 255, A: 255}
		t.Accent = a
		if t.Dark {
			t.AccentHover = a.Mix(white, 0.2)
			t.AccentPressed = a.Mix(black, 0.15)
			t.Selection = a.Alpha(0.4)
			t.Focus = a.Mix(white, 0.2).Alpha(0.6)
		} else {
			t.AccentHover = a.Mix(black, 0.12)
			t.AccentPressed = a.Mix(black, 0.25)
			t.Selection = a.Alpha(0.25)
			t.Focus = a.Alpha(0.55)
		}
		l := uint8(0.2126*float32(a.R) + 0.7152*float32(a.G) + 0.0722*float32(a.B) + 0.5)
		t.AccentText = white
		if l > 165 {
			t.AccentText = ui.Color{R: 24, G: 24, B: 27, A: 255}
		}
	}
	if p.HighContrast {
		t.Border = t.Border.Mix(t.Text, 0.45)
		t.TextMuted = t.TextMuted.Mix(t.Text, 0.4)
		t.Focus.A = 255
		t.Scrollbar = t.Scrollbar.Alpha(1.6)
	}
	if s := p.TextScale; s > 0 && s != 1 {
		t.FontSize *= s
	}
}

// Shell is the window layout: the asset tree on the left, the protocol's
// workbench on the right, with the ops assistant and file drawers around it.
func Shell(a *app.App) func(*ui.Context) {
	return func(c *ui.Context) {
		applyTheme(c, a)
		if a.View == app.ViewSettings {
			SettingsView(c, a)
			// 资产表单可能从菜单打开；设置页也渲染对话框，避免弹窗被隐藏后残留。
			HostsDialogs(c, a)
			return
		}
		t := c.Theme()
		c.OnShortcut(ui.Ctrl|ui.Shift, ui.KeyJ, a.ToggleAIDrawer)
		c.OnShortcut(ui.Super|ui.Shift, ui.KeyJ, a.ToggleAIDrawer)
		c.OnShortcut(ui.Ctrl|ui.Shift, ui.KeyE, a.ToggleSFTPDrawer)
		c.OnShortcut(ui.Super|ui.Shift, ui.KeyE, a.ToggleSFTPDrawer)
		ui.Row(c).Fill().Children(func() {
			ui.Column(c).Width(280).FillHeight().Background(t.Surface).
				Border(1, t.Border).Children(func() {
				ui.Row(c).Padding(12).Gap(8).AlignItems(ui.Center).Children(func() {
					ui.Text(c, i18n.T("shell.assets")).FontSize(fontSize(16)).Bold()
					ui.Text(c, fmt.Sprintf("%d", countHosts(a.Roots))).TextColor(t.TextMuted)
					ui.Spacer(c)
					ui.Button(c, i18n.T("shell.newAsset")).FontSize(fontSize(12)).OnClick(a.OpenNewHostForm)
					ui.Button(c, i18n.T("shell.newFolder")).FontSize(fontSize(12)).OnClick(a.OpenNewFolderForm)
				})
				tree := ui.Scroll(c).Grow(1)
				tree.ContextMenu(func(m *ui.Menu) {
					if m.Item(i18n.T("shell.newAssetMenu")).Chosen() {
						a.Select(nil)
						a.OpenNewHostForm()
					}
					if m.Item(i18n.T("shell.newFolderMenu")).Chosen() {
						a.Select(nil)
						a.OpenNewFolderForm()
					}
					m.Separator()
					if m.Item(i18n.T("shell.refresh")).Chosen() {
						a.RefreshHosts()
					}
				})
				tree.Children(func() {
					HostTree(c, a)
				})
				ui.Divider(c)
				ui.Row(c).Padding(8).Children(func() {
					ui.Button(c, i18n.T("shell.settings")).Grow(1).
						OnClick(func() { a.OpenSettings(a.Settings.Tab) })
				})
			})

			ui.Column(c).Grow(1).FillHeight().MinWidth(0).Children(func() {
				if a.Selected == nil {
					Welcome(c, a)
					return
				}
				ui.Row(c).Padding(10, 16).Gap(10).AlignItems(ui.Center).Children(func() {
					ui.Text(c, a.Selected.Name).FontSize(fontSize(15)).Bold()
					if a.Selected.NodeType == "host" {
						ui.Text(c, fmt.Sprintf("%s@%s:%d", a.Selected.User, a.Selected.Addr, a.Selected.Port)).
							FontSize(fontSize(13)).TextColor(t.TextMuted)
						ui.Spacer(c)
						if a.Selected.Protocol == "winrm" {
							ui.Button(c, i18n.T("shell.remoteDesktop")).FontSize(fontSize(12)).OnClick(a.OpenRDP)
						}
						shellToggle(c, t, a.Layout.AIOpen, i18n.T("shell.tab.ai"), "spark", a.ToggleAIDrawer)
						if label, color := aiStateBadge(a, t); label != "" {
							ui.Box(c).Size(6, 6).Radius(3).Background(color)
						}
						if a.IsSSHProtocol(a.Selected.Protocol) {
							shellToggle(c, t, a.Layout.SFTPOpen, "SFTP", "folder", a.ToggleSFTPDrawer)
							snippets := ui.Button(c, i18n.T("shell.history")).
								FontSize(fontSize(12)).
								Tooltip(i18n.T("history.title")).
								OnClick(a.ToggleSnippetPicker)
							if !a.CanUseSnippet() {
								snippets = snippets.Disabled(true).
									Tooltip(i18n.T("snippets.picker.connectFirst"))
							}
							ui.Popover(c, snippets, &a.Snippets.PickerOpen, func() {
								snippetPicker(c, a, t)
							})
						}
						return
					}
					ui.Text(c, i18n.T("shell.folder")).FontSize(fontSize(13)).TextColor(t.TextMuted)
				})
				if a.ErrMsg != "" {
					ui.Row(c).Padding(0, 16).Children(func() {
						ui.Text(c, a.ErrMsg).FontSize(fontSize(12)).TextColor(t.Danger).Wrap()
					})
				}
				ui.Divider(c)
				if a.Selected.NodeType != "host" {
					Welcome(c, a)
					return
				}
				hostWorkbench(c, a, t)
			})
		})
		HostsDialogs(c, a)
	}
}

// shellToggle is the header's pane switch: a quiet button that fills in while
// its drawer is open.
func shellToggle(c *ui.Context, t *ui.Theme, active bool, label, icon string, onClick func()) {
	color := t.TextMuted
	if active {
		color = t.Accent
	}
	btn := ui.ButtonBase(c).Padding(4, 8).Radius(6).Gap(6).
		AlignItems(ui.Center).Cursor(ui.CursorPointer).OnClick(onClick)
	if active {
		btn = btn.Background(t.SurfaceHover)
	} else if btn.Hovered() {
		btn = btn.Background(t.Surface)
	}
	btn.Children(func() {
		if s := assets.SVG(icon); s != nil {
			ui.Icon(c, s).Size(14, 14).TextColor(color)
		}
		ui.Text(c, label).FontSize(fontSize(12)).TextColor(color)
	})
}

// hostWorkbench is the persistent main view with the optional file drawer on
// the right and the intelligent ops drawer below.
func hostWorkbench(c *ui.Context, a *app.App, t *ui.Theme) {
	ui.Column(c).Grow(1).FillHeight().FillWidth().MinHeight(0).Children(func() {
		ui.Row(c).Grow(1).FillWidth().MinHeight(0).Children(func() {
			ui.Column(c).Grow(1).FillHeight().MinWidth(0).Children(func() {
				hostMain(c, a, t)
			})
			if a.Layout.SFTPOpen && a.Selected != nil &&
				a.IsSSHProtocol(a.Selected.Protocol) {
				handle := ui.Box(c).Width(5).FillHeight().Cursor(ui.CursorResizeEW)
				if handle.Hovered() {
					handle = handle.Background(t.Accent)
				}
				if dx, _, ok := handle.Dragged(); ok && dx != 0 {
					a.SetSFTPDrawerWidth(a.SFTPDrawerWidth() - dx)
				}
				if _, _, done := handle.DragEnded(); done {
					a.SaveLayout()
				}
				ui.Column(c).Width(a.SFTPDrawerWidth()).FillHeight().MinWidth(0).
					BorderWidth(0, 0, 0, 1).BorderColor(t.Border).Children(func() {
					SFTPPanel(c, a)
				})
			}
		})
		if a.Layout.AIOpen {
			handle := ui.Box(c).Height(5).FillWidth().Cursor(ui.CursorResizeNS)
			if handle.Hovered() {
				handle = handle.Background(t.Accent)
			}
			if _, dy, ok := handle.Dragged(); ok && dy != 0 {
				a.SetAIDrawerHeight(a.AIDrawerHeight() - dy)
			}
			if _, _, done := handle.DragEnded(); done {
				a.SaveLayout()
			}
			ui.Column(c).Height(a.AIDrawerHeight()).Shrink(0).FillWidth().MinHeight(0).
				BorderWidth(1, 0, 0, 0).BorderColor(t.Border).Children(func() {
				AIPanel(c, a)
			})
		} else {
			aiDrawerBar(c, a, t)
		}
	})
}

// hostMain picks the workbench from the asset's protocol. There is no pane
// switch: the protocol's main view is always the base.
func hostMain(c *ui.Context, a *app.App, t *ui.Theme) {
	if a.Selected == nil {
		Welcome(c, a)
		return
	}
	switch {
	case a.IsRedisProtocol(a.Selected.Protocol):
		RedisPanel(c, a)
	case a.IsESProtocol(a.Selected.Protocol):
		ESPanel(c, a)
	case a.IsDBProtocol(a.Selected.Protocol):
		DBWorkbench(c, a)
	case a.IsSSHProtocol(a.Selected.Protocol):
		if a.Term != nil && a.TermHostID == a.Selected.ID {
			terminalWithCompletion(c, a, t)
			return
		}
		Welcome(c, a)
	default:
		Welcome(c, a)
	}
}

// aiDrawerBar keeps the intelligent ops drawer one click away when collapsed.
func aiDrawerBar(c *ui.Context, a *app.App, t *ui.Theme) {
	bar := ui.Row(c).Height(34).Shrink(0).FillWidth().Padding(0, 12).Gap(8).
		AlignItems(ui.Center).Background(t.Surface).
		BorderWidth(1, 0, 0, 0).BorderColor(t.Border).
		Cursor(ui.CursorPointer).OnClick(a.ToggleAIDrawer)
	bar.Children(func() {
		if s := assets.SVG("spark"); s != nil {
			ui.Icon(c, s).Size(13, 13).TextColor(t.Accent)
		}
		ui.Text(c, i18n.T("ai.conversation.defaultTitle")).FontSize(fontSize(12)).Bold()
		if label, color := aiStateBadge(a, t); label != "" {
			ui.Text(c, label).FontSize(fontSize(10)).TextColor(color)
		}
		ui.Spacer(c)
		ui.Text(c, i18n.T("ai.drawer.expand")).FontSize(fontSize(10)).TextColor(t.TextMuted)
	})
}

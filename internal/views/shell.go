// Package views renders the native UI from app state. It performs no I/O:
// actions go through *app.App.
package views

import (
	"fmt"

	"github.com/egoist/mygo/plugins/terminal"
	"github.com/egoist/mygo/ui"

	"nextopsx/internal/app"
	configstore "nextopsx/internal/store/config"
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

// Shell is the window layout: the asset tree on the left, the terminal or
// AI chat on the right.
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
		ui.Row(c).Fill().Children(func() {
			ui.Column(c).Width(280).FillHeight().Background(t.Surface).
				Border(1, t.Border).Children(func() {
				ui.Row(c).Padding(12).Gap(8).AlignItems(ui.Center).Children(func() {
					ui.Text(c, "资产").FontSize(fontSize(16)).Bold()
					ui.Text(c, fmt.Sprintf("%d", countHosts(a.Roots))).TextColor(t.TextMuted)
					ui.Spacer(c)
					ui.Button(c, "＋资产").FontSize(fontSize(12)).OnClick(a.OpenNewHostForm)
					ui.Button(c, "＋目录").FontSize(fontSize(12)).OnClick(a.OpenNewFolderForm)
				})
				tree := ui.Scroll(c).Grow(1)
				tree.ContextMenu(func(m *ui.Menu) {
					if m.Item("新建资产").Chosen() {
						a.Select(nil)
						a.OpenNewHostForm()
					}
					if m.Item("新建目录").Chosen() {
						a.Select(nil)
						a.OpenNewFolderForm()
					}
					m.Separator()
					if m.Item("刷新").Chosen() {
						a.RefreshHosts()
					}
				})
				tree.Children(func() {
					HostTree(c, a)
				})
				ui.Divider(c)
				ui.Row(c).Padding(8).Children(func() {
					ui.Button(c, "设置").Grow(1).
						OnClick(func() { a.OpenSettings(a.Settings.Tab) })
				})
			})

			ui.Column(c).Grow(1).FillHeight().MinWidth(0).Children(func() {
				if a.Selected == nil {
					Welcome(c, a)
					return
				}
				ui.Row(c).Padding(10, 16).Gap(12).AlignItems(ui.Center).Children(func() {
					ui.Text(c, a.Selected.Name).FontSize(fontSize(15)).Bold()
					if a.Selected.NodeType == "host" {
						ui.Text(c, fmt.Sprintf("%s@%s:%d", a.Selected.User, a.Selected.Addr, a.Selected.Port)).
							FontSize(fontSize(13)).TextColor(t.TextMuted)
						ui.Spacer(c)
						if a.Selected.Protocol == "winrm" {
							ui.Button(c, "远程桌面").FontSize(fontSize(12)).OnClick(a.OpenRDP)
						}
						labels := []string{"终端", "AI 对话"}
						switch {
						case a.IsRedisProtocol(a.Selected.Protocol):
							labels = []string{"Redis", "AI 对话"}
						case a.IsESProtocol(a.Selected.Protocol):
							labels = []string{"ES", "AI 对话"}
						case a.IsDBProtocol(a.Selected.Protocol):
							labels = []string{"数据库", "AI 对话"}
						case a.IsSSHProtocol(a.Selected.Protocol):
							labels = []string{"终端", "AI 对话", "SFTP"}
						default:
							labels = []string{"概览", "AI 对话"}
						}
						tabs := ui.Tabs(c, &a.RightTab, labels...)
						if tabs.Changed() {
							switch {
							case a.RightTab == 2 && a.IsSSHProtocol(a.Selected.Protocol):
								a.EnsureSFTP()
							case a.RightTab == 0 && a.IsRedisProtocol(a.Selected.Protocol):
								a.EnsureRedis()
							case a.RightTab == 0 && a.IsESProtocol(a.Selected.Protocol):
								a.EnsureES()
							}
						}
						return
					}
					ui.Text(c, "目录").FontSize(fontSize(13)).TextColor(t.TextMuted)
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
				ui.Column(c).Grow(1).FillHeight().Children(func() {
					switch a.RightTab {
					case 1:
						AIPanel(c, a)
						return
					case 2:
						SFTPPanel(c, a)
						return
					}
					if a.IsRedisProtocol(a.Selected.Protocol) {
						RedisPanel(c, a)
						return
					}
					if a.IsESProtocol(a.Selected.Protocol) {
						ESPanel(c, a)
						return
					}
					if a.IsDBProtocol(a.Selected.Protocol) {
						DBWorkbench(c, a)
						return
					}
					if a.Term != nil && a.TermHostID == a.Selected.ID {
						terminal.View(c, a.Term).Fill()
						return
					}
					Welcome(c, a)
				})
			})
		})
		HostsDialogs(c, a)
	}
}

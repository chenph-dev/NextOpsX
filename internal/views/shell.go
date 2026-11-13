// Package views renders the native UI from app state. It performs no I/O:
// actions go through *app.App.
package views

import (
	"fmt"

	"github.com/egoist/mygo/plugins/terminal"
	"github.com/egoist/mygo/ui"

	"nextopsx/internal/app"
)

// Shell is the window layout: the asset tree on the left, the terminal or
// AI chat on the right.
func Shell(a *app.App) func(*ui.Context) {
	return func(c *ui.Context) {
		if a.View == app.ViewSettings {
			SettingsView(c, a)
			return
		}
		t := c.Theme()
		ui.Row(c).Fill().Children(func() {
			ui.Column(c).Width(280).FillHeight().Background(t.Surface).
				Border(1, t.Border).Children(func() {
				ui.Row(c).Padding(12).Gap(8).AlignItems(ui.Center).Children(func() {
					ui.Text(c, "资产").FontSize(16).Bold()
					ui.Text(c, fmt.Sprintf("%d", countHosts(a.Roots))).TextColor(t.TextMuted)
					ui.Spacer(c)
					ui.Button(c, "＋资产").FontSize(11).OnClick(a.OpenNewHostForm)
					ui.Button(c, "＋目录").FontSize(11).OnClick(a.OpenNewFolderForm)
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
					ui.Text(c, a.Selected.Name).FontSize(15).Bold()
					if a.Selected.NodeType == "host" {
						ui.Text(c, fmt.Sprintf("%s@%s:%d", a.Selected.User, a.Selected.Addr, a.Selected.Port)).
							FontSize(12).TextColor(t.TextMuted)
						ui.Spacer(c)
						if a.Selected.Protocol == "winrm" {
							ui.Button(c, "远程桌面").FontSize(11).OnClick(a.OpenRDP)
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
					ui.Text(c, "目录").FontSize(12).TextColor(t.TextMuted)
				})
				if a.ErrMsg != "" {
					ui.Row(c).Padding(0, 16).Children(func() {
						ui.Text(c, a.ErrMsg).FontSize(11).TextColor(t.Danger).Wrap()
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

package views

import (
	"fmt"

	"github.com/egoist/mygo/ui"

	"nextopsx/internal/app"
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
			ui.Text(c, "选择左侧资产").FontSize(fontSize(18)).TextColor(t.TextMuted)
			ui.Text(c, "还没有资产时，可以先新建资产或目录。").
				FontSize(fontSize(13)).TextColor(t.TextMuted)
			ui.Row(c).Gap(8).Children(func() {
				ui.PrimaryButton(c, "新建资产").OnClick(a.OpenNewHostForm)
				ui.Button(c, "新建目录").OnClick(a.OpenNewFolderForm)
			})
			backgroundTerminal(c, a, t)
			return
		}

		if n.NodeType == "folder" {
			ui.Text(c, n.Name).FontSize(fontSize(22)).Bold()
			ui.Text(c, fmt.Sprintf("目录 · %d 个直接子节点", len(n.Children))).
				TextColor(t.TextMuted)
			ui.Row(c).Gap(8).Children(func() {
				ui.PrimaryButton(c, "在此新建资产").OnClick(a.OpenNewHostForm)
				ui.Button(c, "在此新建目录").OnClick(a.OpenNewFolderForm)
				ui.Button(c, "重命名").OnClick(a.OpenRenameDialog)
				ui.Button(c, "移动").OnClick(a.OpenMoveDialog)
				ui.Button(c, "删除").OnClick(a.OpenDeleteDialog)
			})
			return
		}

		ui.Text(c, n.Name).FontSize(fontSize(22)).Bold()
		ui.Text(c, fmt.Sprintf("%s@%s:%d", n.User, n.Addr, n.Port)).
			TextColor(t.TextMuted)
		ui.Text(c, fmt.Sprintf("协议：%s · 认证：%s · 自动放行：%s",
			a.HostProtocolName(n.Protocol), a.AuthDisplay(n.AuthType), a.AutoApproveDisplay(n.AutoApprove))).
			FontSize(fontSize(13)).TextColor(t.TextMuted)

		ui.Row(c).Gap(8).AlignItems(ui.Center).Children(func() {
			if a.IsSSHProtocol(n.Protocol) {
				if a.Busy {
					ui.Spinner(c)
					ui.Text(c, "正在连接…").TextColor(t.TextMuted)
				} else if ui.PrimaryButton(c, "连接").Clicked() {
					a.Connect(n)
				}
			}
			ui.Button(c, "编辑").OnClick(a.OpenEditHostForm)
			ui.Button(c, "移动").OnClick(a.OpenMoveDialog)
			ui.Button(c, "删除").OnClick(a.OpenDeleteDialog)
		})
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
		ui.Text(c, "后台仍保持到「"+a.TermHostLabel()+"」的终端会话。").
			FontSize(fontSize(12)).TextColor(t.TextMuted)
		ui.Button(c, "关闭后台终端").FontSize(fontSize(12)).OnClick(a.Disconnect)
	})
}

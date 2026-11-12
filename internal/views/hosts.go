package views

import (
	"github.com/egoist/mygo/ui"

	"nextopsx/internal/app"
	hoststore "nextopsx/internal/store/hosts"
	"nextopsx/resources"
)

// HostTree builds the folder/host tree and reports selection.
func HostTree(c *ui.Context, a *app.App) {
	ui.Tree(c, func() {
		buildNodes(c, a, a.Roots)
	})
}

func buildNodes(c *ui.Context, a *app.App, nodes []hoststore.TreeNode) {
	for i := range nodes {
		n := &nodes[i]
		isFolder := n.NodeType == "folder"
		label := n.Name

		// TreeItem always appends its label after the arrow. Build it with
		// an empty label and append the icon and text ourselves, so the
		// icon leads the node name.
		var item ui.Element
		if isFolder {
			item = ui.TreeItem(c, "", a.Open[n.ID], func() {
				buildNodes(c, a, n.Children)
			})
		} else {
			item = ui.TreeItem(c, "", nil, nil)
		}
		item.Children(func() {
			hostTreeIcon(c, n)
			ui.Text(c, label).SingleLine()
		})

		selected := a.Selected != nil && a.Selected.ID == n.ID
		item = item.Selected(selected)
		if isFolder {
			item.ContextMenu(func(m *ui.Menu) { folderMenu(m, a, n) })
		} else {
			item = item.Tooltip("协议：" + a.HostProtocolName(n.Protocol))
			item.ContextMenu(func(m *ui.Menu) { hostMenu(m, a, n) })
		}
		if item.Clicked() {
			a.Select(n)
		}
	}
}

func hostTreeIcon(c *ui.Context, n *hoststore.TreeNode) {
	if n.NodeType == "folder" {
		if s := resources.FolderSVG(); s != nil {
			ui.Icon(c, s).Size(15, 15).TextColor(c.Theme().TextMuted)
		}
		return
	}
	if b := resources.ProtocolBitmap(n.Protocol); b != nil {
		ui.Image(c, b).Size(16, 16)
	}
}

func folderMenu(m *ui.Menu, a *app.App, n *hoststore.TreeNode) {
	if m.Item("在此新建资产").Chosen() {
		a.Select(n)
		a.OpenNewHostForm()
	}
	if m.Item("在此新建目录").Chosen() {
		a.Select(n)
		a.OpenNewFolderForm()
	}
	m.Separator()
	if m.Item("重命名").Chosen() {
		a.Select(n)
		a.OpenRenameDialog()
	}
	if m.Item("移动").Chosen() {
		a.Select(n)
		a.OpenMoveDialog()
	}
	m.Separator()
	if m.Item("删除").Chosen() {
		a.Select(n)
		a.OpenDeleteDialog()
	}
}

func hostMenu(m *ui.Menu, a *app.App, n *hoststore.TreeNode) {
	if a.IsSSHProtocol(n.Protocol) {
		if m.Item("打开终端").Chosen() {
			a.Select(n)
			a.RightTab = 0
			a.Connect(n)
		}
	}
	if m.Item("打开 AI 对话").Chosen() {
		a.Select(n)
		a.RightTab = 1
	}
	m.Separator()
	if m.Item("编辑").Chosen() {
		a.Select(n)
		a.OpenEditHostForm()
	}
	if m.Item("移动").Chosen() {
		a.Select(n)
		a.OpenMoveDialog()
	}
	m.Separator()
	if m.Item("删除").Chosen() {
		a.Select(n)
		a.OpenDeleteDialog()
	}
}

func countHosts(nodes []hoststore.TreeNode) int {
	n := 0
	for i := range nodes {
		if nodes[i].NodeType == "host" {
			n++
		}
		n += countHosts(nodes[i].Children)
	}
	return n
}

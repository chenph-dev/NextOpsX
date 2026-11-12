package views

import (
	"github.com/egoist/mygo/ui"

	"nextopsx/internal/app"
	hoststore "nextopsx/internal/store/hosts"
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
		if n.NodeType == "folder" {
			item := ui.TreeItem(c, n.Name, a.Open[n.ID], func() {
				buildNodes(c, a, n.Children)
			})
			selected := a.Selected != nil && a.Selected.ID == n.ID
			item = item.Selected(selected)
			item.ContextMenu(func(m *ui.Menu) { folderMenu(m, a, n) })
			if item.Clicked() {
				a.Select(n)
			}
			continue
		}
		label := n.Name
		if n.Protocol != "" && n.Protocol != "ssh" {
			label = n.Name + "  (" + n.Protocol + ")"
		}
		selected := a.Selected != nil && a.Selected.ID == n.ID
		item := ui.TreeItem(c, label, nil, nil).Selected(selected)
		item.ContextMenu(func(m *ui.Menu) { hostMenu(m, a, n) })
		if item.Clicked() {
			a.Select(n)
		}
	}
}

func folderMenu(m *ui.Menu, a *app.App, n *hoststore.TreeNode) {
	if m.Item("在此新建主机").Chosen() {
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

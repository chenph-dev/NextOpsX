package hosts

import (
	"github.com/egoist/mygo/ui"

	"nextopsx/i18n"
	hoststore "nextopsx/store/hosts"
	"nextopsx/ui/assets"
	"nextopsx/ui/workspace"
)

// Tree builds the folder/host tree and reports selection.
func Tree(c *ui.Context, a *workspace.State) {
	ui.Tree(c, func() {
		buildNodes(c, a, a.Roots)
	})
}

func buildNodes(c *ui.Context, a *workspace.State, nodes []hoststore.TreeNode) {
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
			item = item.Tooltip(i18n.Tf("hosts.tooltip.protocol", a.HostProtocolName(n.Protocol)))
			item.ContextMenu(func(m *ui.Menu) { hostMenu(m, a, n) })
		}
		if item.DoubleClicked() {
			a.OpenAsset(n)
		} else if item.Clicked() {
			a.Select(n)
		}
	}
}

func hostTreeIcon(c *ui.Context, n *hoststore.TreeNode) {
	if n.NodeType == "folder" {
		if s := assets.FolderSVG(); s != nil {
			ui.Icon(c, s).Size(15, 15).TextColor(c.Theme().TextMuted)
		}
		return
	}
	if b := assets.ProtocolBitmap(n.Protocol); b != nil {
		ui.Image(c, b).Size(16, 16)
	}
}

func folderMenu(m *ui.Menu, a *workspace.State, n *hoststore.TreeNode) {
	if m.Item(i18n.T("hosts.menu.newAssetHere")).Chosen() {
		a.Select(n)
		a.OpenNewHostForm()
	}
	if m.Item(i18n.T("hosts.menu.newFolderHere")).Chosen() {
		a.Select(n)
		a.OpenNewFolderForm()
	}
	m.Separator()
	if m.Item(i18n.T("hosts.menu.rename")).Chosen() {
		a.Select(n)
		a.OpenRenameDialog()
	}
	if m.Item(i18n.T("hosts.menu.move")).Chosen() {
		a.Select(n)
		a.OpenMoveDialog()
	}
	m.Separator()
	if m.Item(i18n.T("hosts.menu.delete")).Chosen() {
		a.Select(n)
		a.OpenDeleteDialog()
	}
}

func hostMenu(m *ui.Menu, a *workspace.State, n *hoststore.TreeNode) {
	if a.IsSSHProtocol(n.Protocol) {
		if m.Item(i18n.T("hosts.menu.openTerminal")).Chosen() {
			a.OpenAsset(n)
		}
	}
	if a.IsSQLProtocol(n.Protocol) {
		if m.Item(i18n.T("hosts.menu.connectDatabase")).Chosen() {
			a.OpenAsset(n)
		}
	}
	if a.IsRedisProtocol(n.Protocol) || a.IsESProtocol(n.Protocol) {
		if m.Item(i18n.T("hosts.menu.connect")).Chosen() {
			a.OpenAsset(n)
		}
	}
	if n.Protocol == "winrm" {
		if m.Item(i18n.T("hosts.menu.openWinRM")).Chosen() {
			a.OpenAsset(n)
		}
		if m.Item(i18n.T("shell.remoteDesktop")).Chosen() {
			a.Select(n)
			a.OpenRDP()
		}
	}
	if m.Item(i18n.T("hosts.menu.openAI")).Chosen() {
		a.Select(n)
		a.OpenAIDrawer()
	}
	m.Separator()
	if m.Item(i18n.T("hosts.menu.edit")).Chosen() {
		a.Select(n)
		a.OpenEditHostForm()
	}
	if m.Item(i18n.T("hosts.menu.move")).Chosen() {
		a.Select(n)
		a.OpenMoveDialog()
	}
	m.Separator()
	if m.Item(i18n.T("hosts.menu.delete")).Chosen() {
		a.Select(n)
		a.OpenDeleteDialog()
	}
}

// Count returns the number of host nodes in the tree.
func Count(nodes []hoststore.TreeNode) int {
	n := 0
	for _, node := range nodes {
		if node.NodeType == "host" {
			n++
		}
		n += Count(node.Children)
	}
	return n
}

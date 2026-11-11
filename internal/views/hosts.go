package views

import (
	"github.com/egoist/mygo/ui"

	"nextopsx/internal/app"
	hoststore "nextopsx/internal/store/hosts"
)

// HostTree builds the folder/host tree and reports selection.
func HostTree(c *ui.Context, a *app.App) {
	buildNodes(c, a, a.Roots)
}

func buildNodes(c *ui.Context, a *app.App, nodes []hoststore.TreeNode) {
	for i := range nodes {
		n := &nodes[i]
		if n.NodeType == "folder" {
			ui.TreeItem(c, n.Name, a.Open[n.ID], func() {
				buildNodes(c, a, n.Children)
			})
			continue
		}
		label := n.Name
		if n.Protocol != "" && n.Protocol != "ssh" {
			label = n.Name + "  (" + n.Protocol + ")"
		}
		selected := a.Selected != nil && a.Selected.ID == n.ID
		if ui.TreeItem(c, label, nil, nil).Selected(selected).Clicked() {
			a.Select(n)
		}
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

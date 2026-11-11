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
// welcome pane on the right.
func Shell(a *app.App) func(*ui.Context) {
	return func(c *ui.Context) {
		t := c.Theme()
		ui.Row(c).Fill().Children(func() {
			ui.Column(c).Width(280).FillHeight().Background(t.Surface).
				Border(1, t.Border).Children(func() {
				ui.Row(c).Padding(12).Gap(8).AlignItems(ui.Center).Children(func() {
					ui.Text(c, "Hosts").FontSize(16).Bold()
					ui.Text(c, fmt.Sprintf("%d", countHosts(a.Roots))).TextColor(t.TextMuted)
				})
				ui.Scroll(c).Grow(1).Children(func() {
					HostTree(c, a)
				})
			})

			ui.Column(c).Grow(1).FillHeight().Children(func() {
				if a.Term != nil {
					terminal.View(c, a.Term).Fill()
					return
				}
				Welcome(c, a)
			})
		})
	}
}

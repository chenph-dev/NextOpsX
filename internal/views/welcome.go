package views

import (
	"fmt"

	"github.com/egoist/mygo/ui"

	"nextopsx/internal/app"
)

// Welcome shows the selected asset and its connect action.
func Welcome(c *ui.Context, a *app.App) {
	t := c.Theme()
	ui.Column(c).Fill().Center().Gap(12).Padding(32).Children(func() {
		if a.Selected == nil {
			ui.Text(c, "选择左侧资产").FontSize(18).TextColor(t.TextMuted)
			return
		}
		n := a.Selected
		ui.Text(c, n.Name).FontSize(22).Bold()
		ui.Text(c, fmt.Sprintf("%s@%s:%d", n.User, n.Addr, n.Port)).
			TextColor(t.TextMuted)
		if n.Protocol != "" {
			ui.Text(c, "协议："+n.Protocol).TextColor(t.TextMuted)
		}
		if a.Busy {
			ui.Text(c, "正在连接…").TextColor(t.TextMuted)
		} else if ui.PrimaryButton(c, "连接").Clicked() {
			a.Connect(n)
		}
		if a.ErrMsg != "" {
			ui.Text(c, a.ErrMsg).TextColor(t.Danger).Width(520)
		}
	})
}

package winrm

import (
	"github.com/egoist/mygo/ui"

	"nextopsx/i18n"
	"nextopsx/ui/style"
	"nextopsx/ui/workspace"
)

// Panel is the WinRM PowerShell console for one Windows asset. Commands run
// one at a time through WinRM; the remote process does not persist between
// them, so this is a command console rather than a PTY terminal.
func Panel(c *ui.Context, a *workspace.State, t *ui.Theme) {
	ui.Column(c).Fill().MinWidth(0).Children(func() {
		if a.WinRM.Status == workspace.WinRMStatusFailed {
			ui.Row(c).Padding(6, 12).Gap(8).AlignItems(ui.Center).Children(func() {
				ui.Text(c, i18n.T("winrm.status.failed")).
					FontSize(style.FontSize(12)).TextColor(t.Danger)
				if a.WinRM.Info != "" {
					ui.Text(c, a.WinRM.Info).FontSize(style.FontSize(12)).
						TextColor(t.Danger).SingleLine()
				}
				ui.Button(c, i18n.T("winrm.retry")).FontSize(style.FontSize(11)).
					OnClick(a.WinRMProbe)
			})
			ui.Divider(c)
		}
		ui.Scroll(c).TrackScroll(&a.WinRM.Scroll).Grow(1).Children(func() {
			ui.Column(c).Padding(8, 12).Gap(2).Children(func() {
				if len(a.WinRM.Lines) == 0 {
					ui.Text(c, i18n.T("winrm.empty")).FontSize(style.FontSize(12)).
						TextColor(t.TextMuted)
					return
				}
				for _, line := range a.WinRM.Lines {
					color := t.Text
					prefix := ""
					switch line.Stream {
					case "cmd":
						color = t.Accent
						prefix = "PS> "
					case "stderr", "error":
						color = t.Danger
					case "exit":
						color = t.TextMuted
					}
					ui.Text(c, prefix+line.Text).Font("monospace").
						FontSize(style.FontSize(12)).TextColor(color).Wrap()
				}
			})
		})
		ui.Divider(c)
		ui.Row(c).Padding(6, 8).Gap(6).AlignItems(ui.Center).Children(func() {
			ui.Text(c, "PS>").Font("monospace").FontSize(style.FontSize(12)).
				TextColor(t.Accent)
			c.Key("winrm-command")
			ui.TextInput(c, &a.WinRM.Input).Grow(1).Font("monospace").
				FontSize(style.FontSize(12)).
				Placeholder(i18n.T("winrm.placeholder")).
				AutoFocus().
				OnSubmit(a.WinRMRunCommand)
			switch {
			case a.WinRM.Status == workspace.WinRMStatusConnecting:
				ui.Spinner(c)
				ui.Text(c, i18n.T("winrm.status.connecting")).
					FontSize(style.FontSize(11)).TextColor(t.TextMuted)
			case a.WinRM.Running:
				ui.Spinner(c)
				ui.Text(c, i18n.T("winrm.running")).
					FontSize(style.FontSize(11)).TextColor(t.TextMuted)
			}
		})
	})
}

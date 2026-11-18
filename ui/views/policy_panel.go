package views

import (
	"github.com/egoist/mygo/ui"

	"nextopsx/i18n"
	"nextopsx/ui/app"
)

// policySettings is the approval-policy section of the settings page.
func policySettings(c *ui.Context, a *app.App, t *ui.Theme) {
	ui.Column(c).Fill().MinWidth(0).Children(func() {
		ui.Row(c).Padding(8, 14).Gap(8).AlignItems(ui.Center).Children(func() {
			ui.Text(c, i18n.T("settings.policy.title")).FontSize(fontSize(13)).Bold()
			ui.Spacer(c)
			ui.Button(c, i18n.T("settings.policy.reload")).OnClick(a.LoadApprovalPolicy)
		})
		ui.Divider(c)
		if a.Policy.Loading {
			ui.Row(c).Grow(1).Center().Children(func() {
				ui.Spinner(c)
				ui.Text(c, i18n.T("settings.policy.loading")).FontSize(fontSize(12)).TextColor(t.TextMuted)
			})
			return
		}
		if !a.Policy.Loaded {
			ui.Column(c).Padding(16).Gap(6).Children(func() {
				msg := a.Policy.Err
				if msg == "" {
					msg = i18n.T("settings.policy.notLoaded")
				}
				dbErrorText(c, msg, t)
				ui.Button(c, i18n.T("settings.policy.retry")).OnClick(a.LoadApprovalPolicy)
			})
			return
		}
		ui.Scroll(c).Grow(1).Children(func() {
			ui.Row(c).Justify(ui.Center).Children(func() {
				ui.Column(c).Width(settingsContentWidth(c)).Padding(20, 28).Gap(16).
					Children(func() {
						settingsSectionTitle(c, t, i18n.T("settings.policy.commandApproval"))
						settingsCard(c, t, func() {
							settingsRow(c, i18n.T("settings.policy.autoAllow"), func() {
								ui.Switch(c, &a.Policy.EnableAuto)
							})
						})
						ui.Text(c, i18n.T("settings.policy.autoHint")).
							FontSize(fontSize(12)).TextColor(t.TextMuted).Wrap()

						settingsSectionTitle(c, t, i18n.T("settings.policy.whitelist"))
						settingsCard(c, t, func() {
							ui.Column(c).Padding(12, 16).Gap(8).Children(func() {
								c.Key("policy-whitelist")
								ui.TextArea(c, &a.Policy.ReadOnly).Height(220).
									Font("monospace").FontSize(fontSize(12))
								ui.Text(c, i18n.T("settings.policy.whitelistHint")).
									FontSize(fontSize(12)).TextColor(t.TextMuted).Wrap()
							})
						})

						if a.Policy.Err != "" {
							ui.Text(c, a.Policy.Err).FontSize(fontSize(13)).TextColor(t.Danger).Wrap()
						} else if a.Policy.Notice != "" {
							ui.Text(c, a.Policy.Notice).FontSize(fontSize(13)).TextColor(t.Success).Wrap()
						}
						ui.Row(c).Justify(ui.End).Gap(6).AlignItems(ui.Center).Children(func() {
							ui.Button(c, i18n.T("settings.policy.clear")).OnClick(func() {
								a.Policy.ReadOnly = ""
							})
							save := ui.PrimaryButton(c, i18n.T("settings.policy.save")).OnClick(a.SaveApprovalPolicy)
							if a.Policy.Saving {
								save = save.Disabled(true)
							}
						})
					})
			})
		})
	})
}

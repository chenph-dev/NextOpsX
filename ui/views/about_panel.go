package views

import (
	"github.com/egoist/mygo/ui"

	"nextopsx/i18n"
	"nextopsx/ui/app"
)

// aboutSettings is the settings page's About section: version and automatic
// updates.
func aboutSettings(c *ui.Context, a *app.App, t *ui.Theme) {
	about := a.Settings.About
	ui.Scroll(c).Grow(1).Children(func() {
		ui.Row(c).Justify(ui.Center).Children(func() {
			ui.Column(c).Width(settingsContentWidth(c)).Padding(20, 28).Gap(16).Children(func() {
				settingsSectionTitle(c, t, i18n.T("settings.about.title"))
				settingsCard(c, t, func() {
					settingsRow(c, i18n.T("settings.about.version"), func() {
						ui.Text(c, about.Version).FontSize(fontSize(13)).TextColor(t.TextMuted)
					})
					ui.Divider(c)
					settingsRow(c, i18n.T("settings.about.updates"), func() {
						ui.Row(c).Gap(8).AlignItems(ui.Center).Children(func() {
							if !about.Enabled {
								ui.Text(c, i18n.T("settings.about.disabled")).
									FontSize(fontSize(12)).TextColor(t.Warning)
							} else {
								aboutStatusText(c, about, t)
							}
							if about.Enabled && !about.Installed {
								check := ui.Button(c, i18n.T("settings.about.check"))
								if about.Checking || about.Installing {
									check = check.Disabled(true)
								}
								check.OnClick(a.CheckForUpdates)
							}
							if about.Available {
								install := ui.PrimaryButton(c, i18n.T("settings.about.install"))
								if about.Installing {
									install = install.Disabled(true)
								}
								install.OnClick(a.InstallUpdate)
							}
							if about.Installed {
								ui.PrimaryButton(c, i18n.T("settings.about.relaunch")).
									OnClick(a.RelaunchApp)
							}
						})
					})
				})

				if !about.Enabled {
					ui.Text(c, i18n.T("settings.about.disabledHint")).
						FontSize(fontSize(12)).TextColor(t.TextMuted).Wrap()
				}
				if about.Err != "" {
					ui.Text(c, about.Err).FontSize(fontSize(12)).TextColor(t.Danger).Wrap()
				}
				if about.Installing {
					ui.Column(c).Gap(6).Children(func() {
						ui.Text(c, i18n.T("settings.about.installing")).
							FontSize(fontSize(12)).TextColor(t.TextMuted)
						if pct := a.UpdatePercent(); pct >= 0 {
							ui.Progress(c, float64(pct)/100)
							ui.Text(c, i18n.Tf("settings.about.progress", pct)).
								FontSize(fontSize(12)).TextColor(t.TextMuted)
						}
					})
				}

				if about.Available && about.Notes != "" {
					settingsSectionTitle(c, t, i18n.T("settings.about.notes"))
					settingsCard(c, t, func() {
						ui.Column(c).Padding(12, 16).Children(func() {
							MarkdownView(c, about.Notes)
						})
					})
				}
			})
		})
	})
}

// aboutStatusText shows the update state next to the action buttons.
func aboutStatusText(c *ui.Context, about app.AboutState, t *ui.Theme) {
	switch {
	case about.Checking:
		ui.Spinner(c)
		ui.Text(c, i18n.T("settings.about.checking")).
			FontSize(fontSize(12)).TextColor(t.TextMuted)
	case about.Available:
		ui.Text(c, i18n.Tf("settings.about.available", about.Latest)).
			FontSize(fontSize(12)).TextColor(t.Accent)
	case about.Installed:
		ui.Text(c, i18n.T("settings.about.installed")).
			FontSize(fontSize(12)).TextColor(t.Success)
	case about.Checked && about.Err == "":
		ui.Text(c, i18n.T("settings.about.upToDate")).
			FontSize(fontSize(12)).TextColor(t.TextMuted)
	}
}

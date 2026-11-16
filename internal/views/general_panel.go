package views

import (
	"fmt"

	"github.com/egoist/mygo/ui"

	"nextopsx/internal/app"
	"nextopsx/internal/i18n"
	configstore "nextopsx/internal/store/config"
)

// generalSettings is the general section of the settings page: appearance
// and local data location.
func generalSettings(c *ui.Context, a *app.App, t *ui.Theme) {
	ui.Scroll(c).Grow(1).Children(func() {
		ui.Row(c).Justify(ui.Center).Children(func() {
			ui.Column(c).Width(settingsContentWidth(c)).Padding(20, 28).Gap(16).Children(func() {
				settingsSectionTitle(c, t, i18n.T("settings.general.appearance"))
				settingsCard(c, t, func() {
					settingsRow(c, i18n.T("settings.general.language"), func() {
						// Language names stay in their own language so the
						// choice is recognizable even when the UI is not.
						seg := ui.Segmented(c, &a.Settings.LanguageIndex,
							"简体中文", "English")
						if seg.Changed() {
							a.SetLanguage(string(i18n.LangAt(a.Settings.LanguageIndex)))
						}
					})
					ui.Divider(c)
					settingsRow(c, i18n.T("settings.general.theme"), func() {
						ui.Row(c).Gap(8).AlignItems(ui.Center).Children(func() {
							if a.Settings.UISaving {
								ui.Spinner(c)
							}
							seg := ui.Segmented(c, &a.Settings.ThemeIndex,
								i18n.T("settings.theme.system"),
								i18n.T("settings.theme.light"),
								i18n.T("settings.theme.dark"))
							if seg.Changed() {
								a.SetTheme(app.ThemeAt(a.Settings.ThemeIndex))
							}
						})
					})
					ui.Divider(c)
					settingsRow(c, i18n.T("settings.general.scale"), func() {
						labels := make([]string, len(configstore.Scales))
						for i, s := range configstore.Scales {
							labels[i] = fmt.Sprintf("%d%%", s)
						}
						seg := ui.Segmented(c, &a.Settings.ScaleIndex, labels...)
						if seg.Changed() {
							a.SetScale(app.ScaleAt(a.Settings.ScaleIndex))
						}
					})
				})
				ui.Text(c, i18n.T("settings.general.hint")).
					FontSize(fontSize(12)).TextColor(t.TextMuted).Wrap()
				if a.Settings.UIErr != "" {
					ui.Text(c, a.Settings.UIErr).FontSize(fontSize(12)).TextColor(t.Danger).Wrap()
				}

				settingsSectionTitle(c, t, i18n.T("settings.general.data"))
				settingsCard(c, t, func() {
					ui.Column(c).Padding(12, 16).Gap(4).Children(func() {
						ui.Text(c, i18n.T("settings.general.dataDir")).FontSize(fontSize(13))
						ui.Text(c, a.Settings.DataDir).Font("monospace").FontSize(fontSize(12)).
							TextColor(t.TextMuted).Wrap()
						ui.Text(c, i18n.T("settings.general.dataHint")).
							FontSize(fontSize(12)).TextColor(t.TextMuted).Wrap()
					})
				})
			})
		})
	})
}

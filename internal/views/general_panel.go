package views

import (
	"fmt"

	"github.com/egoist/mygo/ui"

	"nextopsx/internal/app"
	configstore "nextopsx/internal/store/config"
)

// generalSettings is the general section of the settings page: appearance
// and local data location.
func generalSettings(c *ui.Context, a *app.App, t *ui.Theme) {
	ui.Scroll(c).Grow(1).Children(func() {
		ui.Row(c).Justify(ui.Center).Children(func() {
			ui.Column(c).Width(settingsContentWidth(c)).Padding(20, 28).Gap(16).Children(func() {
				settingsSectionTitle(c, t, "外观")
				settingsCard(c, t, func() {
					settingsRow(c, "界面语言", func() {
						ui.Text(c, "简体中文").FontSize(fontSize(13)).TextColor(t.TextMuted)
					})
					ui.Divider(c)
					settingsRow(c, "主题", func() {
						ui.Row(c).Gap(8).AlignItems(ui.Center).Children(func() {
							if a.Settings.UISaving {
								ui.Spinner(c)
							}
							seg := ui.Segmented(c, &a.Settings.ThemeIndex,
								"跟随系统", "浅色", "深色")
							if seg.Changed() {
								a.SetTheme(app.ThemeAt(a.Settings.ThemeIndex))
							}
						})
					})
					ui.Divider(c)
					settingsRow(c, "界面缩放", func() {
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
				ui.Text(c, "默认跟随系统外观；界面缩放会同时放大文字与控件（90%–125%）。").
					FontSize(fontSize(12)).TextColor(t.TextMuted).Wrap()
				if a.Settings.UIErr != "" {
					ui.Text(c, a.Settings.UIErr).FontSize(fontSize(12)).TextColor(t.Danger).Wrap()
				}

				settingsSectionTitle(c, t, "数据")
				settingsCard(c, t, func() {
					ui.Column(c).Padding(12, 16).Gap(4).Children(func() {
						ui.Text(c, "数据目录").FontSize(fontSize(13))
						ui.Text(c, a.Settings.DataDir).Font("monospace").FontSize(fontSize(12)).
							TextColor(t.TextMuted).Wrap()
						ui.Text(c, "数据库、技能文件与密钥保存在此目录。").
							FontSize(fontSize(12)).TextColor(t.TextMuted).Wrap()
					})
				})
			})
		})
	})
}

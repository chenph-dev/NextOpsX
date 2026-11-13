package views

import (
	"github.com/egoist/mygo/ui"

	"nextopsx/internal/app"
	configstore "nextopsx/internal/store/config"
)

// generalSettings is the general section of the settings page: appearance
// and local data location.
func generalSettings(c *ui.Context, a *app.App, t *ui.Theme) {
	ui.Column(c).Fill().MinWidth(0).Children(func() {
		settingsHeader(c, t, "常规", "界面外观与本地数据位置。", nil)
		ui.Scroll(c).Grow(1).Children(func() {
			ui.Column(c).Padding(16, 20).Gap(14).MaxWidth(720).Children(func() {
				ui.Row(c).Gap(10).AlignItems(ui.Center).Children(func() {
					ui.Text(c, "主题").Width(72).FontSize(12).Bold()
					themeSelect(c, a, t)
					if a.Settings.ThemeSaving {
						ui.Spinner(c)
					}
				})
				ui.Text(c, "默认跟随系统外观；选择浅色或深色后固定使用。").
					FontSize(11).TextColor(t.TextMuted).Wrap()
				if a.Settings.ThemeErr != "" {
					ui.Text(c, a.Settings.ThemeErr).FontSize(12).TextColor(t.Danger).Wrap()
				}

				ui.Divider(c)

				ui.Text(c, "数据目录").FontSize(12).Bold()
				ui.Text(c, a.Settings.DataDir).Font("monospace").FontSize(11).
					TextColor(t.TextMuted).Wrap()
				ui.Text(c, "数据库、技能文件与密钥保存在此目录。").
					FontSize(11).TextColor(t.TextMuted).Wrap()
			})
		})
	})
}

func themeSelect(c *ui.Context, a *app.App, t *ui.Theme) {
	sel := ui.SelectBase(c, &a.Settings.Theme)
	sel.Trigger.Padding(4, 8).Border(1, t.Border).Radius(4).Children(func() {
		ui.Text(c, themeLabel(a.Settings.Theme)).FontSize(11)
	})
	if sel.Trigger.Changed() {
		a.SetTheme(a.Settings.Theme)
	}
	sel.Popup(func(panel ui.Element) {
		panel.Padding(4).Background(t.Surface).Border(1, t.Border).Radius(6)
		for _, theme := range configstore.Themes {
			item := sel.Item(theme).Padding(5, 9)
			if item.Highlighted() {
				item = item.Background(t.SurfaceHover)
			}
			item.Children(func() {
				ui.Text(c, themeLabel(theme)).FontSize(11)
			})
		}
	})
}

func themeLabel(theme string) string {
	switch theme {
	case configstore.ThemeLight:
		return "浅色"
	case configstore.ThemeDark:
		return "深色"
	default:
		return "跟随系统"
	}
}

package views

import (
	"github.com/egoist/mygo/ui"

	"nextopsx/internal/app"
	"nextopsx/internal/assets"
)

// SettingsView is the full-window settings page: a section nav on the left
// and the selected section's content on the right.
func SettingsView(c *ui.Context, a *app.App) {
	t := c.Theme()
	ui.Row(c).Fill().Children(func() {
		settingsNav(c, a, t)
		ui.Column(c).Grow(1).FillHeight().MinWidth(0).Children(func() {
			settingsPageHeader(c, a, t)
			switch a.Settings.Tab {
			case app.SettingsGeneral:
				generalSettings(c, a, t)
			case app.SettingsSkills:
				skillsSettings(c, a, t)
			case app.SettingsPolicy:
				policySettings(c, a, t)
			case app.SettingsAudit:
				auditSettings(c, a, t)
			default:
				modelSettings(c, a, t)
			}
		})
	})
}

const settingsNavWidth float32 = 240

// settingsContentWidth keeps a card column readable and centered without
// overflowing narrow windows.
func settingsContentWidth(c *ui.Context) float32 {
	w, _ := c.Size()
	width := w - settingsNavWidth - 56
	if width > 880 {
		width = 880
	}
	if width < 320 {
		width = 320
	}
	return width
}

func settingsNav(c *ui.Context, a *app.App, t *ui.Theme) {
	ui.Column(c).Width(settingsNavWidth).FillHeight().Background(t.Surface).
		Border(1, t.Border).Children(func() {
		ui.Row(c).Padding(14, 16).Gap(8).AlignItems(ui.Center).Children(func() {
			if s := assets.ObjectSVG("sliders"); s != nil {
				ui.Icon(c, s).Size(16, 16)
			}
			ui.Text(c, "NextOpsX").FontSize(fontSize(14)).Bold()
		})
		ui.Row(c).Padding(6, 10).Children(func() {
			back := ui.ButtonBase(c).Padding(5, 8).Radius(6).Gap(6).
				AlignItems(ui.Center).OnClick(a.CloseSettings)
			back.Children(func() {
				if s := assets.ObjectSVG("back"); s != nil {
					ui.Icon(c, s).Size(13, 13)
				}
				ui.Text(c, "返回").FontSize(fontSize(13)).TextColor(t.TextMuted)
			})
		})
		ui.Row(c).Padding(14, 8).Children(func() {
			ui.Text(c, "设置").FontSize(fontSize(11)).TextColor(t.TextMuted)
		})
		ui.Column(c).Padding(6).Gap(2).Children(func() {
			settingsNavItem(c, a, t, app.SettingsGeneral, "通用", "sliders")
			settingsNavItem(c, a, t, app.SettingsModel, "模型", "cpu")
			settingsNavItem(c, a, t, app.SettingsSkills, "技能", "spark")
			settingsNavItem(c, a, t, app.SettingsPolicy, "审批", "shield")
			settingsNavItem(c, a, t, app.SettingsAudit, "审计", "chart")
		})
		ui.Spacer(c)
		ui.Divider(c)
		ui.Row(c).Padding(14, 10).Children(func() {
			ui.Text(c, "NextOpsX").FontSize(fontSize(11)).TextColor(t.TextMuted)
		})
	})
}

func settingsNavItem(c *ui.Context, a *app.App, t *ui.Theme, tab app.SettingsTab, label, icon string) {
	selected := a.Settings.Tab == tab
	row := ui.Row(c).Padding(8, 10).Radius(6).Gap(8).AlignItems(ui.Center).OnClick(func() {
		// 点击当前分区不应重新加载，否则会丢掉未保存的编辑。
		if a.Settings.Tab != tab {
			a.OpenSettings(tab)
		}
	})
	if selected {
		row = row.Background(t.SurfaceHover)
	}
	row.Children(func() {
		color := t.TextMuted
		if selected {
			color = t.Accent
		}
		if s := assets.ObjectSVG(icon); s != nil {
			ui.Icon(c, s).Size(14, 14).TextColor(color)
		}
		ui.Text(c, label).FontSize(fontSize(13)).TextColor(color)
	})
}

func settingsPageHeader(c *ui.Context, a *app.App, t *ui.Theme) {
	title, icon := settingsSectionMeta(a.Settings.Tab)
	ui.Row(c).Padding(14, 18).Gap(10).AlignItems(ui.Center).Children(func() {
		if s := assets.ObjectSVG(icon); s != nil {
			ui.Icon(c, s).Size(16, 16)
		}
		ui.Text(c, title).FontSize(fontSize(16)).Bold()
	})
	ui.Divider(c)
}

func settingsSectionMeta(tab app.SettingsTab) (title, icon string) {
	switch tab {
	case app.SettingsGeneral:
		return "通用", "sliders"
	case app.SettingsSkills:
		return "技能", "spark"
	case app.SettingsPolicy:
		return "审批", "shield"
	case app.SettingsAudit:
		return "审计", "chart"
	default:
		return "模型", "cpu"
	}
}

// settingsCard is a rounded group of settings rows.
func settingsCard(c *ui.Context, t *ui.Theme, fn func()) {
	ui.Column(c).Background(t.Surface).Border(1, t.Border).Radius(8).Children(fn)
}

// settingsRow puts a label on the left and its control on the right.
func settingsRow(c *ui.Context, label string, control func()) {
	ui.Row(c).Padding(12, 16).Gap(12).AlignItems(ui.Center).Children(func() {
		ui.Text(c, label).FontSize(fontSize(13))
		ui.Spacer(c)
		control()
	})
}

func settingsSectionTitle(c *ui.Context, t *ui.Theme, title string) {
	ui.Text(c, title).FontSize(fontSize(13)).Bold().TextColor(t.TextMuted)
}

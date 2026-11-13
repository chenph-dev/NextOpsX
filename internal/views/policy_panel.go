package views

import (
	"github.com/egoist/mygo/ui"

	"nextopsx/internal/app"
)

// policySettings is the approval-policy section of the settings page.
func policySettings(c *ui.Context, a *app.App, t *ui.Theme) {
	ui.Column(c).Fill().MinWidth(0).Children(func() {
		ui.Row(c).Padding(8, 14).Gap(8).AlignItems(ui.Center).Children(func() {
			ui.Text(c, "审批策略").FontSize(fontSize(13)).Bold()
			ui.Spacer(c)
			ui.Button(c, "重新加载").OnClick(a.LoadApprovalPolicy)
		})
		ui.Divider(c)
		if a.Policy.Loading {
			ui.Row(c).Grow(1).Center().Children(func() {
				ui.Spinner(c)
				ui.Text(c, "读取审批策略…").FontSize(fontSize(12)).TextColor(t.TextMuted)
			})
			return
		}
		if !a.Policy.Loaded {
			ui.Column(c).Padding(16).Gap(6).Children(func() {
				msg := a.Policy.Err
				if msg == "" {
					msg = "审批策略尚未加载"
				}
				dbErrorText(c, msg, t)
				ui.Button(c, "重试").OnClick(a.LoadApprovalPolicy)
			})
			return
		}
		ui.Scroll(c).Grow(1).Children(func() {
			ui.Row(c).Justify(ui.Center).Children(func() {
				ui.Column(c).Width(settingsContentWidth(c)).Padding(20, 28).Gap(16).
					Children(func() {
						settingsSectionTitle(c, t, "命令审批")
						settingsCard(c, t, func() {
							settingsRow(c, "自动放行只读命令", func() {
								ui.Switch(c, &a.Policy.EnableAuto)
							})
						})
						ui.Text(c, "开启后，命中只读白名单的命令直接执行；未命中的命令仍需人工审批。资产级覆盖在资产编辑中设置。").
							FontSize(fontSize(12)).TextColor(t.TextMuted).Wrap()

						settingsSectionTitle(c, t, "只读白名单")
						settingsCard(c, t, func() {
							ui.Column(c).Padding(12, 16).Gap(8).Children(func() {
								c.Key("policy-whitelist")
								ui.TextArea(c, &a.Policy.ReadOnly).Height(220).
									Font("monospace").FontSize(fontSize(12))
								ui.Text(c, "每行写一个命令名（如 df、free），不要写完整命令行；带参数的行不会命中。含管道、重定向、命令替换或命令分隔符时不会自动放行。").
									FontSize(fontSize(12)).TextColor(t.TextMuted).Wrap()
							})
						})

						if a.Policy.Err != "" {
							ui.Text(c, a.Policy.Err).FontSize(fontSize(13)).TextColor(t.Danger).Wrap()
						} else if a.Policy.Notice != "" {
							ui.Text(c, a.Policy.Notice).FontSize(fontSize(13)).TextColor(t.Success).Wrap()
						}
						ui.Row(c).Justify(ui.End).Gap(6).AlignItems(ui.Center).Children(func() {
							ui.Button(c, "清空（使用内置默认）").OnClick(func() {
								a.Policy.ReadOnly = ""
							})
							save := ui.PrimaryButton(c, "保存").OnClick(a.SaveApprovalPolicy)
							if a.Policy.Saving {
								save = save.Disabled(true)
							}
						})
					})
			})
		})
	})
}

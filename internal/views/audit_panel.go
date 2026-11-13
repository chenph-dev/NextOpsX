package views

import (
	"fmt"
	"time"

	"github.com/egoist/mygo/ui"

	"nextopsx/internal/app"
)

// auditSettings is the AI call audit section of the settings page.
func auditSettings(c *ui.Context, a *app.App, t *ui.Theme) {
	ui.Column(c).Fill().MinWidth(0).Children(func() {
		settingsHeader(c, t, "审计", "AI 模型与工具的调用记录，用于排查与 token 用量统计。", func() {
			ui.Button(c, "刷新").OnClick(a.LoadAudit)
		})
		switch {
		case a.Audit.Loading && len(a.Audit.Logs) == 0:
			ui.Row(c).Grow(1).Center().Children(func() {
				ui.Spinner(c)
				ui.Text(c, "读取审计日志…").FontSize(11).TextColor(t.TextMuted)
			})
		case a.Audit.Err != "" && len(a.Audit.Logs) == 0:
			ui.Column(c).Padding(16).Gap(6).Children(func() {
				dbErrorText(c, a.Audit.Err, t)
				ui.Button(c, "重试").OnClick(a.LoadAudit)
			})
		default:
			if a.Audit.Err != "" {
				ui.Text(c, a.Audit.Err).FontSize(11).TextColor(t.Danger).
					Padding(4, 16).Wrap()
			}
			auditSummary(c, a, t)
			ui.Divider(c)
			if len(a.Audit.Logs) == 0 {
				ui.Column(c).Padding(16).Gap(6).Children(func() {
					ui.Text(c, "还没有调用记录").FontSize(12).TextColor(t.TextMuted)
					ui.Text(c, "使用 AI 对话后，模型与工具调用会记录在这里。").
						FontSize(11).TextColor(t.TextMuted)
				})
				return
			}
			auditTable(c, a, t)
			auditDetail(c, a, t)
		}
	})
}

func auditSummary(c *ui.Context, a *app.App, t *ui.Theme) {
	s := a.Audit.Summary
	ui.Row(c).Padding(10, 16).Gap(20).AlignItems(ui.Center).Children(func() {
		auditStat(c, t, "调用", fmt.Sprintf("%d", s.TotalCalls))
		auditStat(c, t, "模型", fmt.Sprintf("%d", s.ModelCalls))
		auditStat(c, t, "工具", fmt.Sprintf("%d", s.ToolCalls))
		auditStat(c, t, "Tokens 入", fmt.Sprintf("%d", s.TokensIn))
		auditStat(c, t, "Tokens 出", fmt.Sprintf("%d", s.TokensOut))
		auditStat(c, t, "Tokens 合计", fmt.Sprintf("%d", s.TokensTotal))
	})
}

func auditStat(c *ui.Context, t *ui.Theme, label, value string) {
	ui.Column(c).Gap(1).Children(func() {
		ui.Text(c, value).FontSize(15).Bold().SingleLine()
		ui.Text(c, label).FontSize(10).TextColor(t.TextMuted).SingleLine()
	})
}

func auditTable(c *ui.Context, a *app.App, t *ui.Theme) {
	cols := []ui.TableColumn{
		{Title: "时间", Width: 110},
		{Title: "组件", Width: 64},
		{Title: "名称", Width: 130},
		{Title: "实现", Width: 110},
		{Title: "入", Width: 72, Align: ui.End},
		{Title: "出", Width: 72, Align: ui.End},
		{Title: "合计", Width: 82, Align: ui.End},
		{Title: "耗时", Width: 80, Align: ui.End},
		{Title: "状态", Width: 64},
	}
	a.Audit.List.Selected = &a.Audit.Selected
	// 以记录 ID 作为行标识，刷新后选中项跟随记录而不是下标。
	a.Audit.List.Key = func(row int) any {
		if row >= 0 && row < len(a.Audit.Logs) {
			return a.Audit.Logs[row].ID
		}
		return row
	}
	ui.Table(c, &a.Audit.List, cols, len(a.Audit.Logs), func(row, col int) {
		if row < 0 || row >= len(a.Audit.Logs) {
			return
		}
		l := a.Audit.Logs[row]
		switch col {
		case 0:
			ui.Text(c, time.Unix(l.Ts, 0).Format("01-02 15:04:05")).
				FontSize(11).SingleLine()
		case 1:
			component := "其他"
			switch l.Component {
			case "model":
				component = "模型"
			case "tool":
				component = "工具"
			}
			ui.Text(c, component).FontSize(11).SingleLine()
		case 2:
			ui.Text(c, l.Name).FontSize(11).SingleLine()
		case 3:
			ui.Text(c, l.Provider).FontSize(11).SingleLine()
		case 4:
			ui.Text(c, fmt.Sprintf("%d", l.TokensIn)).FontSize(11).SingleLine()
		case 5:
			ui.Text(c, fmt.Sprintf("%d", l.TokensOut)).FontSize(11).SingleLine()
		case 6:
			ui.Text(c, fmt.Sprintf("%d", l.TokensTotal)).FontSize(11).SingleLine()
		case 7:
			ui.Text(c, fmt.Sprintf("%dms", l.DurationMS)).FontSize(11).SingleLine()
		case 8:
			if l.OK {
				ui.Text(c, "成功").FontSize(11).TextColor(t.Success).SingleLine()
			} else {
				ui.Text(c, "失败").FontSize(11).TextColor(t.Danger).SingleLine()
			}
		}
	}).Grow(1).MinWidth(0)
}

func auditDetail(c *ui.Context, a *app.App, t *ui.Theme) {
	if a.Audit.Selected < 0 || a.Audit.Selected >= len(a.Audit.Logs) {
		return
	}
	l := a.Audit.Logs[a.Audit.Selected]
	ui.Divider(c)
	ui.Column(c).Padding(6, 16).Gap(2).Children(func() {
		line := fmt.Sprintf("会话 %s · %s",
			l.SessionID, time.Unix(l.Ts, 0).Format("2006-01-02 15:04:05"))
		if l.SessionID == "" {
			line = time.Unix(l.Ts, 0).Format("2006-01-02 15:04:05")
		}
		ui.Text(c, line).FontSize(11).TextColor(t.TextMuted).SingleLine()
		if l.Error != "" {
			ui.Text(c, "错误："+l.Error).FontSize(11).TextColor(t.Danger).Wrap()
		}
	})
}

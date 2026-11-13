package views

import (
	"fmt"

	"github.com/egoist/mygo/ui"

	"nextopsx/internal/app"
)

func aiApprovalLabel(status string, t *ui.Theme) (string, ui.Color) {
	switch status {
	case app.AIStatusApproved:
		return "已批准", t.Success
	case app.AIStatusRejected:
		return "已拒绝", t.Danger
	case app.AIStatusAuto:
		return "自动放行", t.Accent
	default:
		return "待审批", t.Warning
	}
}

// commandCard renders one command proposal. history cards only replay the
// persisted approval result; live cards offer edit, approve and reject.
func commandCard(c *ui.Context, a *app.App, cmd *app.CommandSuggestion,
	draft, status string, history bool, t *ui.Theme) {
	highRisk := cmd.AssessedRisk == "high" || cmd.Risk == "high"
	border := t.Border
	if highRisk {
		border = t.Danger
	}
	display := draft
	if display == "" {
		display = cmd.Command
	}

	ui.Column(c).Border(1, border).Radius(6).Padding(8, 10).Gap(6).
		Background(t.Surface).Children(func() {
		ui.Row(c).Gap(8).AlignItems(ui.Center).Children(func() {
			label, color := aiApprovalLabel(status, t)
			ui.Text(c, label).FontSize(fontSize(12)).TextColor(color)
			why := cmd.Why
			if why == "" {
				why = "—"
			}
			ui.Text(c, "原因："+why).FontSize(fontSize(12)).TextColor(t.TextMuted).SingleLine()
			ui.Spacer(c)
			if highRisk {
				ui.Text(c, "高风险").FontSize(fontSize(12)).TextColor(t.Danger)
			}
		})

		if !history && a.AI.CommandEditing {
			c.Key("ai-command-draft")
			ui.TextArea(c, &a.AI.CommandDraft).Lines(1, 5).Font("monospace").
				FontSize(fontSize(12))
		} else {
			ui.Text(c, "$ "+display).Font("monospace").FontSize(fontSize(12)).Wrap()
		}

		if history || status != app.AIStatusPending {
			return
		}
		if a.AI.CommandConfirm {
			ui.Text(c, "这是高风险命令，请再次点击确认。").
				FontSize(fontSize(12)).TextColor(t.Danger)
		}
		ui.Row(c).Justify(ui.End).Gap(6).Children(func() {
			editLabel := "编辑"
			if a.AI.CommandEditing {
				editLabel = "完成"
			}
			ui.Button(c, editLabel).FontSize(fontSize(12)).OnClick(func() {
				a.AI.CommandEditing = !a.AI.CommandEditing
				a.AI.CommandConfirm = false
			})
			ui.Button(c, "拒绝").FontSize(fontSize(12)).OnClick(func() {
				a.RejectAICommand()
			})
			approveLabel := "批准"
			if highRisk && a.AI.CommandConfirm {
				approveLabel = "再次确认批准"
			}
			ui.PrimaryButton(c, approveLabel).FontSize(fontSize(12)).OnClick(func() {
				if highRisk && !a.AI.CommandConfirm {
					a.AI.CommandConfirm = true
					return
				}
				a.AI.CommandConfirm = false
				a.ApproveAICommand()
			})
		})
	})
}

// planCard renders an execution plan for approval (live) or replay (history).
func planCard(c *ui.Context, a *app.App, plan *app.PlanSuggestion, status string, history bool, t *ui.Theme) {
	ui.Column(c).Border(1, t.Accent).Radius(6).Padding(8, 10).Gap(6).
		Background(t.Surface).Children(func() {
		ui.Row(c).Gap(8).AlignItems(ui.Center).Children(func() {
			ui.Text(c, "执行计划").FontSize(fontSize(13)).Bold()
			label, color := aiApprovalLabel(status, t)
			ui.Text(c, label).FontSize(fontSize(12)).TextColor(color)
		})
		if plan.Goal != "" {
			ui.Text(c, "目标："+plan.Goal).FontSize(fontSize(13)).TextColor(t.TextMuted).Wrap()
		}
		for i, step := range plan.Steps {
			ui.Text(c, fmt.Sprintf("%d. %s", i+1, step)).FontSize(fontSize(13)).Wrap()
		}
		if history || status != app.AIStatusPending {
			return
		}
		ui.Row(c).Justify(ui.End).Gap(6).Children(func() {
			ui.Button(c, "拒绝").FontSize(fontSize(12)).OnClick(a.RejectAIPlan)
			ui.PrimaryButton(c, "批准").FontSize(fontSize(12)).OnClick(a.ApproveAIPlan)
		})
	})
}

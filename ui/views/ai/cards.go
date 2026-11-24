package ai

import (
	"fmt"

	"github.com/egoist/mygo/ui"

	"nextopsx/i18n"
	"nextopsx/ui/style"
	"nextopsx/ui/workspace"
)

func aiApprovalLabel(status string, t *ui.Theme) (string, ui.Color) {
	switch status {
	case workspace.AIStatusApproved:
		return i18n.T("ai.status.approved"), t.Success
	case workspace.AIStatusRejected:
		return i18n.T("ai.status.rejected"), t.Danger
	case workspace.AIStatusAuto:
		return i18n.T("ai.status.auto"), t.Accent
	case workspace.AIStatusPlan:
		return i18n.T("ai.status.plan"), t.Accent
	default:
		return i18n.T("ai.status.pending"), t.Warning
	}
}

// commandCard renders one command proposal. history cards only replay the
// persisted approval result; live cards offer edit, approve and reject.
func commandCard(c *ui.Context, a *workspace.State, cmd *workspace.CommandSuggestion,
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
			ui.Text(c, label).FontSize(style.FontSize(12)).TextColor(color)
			why := cmd.Why
			if why == "" {
				why = "—"
			}
			ui.Text(c, i18n.Tf("ai.command.why", why)).FontSize(style.FontSize(12)).TextColor(t.TextMuted).SingleLine()
			ui.Spacer(c)
			if highRisk {
				ui.Text(c, i18n.T("ai.command.highRisk")).FontSize(style.FontSize(12)).TextColor(t.Danger)
			}
		})

		if !history && a.AI.CommandEditing {
			c.Key("ai-command-draft")
			ui.TextArea(c, &a.AI.CommandDraft).Lines(1, 5).Font("monospace").
				FontSize(style.FontSize(12))
		} else {
			ui.Text(c, "$ "+display).Font("monospace").FontSize(style.FontSize(12)).Wrap()
		}

		if history || status != workspace.AIStatusPending {
			return
		}
		if a.AI.CommandConfirm {
			ui.Text(c, i18n.T("ai.command.confirmHint")).
				FontSize(style.FontSize(12)).TextColor(t.Danger)
		}
		ui.Row(c).Justify(ui.End).Gap(6).Children(func() {
			editLabel := i18n.T("ai.command.edit")
			if a.AI.CommandEditing {
				editLabel = i18n.T("ai.command.done")
			}
			ui.Button(c, editLabel).FontSize(style.FontSize(12)).OnClick(func() {
				a.AI.CommandEditing = !a.AI.CommandEditing
				a.AI.CommandConfirm = false
			})
			ui.Button(c, i18n.T("ai.command.reject")).FontSize(style.FontSize(12)).OnClick(func() {
				a.RejectAICommand()
			})
			approveLabel := i18n.T("ai.command.approve")
			if highRisk && a.AI.CommandConfirm {
				approveLabel = i18n.T("ai.command.approveAgain")
			}
			ui.PrimaryButton(c, approveLabel).FontSize(style.FontSize(12)).OnClick(func() {
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
func planCard(c *ui.Context, a *workspace.State, plan *workspace.PlanSuggestion, status string, history bool, t *ui.Theme) {
	ui.Column(c).Border(1, t.Accent).Radius(6).Padding(8, 10).Gap(6).
		Background(t.Surface).Children(func() {
		ui.Row(c).Gap(8).AlignItems(ui.Center).Children(func() {
			ui.Text(c, i18n.T("ai.plan.title")).FontSize(style.FontSize(13)).Bold()
			label, color := aiApprovalLabel(status, t)
			ui.Text(c, label).FontSize(style.FontSize(12)).TextColor(color)
		})
		if plan.Goal != "" {
			ui.Text(c, i18n.Tf("ai.plan.goal", plan.Goal)).FontSize(style.FontSize(13)).TextColor(t.TextMuted).Wrap()
		}
		for i, step := range plan.Steps {
			ui.Text(c, fmt.Sprintf("%d. %s", i+1, step)).FontSize(style.FontSize(13)).Wrap()
		}
		if !history && status == workspace.AIStatusPending {
			ui.Text(c, i18n.T("ai.plan.autoHint")).FontSize(style.FontSize(12)).
				TextColor(t.TextMuted).Wrap()
		}
		if history || status != workspace.AIStatusPending {
			return
		}
		ui.Row(c).Justify(ui.End).Gap(6).Children(func() {
			ui.Button(c, i18n.T("ai.command.reject")).FontSize(style.FontSize(12)).OnClick(a.RejectAIPlan)
			ui.PrimaryButton(c, i18n.T("ai.plan.approve")).FontSize(style.FontSize(12)).OnClick(a.ApproveAIPlan)
		})
	})
}

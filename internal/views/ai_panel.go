package views

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/egoist/mygo/ui"

	"nextopsx/internal/app"
	"nextopsx/internal/einoagent/session"
	"nextopsx/internal/i18n"
)

// AIPanel is the native AI chat: header, message stream, composer.
func AIPanel(c *ui.Context, a *app.App) {
	t := c.Theme()
	ui.Column(c).Fill().Children(func() {
		aiHeader(c, a, t)
		if a.AI.Loading {
			ui.Row(c).Padding(24).Center().Children(func() {
				ui.Spinner(c)
				ui.Text(c, i18n.T("ai.loading")).TextColor(t.TextMuted)
			})
			return
		}
		aiMessageStream(c, a, t)
		aiComposer(c, a, t)
	})
}

func aiHeader(c *ui.Context, a *app.App, t *ui.Theme) {
	ui.Row(c).Padding(8, 12).Gap(8).AlignItems(ui.Center).Children(func() {
		if label, color := aiStateBadge(a, t); label != "" {
			ui.Text(c, label).FontSize(fontSize(12)).TextColor(color)
		}
		// 会话加载期间供应商信息尚未就绪，先不展示，避免显示上一个资产的数据。
		if !a.AI.Loading {
			if a.AI.Configured {
				aiProviderSelect(c, a, t)
				if a.AI.Model != "" {
					ui.Text(c, a.AI.Model).FontSize(fontSize(12)).TextColor(t.TextMuted)
				}
			} else {
				ui.Text(c, i18n.T("ai.header.unconfigured")).FontSize(fontSize(12)).TextColor(t.Warning)
			}
		}
		ui.Spacer(c)

		if a.AI.Busy() {
			ui.Button(c, i18n.T("ai.action.cancel")).OnClick(a.CancelAIRun)
		}
		ui.Button(c, i18n.T("ai.action.newConversation")).OnClick(a.NewAIConversation)

		history := ui.Button(c, i18n.T("ai.action.history"))
		history = history.OnClick(func() { a.AI.HistoryOpen = !a.AI.HistoryOpen })
		ui.Popover(c, history, &a.AI.HistoryOpen, func() {
			aiHistoryPanel(c, a, t)
		})

		clear := ui.Button(c, i18n.T("ai.action.clear"))
		if a.AI.Busy() || a.AI.State == session.StateAwaitingApproval || len(a.AI.Messages) == 0 {
			clear = clear.Disabled(true)
		}
		clear.OnClick(a.ClearAIConversation)

		ui.Button(c, i18n.T("ai.action.settings")).OnClick(func() { a.OpenSettings(app.SettingsModel) })
	})
	ui.Divider(c)
}

// aiProviderSelect lets the user pick the provider used by this conversation.
// The next message takes effect; while a turn runs or an approval is pending
// the choice is disabled and the previous provider stays selected.
func aiProviderSelect(c *ui.Context, a *app.App, t *ui.Theme) {
	if len(a.AI.Providers) == 0 || a.AI.ProviderID == "" {
		return
	}
	sel := ui.SelectBase(c, &a.AI.ProviderChoice)
	trigger := sel.Trigger.Padding(3, 8).Border(1, t.Border).Radius(4)
	if a.AI.Busy() || a.AI.State == session.StateAwaitingApproval {
		trigger = trigger.Disabled(true)
	}
	trigger.Children(func() {
		ui.Text(c, a.AI.ProviderName).FontSize(fontSize(12)).SingleLine()
	})
	if sel.Trigger.Changed() {
		a.SwitchAIProvider(a.AI.ProviderChoice)
	}
	sel.Popup(func(panel ui.Element) {
		panel.Padding(4).Background(t.Surface).Border(1, t.Border).Radius(6)
		for _, p := range a.AI.Providers {
			item := sel.Item(p.ID).Padding(5, 9)
			if item.Highlighted() {
				item = item.Background(t.SurfaceHover)
			}
			item.Children(func() {
				ui.Text(c, p.Name).FontSize(fontSize(12))
			})
		}
	})
}

func aiStateBadge(a *app.App, t *ui.Theme) (string, ui.Color) {
	switch a.AI.State {
	case session.StateThinking:
		return i18n.T("ai.state.thinking"), t.Accent
	case session.StateAwaitingApproval:
		return i18n.T("ai.state.awaitingApproval"), t.Warning
	case session.StateRunning:
		return i18n.T("ai.state.running"), t.Success
	default:
		return "", t.TextMuted
	}
}

func aiMessageStream(c *ui.Context, a *app.App, t *ui.Theme) {
	// Follow the end unless the user scrolled away from it.
	if a.AI.Scroll.Y >= a.AI.Scroll.MaxY {
		a.AI.Scroll.Y = math.MaxFloat32
	}
	ui.Scroll(c).Grow(1).TrackScroll(&a.AI.Scroll).Children(func() {
		ui.Column(c).Padding(10, 14).Gap(8).Children(func() {
			if !a.AI.Configured {
				aiUnconfigured(c, a, t)
				aiErrorText(c, a, t)
				return
			}
			empty := len(a.AI.Messages) == 0 && a.AI.Streaming == "" &&
				a.AI.PendingCommand == nil && a.AI.PendingPlan == nil &&
				a.AI.RunningCommand == ""
			if empty {
				aiEmptyState(c, a, t)
				aiErrorText(c, a, t)
				return
			}
			for i := range a.AI.Messages {
				aiMessage(c, a, i, t)
			}
			if a.AI.Streaming != "" {
				aiAssistantBubble(c, a.AI.Streaming+"▌", t)
			}
			if a.AI.PendingCommand != nil {
				commandCard(c, a, a.AI.PendingCommand, a.AI.CommandDraft,
					a.AI.CommandStatus, false, t)
			}
			if a.AI.PendingPlan != nil {
				planCard(c, a, a.AI.PendingPlan, a.AI.PlanStatus, false, t)
			}
			aiErrorText(c, a, t)
			if a.AI.State == session.StateThinking && a.AI.Streaming == "" {
				ui.Row(c).Gap(6).AlignItems(ui.Center).Children(func() {
					ui.Spinner(c)
					ui.Text(c, i18n.T("ai.state.thinking")).FontSize(fontSize(12)).TextColor(t.TextMuted)
				})
			}
			if a.AI.RunningCommand != "" {
				aiRunningBlock(c, a, t)
			}
		})
	})
}

func aiErrorText(c *ui.Context, a *app.App, t *ui.Theme) {
	if a.AI.Err == "" {
		return
	}
	color := t.Danger
	if a.AI.ErrCancelled {
		color = t.TextMuted
	}
	ui.Text(c, a.AI.Err).FontSize(fontSize(13)).TextColor(color).Wrap()
}

func aiUnconfigured(c *ui.Context, a *app.App, t *ui.Theme) {
	ui.Column(c).Padding(24).Gap(10).Center().Children(func() {
		ui.Text(c, i18n.T("ai.empty.unconfiguredTitle")).FontSize(fontSize(15)).Bold()
		ui.Text(c, i18n.T("ai.empty.unconfiguredHint")).
			FontSize(fontSize(13)).TextColor(t.TextMuted)
		ui.PrimaryButton(c, i18n.T("ai.empty.configure")).OnClick(func() { a.OpenSettings(app.SettingsModel) })
	})
}

func aiEmptyState(c *ui.Context, a *app.App, t *ui.Theme) {
	ui.Column(c).Padding(16).Gap(10).Children(func() {
		ui.Text(c, i18n.T("ai.empty.hint")).
			FontSize(fontSize(13)).TextColor(t.TextMuted)
		for _, s := range aiSuggestions(a.SelectedProtocol()) {
			suggestion := s
			ui.Button(c, suggestion).OnClick(func() { a.AI.Input = suggestion })
		}
	})
}

func aiSuggestions(protocol string) []string {
	switch protocol {
	case "winrm":
		return []string{i18n.T("ai.suggestion.systemInfo"), i18n.T("ai.suggestion.diskSpace"),
			i18n.T("ai.suggestion.systemEvents")}
	case "mysql", "postgres", "sqlite", "clickhouse", "sqlserver":
		return []string{i18n.T("ai.suggestion.dbTables"), i18n.T("ai.suggestion.dbSchema"),
			i18n.T("ai.suggestion.dbRowCounts")}
	case "redis":
		return []string{i18n.T("ai.suggestion.redisInfo"), i18n.T("ai.suggestion.redisKeys"),
			i18n.T("ai.suggestion.redisMemory")}
	case "elasticsearch":
		return []string{i18n.T("ai.suggestion.esHealth"), i18n.T("ai.suggestion.esIndices"),
			i18n.T("ai.suggestion.esSearch")}
	default:
		return []string{i18n.T("ai.suggestion.diskUsage"), i18n.T("ai.suggestion.load"),
			i18n.T("ai.suggestion.logs"), i18n.T("ai.suggestion.docker")}
	}
}

func aiRunningBlock(c *ui.Context, a *app.App, t *ui.Theme) {
	ui.Row(c).Gap(6).AlignItems(ui.Center).Children(func() {
		ui.Spinner(c)
		ui.Text(c, "$ "+a.AI.RunningCommand).Font("monospace").FontSize(fontSize(12)).
			TextColor(t.TextMuted).SingleLine()
		if !a.AI.RunStartedAt.IsZero() {
			elapsed := time.Since(a.AI.RunStartedAt).Seconds()
			ui.Text(c, fmt.Sprintf("%.0fs", elapsed)).FontSize(fontSize(12)).TextColor(t.TextMuted)
		}
	})
	if a.AI.RunOutput != "" {
		ui.Scroll(c).MaxHeight(180).Children(func() {
			ui.Column(c).Background(t.Surface).Border(1, t.Border).Radius(4).
				Padding(6, 8).Children(func() {
				ui.Text(c, a.AI.RunOutput).Font("monospace").FontSize(fontSize(12)).Wrap()
			})
		})
	}
}

func aiComposer(c *ui.Context, a *app.App, t *ui.Theme) {
	ui.Divider(c)
	ui.Row(c).Padding(8, 12).Gap(8).AlignItems(ui.Center).Children(func() {
		placeholder := i18n.T("ai.composer.placeholder")
		switch {
		case !a.AI.Configured:
			placeholder = i18n.T("ai.composer.unconfigured")
		case a.AI.State == session.StateAwaitingApproval:
			placeholder = i18n.T("ai.composer.awaitingApproval")
		case a.AI.Busy():
			placeholder = i18n.T("ai.composer.busy")
		}
		c.Key("ai-input")
		input := ui.TextInput(c, &a.AI.Input).Grow(1).
			Placeholder(placeholder)
		if a.AI.InputDisabled() {
			input = input.Disabled(true)
		}
		input.OnSubmit(a.SendAIMessage)

		send := ui.PrimaryButton(c, i18n.T("ai.action.send"))
		if a.AI.InputDisabled() || strings.TrimSpace(a.AI.Input) == "" {
			send = send.Disabled(true)
		}
		send.OnClick(a.SendAIMessage)
	})
}

func aiHistoryPanel(c *ui.Context, a *app.App, t *ui.Theme) {
	ui.Column(c).Width(340).MaxHeight(420).Padding(10).Gap(6).Children(func() {
		ui.Row(c).AlignItems(ui.Center).Children(func() {
			ui.Text(c, i18n.T("ai.history.title")).FontSize(fontSize(13)).Bold()
			ui.Spacer(c)
			ui.Button(c, i18n.T("ai.action.newConversation")).OnClick(a.NewAIConversation)
		})
		if len(a.AI.Conversations) == 0 {
			ui.Text(c, i18n.T("ai.history.empty")).FontSize(fontSize(13)).TextColor(t.TextMuted)
			return
		}
		ui.Scroll(c).MaxHeight(340).Children(func() {
			ui.Column(c).Gap(4).Children(func() {
				for _, conv := range a.AI.Conversations {
					aiHistoryRow(c, a, conv.ID, conv.Title, t)
				}
			})
		})
	})
}

func aiHistoryRow(c *ui.Context, a *app.App, id, title string, t *ui.Theme) {
	if a.AI.RenameID == id {
		ui.Row(c).Gap(4).AlignItems(ui.Center).Children(func() {
			c.Key("ai-rename-" + id)
			ui.TextInput(c, &a.AI.RenameTitle).Grow(1)
			ui.Button(c, i18n.T("ai.action.save")).OnClick(func() {
				a.RenameAIConversation(id, a.AI.RenameTitle)
			})
			ui.Button(c, i18n.T("ai.action.cancel")).OnClick(func() {
				a.AI.RenameID = ""
				a.AI.RenameTitle = ""
			})
		})
		return
	}
	ui.Row(c).Gap(4).AlignItems(ui.Center).Children(func() {
		label := title
		if id == a.AI.SessionID {
			label = "● " + label
		}
		ui.Button(c, label).Grow(1).OnClick(func() { a.SwitchAIConversation(id) })
		ui.Button(c, i18n.T("ai.action.rename")).OnClick(func() {
			a.AI.RenameID = id
			a.AI.RenameTitle = title
			a.AI.DeleteConfirm = ""
		})
		del := ui.Button(c, i18n.T("ai.action.delete"))
		if a.AI.DeleteConfirm == id {
			del = ui.Button(c, i18n.T("ai.action.confirmDelete"))
		}
		del.OnClick(func() {
			if a.AI.DeleteConfirm != id {
				a.AI.DeleteConfirm = id
				return
			}
			a.AI.DeleteConfirm = ""
			a.DeleteAIConversation(id)
		})
	})
}

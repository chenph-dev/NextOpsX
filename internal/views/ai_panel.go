package views

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/egoist/mygo/ui"

	"nextopsx/internal/app"
	"nextopsx/internal/einoagent/session"
)

// AIPanel is the native AI chat: header, message stream, composer.
func AIPanel(c *ui.Context, a *app.App) {
	t := c.Theme()
	ui.Column(c).Fill().Children(func() {
		aiHeader(c, a, t)
		if a.AI.ConfigOpen {
			aiConfigForm(c, a, t)
		}
		if a.AI.Loading {
			ui.Row(c).Padding(24).Center().Children(func() {
				ui.Spinner(c)
				ui.Text(c, "正在载入会话…").TextColor(t.TextMuted)
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
			ui.Text(c, label).FontSize(11).TextColor(color)
		}
		if a.AI.Configured {
			if a.AI.Model != "" {
				ui.Text(c, a.AI.Model).FontSize(11).TextColor(t.TextMuted)
			}
		} else {
			ui.Text(c, "未配置 AI").FontSize(11).TextColor(t.Warning)
		}
		ui.Spacer(c)

		if a.AI.Busy() {
			ui.Button(c, "取消").OnClick(a.CancelAIRun)
		}
		ui.Button(c, "新对话").OnClick(a.NewAIConversation)

		history := ui.Button(c, "历史")
		history = history.OnClick(func() { a.AI.HistoryOpen = !a.AI.HistoryOpen })
		ui.Popover(c, history, &a.AI.HistoryOpen, func() {
			aiHistoryPanel(c, a, t)
		})

		clear := ui.Button(c, "清空")
		if a.AI.Busy() || a.AI.State == session.StateAwaitingApproval || len(a.AI.Messages) == 0 {
			clear = clear.Disabled(true)
		}
		clear.OnClick(a.ClearAIConversation)

		ui.Button(c, "设置").OnClick(func() { a.AI.ConfigOpen = !a.AI.ConfigOpen })
	})
	ui.Divider(c)
}

func aiStateBadge(a *app.App, t *ui.Theme) (string, ui.Color) {
	switch a.AI.State {
	case session.StateThinking:
		return "思考中…", t.Accent
	case session.StateAwaitingApproval:
		return "等待审批", t.Warning
	case session.StateRunning:
		return "执行中…", t.Success
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
					ui.Text(c, "思考中…").FontSize(11).TextColor(t.TextMuted)
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
	ui.Text(c, a.AI.Err).FontSize(12).TextColor(color).Wrap()
}

func aiUnconfigured(c *ui.Context, a *app.App, t *ui.Theme) {
	ui.Column(c).Padding(24).Gap(10).Center().Children(func() {
		ui.Text(c, "尚未配置 AI 后端").FontSize(15).Bold()
		ui.Text(c, "填写协议、模型名称和 API Key 后即可开始对话。").
			FontSize(12).TextColor(t.TextMuted)
		ui.PrimaryButton(c, "配置 AI").OnClick(func() { a.AI.ConfigOpen = true })
	})
}

func aiEmptyState(c *ui.Context, a *app.App, t *ui.Theme) {
	ui.Column(c).Padding(16).Gap(10).Children(func() {
		ui.Text(c, "向 AI 描述要执行的运维任务，命令会先交由你审批。").
			FontSize(12).TextColor(t.TextMuted)
		for _, s := range aiSuggestions(a.SelectedProtocol()) {
			suggestion := s
			ui.Button(c, suggestion).OnClick(func() { a.AI.Input = suggestion })
		}
	})
}

func aiSuggestions(protocol string) []string {
	switch protocol {
	case "winrm":
		return []string{"查看系统信息", "检查磁盘空间", "查看最近的系统事件"}
	case "mysql", "postgres", "sqlite", "clickhouse", "sqlserver":
		return []string{"列出当前数据库的表", "查看表结构", "统计各表行数"}
	case "redis":
		return []string{"查看 Redis INFO", "列出部分 key", "查看内存使用"}
	case "elasticsearch":
		return []string{"查看集群健康状态", "列出索引", "搜索最近文档"}
	default:
		return []string{"查看磁盘空间", "查看系统负载", "查看最近的系统日志", "检查 Docker 容器"}
	}
}

func aiRunningBlock(c *ui.Context, a *app.App, t *ui.Theme) {
	ui.Row(c).Gap(6).AlignItems(ui.Center).Children(func() {
		ui.Spinner(c)
		ui.Text(c, "$ "+a.AI.RunningCommand).Font("monospace").FontSize(11).
			TextColor(t.TextMuted).SingleLine()
		if !a.AI.RunStartedAt.IsZero() {
			elapsed := time.Since(a.AI.RunStartedAt).Seconds()
			ui.Text(c, fmt.Sprintf("%.0fs", elapsed)).FontSize(11).TextColor(t.TextMuted)
		}
	})
	if a.AI.RunOutput != "" {
		ui.Scroll(c).MaxHeight(180).Children(func() {
			ui.Column(c).Background(t.Surface).Border(1, t.Border).Radius(4).
				Padding(6, 8).Children(func() {
				ui.Text(c, a.AI.RunOutput).Font("monospace").FontSize(11).Wrap()
			})
		})
	}
}

func aiComposer(c *ui.Context, a *app.App, t *ui.Theme) {
	ui.Divider(c)
	ui.Row(c).Padding(8, 12).Gap(8).AlignItems(ui.Center).Children(func() {
		placeholder := "输入消息，Enter 发送；/clear 清空，/new 新对话"
		switch {
		case !a.AI.Configured:
			placeholder = "请先配置 AI 后端"
		case a.AI.State == session.StateAwaitingApproval:
			placeholder = "请先处理上方的审批卡"
		case a.AI.Busy():
			placeholder = "AI 正在处理，可点击「取消」中止"
		}
		c.Key("ai-input")
		input := ui.TextInput(c, &a.AI.Input).Grow(1).
			Placeholder(placeholder)
		if a.AI.InputDisabled() {
			input = input.Disabled(true)
		}
		input.OnSubmit(a.SendAIMessage)

		send := ui.PrimaryButton(c, "发送")
		if a.AI.InputDisabled() || strings.TrimSpace(a.AI.Input) == "" {
			send = send.Disabled(true)
		}
		send.OnClick(a.SendAIMessage)
	})
}

func aiConfigForm(c *ui.Context, a *app.App, t *ui.Theme) {
	ui.Column(c).Padding(10, 14).Gap(8).Background(t.Surface).Children(func() {
		ui.Row(c).AlignItems(ui.Center).Children(func() {
			ui.Text(c, "AI 配置").FontSize(13).Bold()
			ui.Spacer(c)
			ui.Button(c, "收起").OnClick(func() { a.AI.ConfigOpen = false })
		})
		ui.Row(c).Gap(8).AlignItems(ui.Center).Children(func() {
			ui.Text(c, "协议").Width(72).FontSize(12).TextColor(t.TextMuted)
			c.Key("ai-cfg-provider")
			ui.Select(c, &a.AI.Provider, []string{"openai", "claude"})
			ui.Text(c, "模型").Width(36).FontSize(12).TextColor(t.TextMuted)
			c.Key("ai-cfg-model")
			ui.TextInput(c, &a.AI.Model).Grow(1).
				Placeholder("如 gpt-4o / claude-sonnet-4-5")
		})
		ui.Row(c).Gap(8).AlignItems(ui.Center).Children(func() {
			ui.Text(c, "Base URL").Width(72).FontSize(12).TextColor(t.TextMuted)
			c.Key("ai-cfg-baseurl")
			ui.TextInput(c, &a.AI.BaseURL).Grow(1).
				Placeholder("可留空使用默认地址")
		})
		ui.Row(c).Gap(8).AlignItems(ui.Center).Children(func() {
			ui.Text(c, "API Key").Width(72).FontSize(12).TextColor(t.TextMuted)
			c.Key("ai-cfg-key")
			ui.TextInput(c, &a.AI.APIKey).Grow(1).Password()
		})
		ui.Text(c, "openai 协议兼容 OpenAI / DeepSeek / Ollama 等；claude 协议用于 Anthropic 官方 API。").
			FontSize(11).TextColor(t.TextMuted).Wrap()
		if a.AI.Err != "" {
			ui.Text(c, a.AI.Err).FontSize(12).TextColor(t.Danger).Wrap()
		}
		ui.Row(c).Justify(ui.End).Gap(6).Children(func() {
			ui.Button(c, "取消").OnClick(func() { a.AI.ConfigOpen = false })
			ui.PrimaryButton(c, "保存").OnClick(a.SaveAIConfig)
		})
	})
	ui.Divider(c)
}

func aiHistoryPanel(c *ui.Context, a *app.App, t *ui.Theme) {
	ui.Column(c).Width(340).MaxHeight(420).Padding(10).Gap(6).Children(func() {
		ui.Row(c).AlignItems(ui.Center).Children(func() {
			ui.Text(c, "历史对话").FontSize(13).Bold()
			ui.Spacer(c)
			ui.Button(c, "新对话").OnClick(a.NewAIConversation)
		})
		if len(a.AI.Conversations) == 0 {
			ui.Text(c, "暂无历史对话").FontSize(12).TextColor(t.TextMuted)
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
			ui.Button(c, "保存").OnClick(func() {
				a.RenameAIConversation(id, a.AI.RenameTitle)
			})
			ui.Button(c, "取消").OnClick(func() {
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
		ui.Button(c, "改名").OnClick(func() {
			a.AI.RenameID = id
			a.AI.RenameTitle = title
			a.AI.DeleteConfirm = ""
		})
		del := ui.Button(c, "删除")
		if a.AI.DeleteConfirm == id {
			del = ui.Button(c, "确认删除")
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

package views

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/egoist/mygo/ui"

	"nextopsx/internal/app"
	convstore "nextopsx/internal/store/conversations"
)

// aiToolCall is the JSON shape persisted in Message.ToolCalls.
type aiToolCall struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

func aiMessage(c *ui.Context, a *app.App, i int, t *ui.Theme) {
	msg := a.AI.Messages[i]
	switch msg.Role {
	case "user":
		aiUserBubble(c, msg.Content, t)
	case "assistant":
		if strings.TrimSpace(msg.Content) != "" {
			aiAssistantBubble(c, msg.Content, t)
		}
		calls := parseAIToolCalls(msg.ToolCalls)
		if len(calls) == 0 {
			return
		}
		call := calls[0]
		status := aiToolStatus(a.AI.Messages, i, call.ID)
		switch call.Name {
		case "create_plan":
			if plan := parseAIPlan(call.Arguments); plan != nil {
				planCard(c, a, plan, status, true, t)
			}
		default:
			if cmd := parseAICommand(call.Arguments); cmd != nil {
				commandCard(c, a, cmd, cmd.Command, status, true, t)
			}
		}
	case "tool":
		toolOutputBlock(c, a, msg, t)
	}
}

func aiUserBubble(c *ui.Context, text string, t *ui.Theme) {
	svc := c.Services()
	ui.Row(c).Justify(ui.End).Gap(6).AlignItems(ui.Center).Children(func() {
		ui.Button(c, "复制").FontSize(11).OnClick(func() { svc.WriteClipboard(text) })
		ui.Column(c).MaxWidth(560).Padding(8, 10).Radius(8).
			Background(t.Accent).Children(func() {
			ui.Text(c, text).TextColor(t.AccentText).FontSize(12).Wrap()
		})
	})
}

func aiAssistantBubble(c *ui.Context, text string, t *ui.Theme) {
	svc := c.Services()
	ui.Row(c).Gap(6).AlignItems(ui.Start).Children(func() {
		ui.Column(c).MaxWidth(720).Padding(8, 10).Radius(8).
			Background(t.Surface).Children(func() {
			MarkdownView(c, text)
		})
		ui.Button(c, "复制").FontSize(11).OnClick(func() { svc.WriteClipboard(text) })
	})
}

func parseAIToolCalls(raw string) []aiToolCall {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	var calls []aiToolCall
	if json.Unmarshal([]byte(raw), &calls) != nil {
		return nil
	}
	return calls
}

func parseAICommand(args string) *app.CommandSuggestion {
	var v struct {
		Command string `json:"command"`
		SQL     string `json:"sql"`
		Why     string `json:"why"`
	}
	if json.Unmarshal([]byte(args), &v) != nil {
		return nil
	}
	command := v.Command
	if command == "" {
		command = v.SQL
	}
	if command == "" {
		return nil
	}
	return &app.CommandSuggestion{Command: command, Why: v.Why}
}

func parseAIPlan(args string) *app.PlanSuggestion {
	var v app.PlanSuggestion
	if json.Unmarshal([]byte(args), &v) != nil || (v.Goal == "" && len(v.Steps) == 0) {
		return nil
	}
	return &v
}

func aiToolStatus(msgs []convstore.Message, i int, callID string) string {
	if i+1 < len(msgs) && msgs[i+1].Role == "tool" && msgs[i+1].ToolCallID == callID {
		switch msgs[i+1].ApprovalStatus {
		case app.AIStatusRejected:
			return app.AIStatusRejected
		case app.AIStatusAuto:
			return app.AIStatusAuto
		default:
			return app.AIStatusApproved
		}
	}
	return app.AIStatusPending
}

// aiToolMeta is the structured metadata persisted in Message.ToolResult.
type aiToolMeta struct {
	Command    string `json:"command"`
	ExitCode   int    `json:"exitCode"`
	DurationMS int64  `json:"durationMs"`
	Status     string `json:"status"`
	Cancelled  bool   `json:"cancelled"`
}

func toolOutputBlock(c *ui.Context, a *app.App, msg convstore.Message, t *ui.Theme) {
	var meta aiToolMeta
	_ = json.Unmarshal([]byte(msg.ToolResult), &meta)
	open := a.AI.ToolOpen[msg.ID]

	head := "▸ "
	if open {
		head = "▾ "
	}
	if meta.Command != "" {
		head += "$ " + meta.Command
	} else if msg.ToolName != "" {
		head += msg.ToolName
	} else {
		head += "工具输出"
	}
	switch {
	case meta.Status == app.AIStatusRejected:
		head += " · 已拒绝"
	case meta.Cancelled:
		head += " · 已取消"
	case msg.ToolResult != "":
		head += fmt.Sprintf(" · exit %d", meta.ExitCode)
	}
	if lines := countOutputLines(msg.Content); lines > 0 {
		head += fmt.Sprintf(" · %d 行", lines)
	}
	ui.Button(c, head).FontSize(11).SingleLine().OnClick(func() {
		a.ToggleToolOutput(msg.ID)
	})
	if !open {
		return
	}
	content := msg.Content
	if strings.TrimSpace(content) == "" {
		content = "(无输出)"
	}
	ui.Scroll(c).MaxHeight(260).Children(func() {
		ui.Column(c).Background(t.Background).Border(1, t.Border).Radius(4).
			Padding(6, 8).Children(func() {
			ui.Text(c, content).Font("monospace").FontSize(11).Wrap()
		})
	})
}

func countOutputLines(s string) int {
	n := 0
	for _, line := range strings.Split(s, "\n") {
		if strings.TrimSpace(line) != "" {
			n++
		}
	}
	return n
}

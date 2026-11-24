package ai

import (
	"encoding/json"
	"strings"

	"github.com/egoist/mygo/ui"

	convstore "nextopsx/store/conversations"
	"nextopsx/ui/markdown"
	"nextopsx/ui/style"
	"nextopsx/ui/workspace"
)

// aiToolCall is the JSON shape persisted in Message.ToolCalls.
type aiToolCall struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

func aiMessage(c *ui.Context, a *workspace.State, i int, t *ui.Theme) {
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
			cmd := parseAICommand(call.Arguments)
			if cmd == nil {
				return
			}
			// A pending command still needs the approval card. Once approved,
			// auto-approved or rejected, the command and its output become one
			// terminal bubble instead.
			if status == workspace.AIStatusPending {
				commandCard(c, a, cmd, cmd.Command, status, true, t)
				return
			}
			tool, ok := aiToolMessageForCall(a.AI.Messages, i, call.ID)
			if !ok {
				// A non-pending status comes from a persisted tool result.
				return
			}
			var meta aiToolMeta
			_ = json.Unmarshal([]byte(tool.ToolResult), &meta)
			if meta.Command == "" {
				meta.Command = cmd.Command
			}
			aiTerminalBubble(c, a, tool, meta, status)
		}
	case "tool":
		// The assistant message that proposed this call already rendered the
		// result as a terminal bubble; do not draw it twice.
		if aiToolHandledByCall(a.AI.Messages, i) {
			return
		}
		var meta aiToolMeta
		_ = json.Unmarshal([]byte(msg.ToolResult), &meta)
		if meta.Command != "" {
			status := meta.Status
			if status == "" {
				status = msg.ApprovalStatus
			}
			aiTerminalBubble(c, a, msg, meta, status)
			return
		}
		aiToolTextBubble(c, msg, t)
	}
}

func aiUserBubble(c *ui.Context, text string, t *ui.Theme) {
	ui.Column(c).AlignItems(ui.End).Gap(2).Children(func() {
		ui.Row(c).Justify(ui.End).AlignItems(ui.Center).Children(func() {
			ui.Column(c).MaxWidth(560).AlignItems(ui.Stretch).Padding(8, 10).Radius(8).
				Background(t.Accent).Children(func() {
				ui.Text(c, text).TextColor(t.AccentText).FontSize(style.FontSize(13)).Wrap()
			})
		})
		ui.Row(c).Justify(ui.End).Children(func() {
			aiBubbleToolbar(c, t.TextMuted, aiCopyAction(text))
		})
	})
}

func aiAssistantBubble(c *ui.Context, text string, t *ui.Theme) {
	ui.Column(c).AlignItems(ui.Start).Gap(2).Children(func() {
		ui.Row(c).AlignItems(ui.Start).Children(func() {
			ui.Column(c).MaxWidth(720).AlignItems(ui.Stretch).Padding(8, 10).Radius(8).
				Background(t.Surface).Children(func() {
				markdown.View(c, text)
			})
		})
		ui.Row(c).Justify(ui.Start).Children(func() {
			aiBubbleToolbar(c, t.TextMuted, aiCopyAction(text))
		})
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

func parseAICommand(args string) *workspace.CommandSuggestion {
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
	return &workspace.CommandSuggestion{Command: command, Why: v.Why}
}

func parseAIPlan(args string) *workspace.PlanSuggestion {
	var v workspace.PlanSuggestion
	if json.Unmarshal([]byte(args), &v) != nil || (v.Goal == "" && len(v.Steps) == 0) {
		return nil
	}
	return &v
}

func aiToolStatus(msgs []convstore.Message, i int, callID string) string {
	tool, ok := aiToolMessageForCall(msgs, i, callID)
	if !ok {
		return workspace.AIStatusPending
	}
	switch tool.ApprovalStatus {
	case workspace.AIStatusRejected:
		return workspace.AIStatusRejected
	case workspace.AIStatusAuto:
		return workspace.AIStatusAuto
	case workspace.AIStatusPlan:
		return workspace.AIStatusPlan
	default:
		return workspace.AIStatusApproved
	}
}

// aiToolMeta is the structured metadata persisted in Message.ToolResult.
type aiToolMeta struct {
	Command    string `json:"command"`
	ExitCode   int    `json:"exitCode"`
	DurationMS int64  `json:"durationMs"`
	Status     string `json:"status"`
	Cancelled  bool   `json:"cancelled"`
}

// aiToolMessageForCall returns the tool result that follows an assistant
// tool call, if it is already persisted.
func aiToolMessageForCall(msgs []convstore.Message, i int, callID string) (convstore.Message, bool) {
	if i+1 < len(msgs) && msgs[i+1].Role == "tool" && msgs[i+1].ToolCallID == callID {
		return msgs[i+1], true
	}
	return convstore.Message{}, false
}

// aiToolHandledByCall reports whether a tool message was already rendered
// together with its assistant call, so the message stream can skip it.
func aiToolHandledByCall(msgs []convstore.Message, i int) bool {
	if i <= 0 || msgs[i-1].Role != "assistant" {
		return false
	}
	calls := parseAIToolCalls(msgs[i-1].ToolCalls)
	if len(calls) == 0 {
		return false
	}
	// aiMessage renders only the first call of an assistant message, and only
	// command calls are drawn together with their tool result. Other tool
	// results (load_skill, run_skill_script, …) are rendered on their own.
	call := calls[0]
	if call.ID != msgs[i].ToolCallID || parseAICommand(call.Arguments) == nil {
		return false
	}
	return aiToolStatus(msgs, i-1, call.ID) != workspace.AIStatusPending
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

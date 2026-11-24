package ai

import (
	"fmt"
	"strings"

	"github.com/egoist/mygo/ui"

	"nextopsx/i18n"
	convstore "nextopsx/store/conversations"
	"nextopsx/ui/ansi"
	"nextopsx/ui/markdown"
	"nextopsx/ui/style"
	"nextopsx/ui/widgets"
	"nextopsx/ui/workspace"
)

// aiTerminalBubble renders one executed command and its output as a terminal.
// The header carries the command, exit status and line count; the body shows
// the output with ANSI SGR colors parsed into spans.
func aiTerminalBubble(c *ui.Context, a *workspace.State, msg convstore.Message, meta aiToolMeta, status string) {
	command := meta.Command
	if command == "" {
		command = msg.ToolName
	}
	if command == "" {
		command = i18n.T("ai.tool.output")
	}

	open := a.AI.ToolOpen[msg.ID]
	head := "▸ "
	if open {
		head = "▾ "
	}
	head += "$ " + ansi.Truncate(command, 72)
	switch {
	case status == workspace.AIStatusRejected:
		head += i18n.T("ai.tool.rejected")
	case meta.Cancelled:
		head += i18n.T("ai.tool.cancelled")
	case meta.Command != "" && msg.ToolResult != "":
		head += fmt.Sprintf(" · exit %d", meta.ExitCode)
	}
	if status == workspace.AIStatusPlan {
		head += i18n.T("ai.tool.plan")
	}
	if lines := countOutputLines(msg.Content); lines > 0 {
		head += i18n.Tf("ai.tool.lines", lines)
	}

	copyText := "$ " + command
	if strings.TrimSpace(msg.Content) != "" {
		copyText += "\n" + msg.Content
	}

	ui.Column(c).AlignItems(ui.Start).Gap(2).Children(func() {
		ui.Column(c).MaxWidthPercent(100).AlignItems(ui.Stretch).
			Background(ansi.Bg).Border(1, ansi.Border).Radius(8).
			Padding(8, 10).Gap(6).Children(func() {
			ui.Button(c, head).Font("monospace").FontSize(style.FontSize(12)).
				TextColor(ansi.Fg).Background(ansi.Bg).SingleLine().MinWidth(0).
				Justify(ui.Start).Tooltip(command).
				OnClick(func() { a.ToggleToolOutput(msg.ID) })
			if open {
				widgets.TerminalOutput(c, msg.Content, 280)
			}
		})
		ui.Row(c).Justify(ui.Start).Children(func() {
			aiBubbleToolbar(c, ansi.Muted, aiCopyAction(copyText))
		})
	})
}

// aiLiveTerminalOutput renders a bounded tail of streaming output. Reparsing
// the whole accumulated buffer on every frame would make a chatty command
// O(n²); the finalized tool message still renders the complete output.
func aiLiveTerminalOutput(c *ui.Context, output string, maxHeight float32) {
	widgets.TerminalOutput(c, ansi.LiveTail(output, 8<<10), maxHeight)
}

// aiToolTextBubble renders a non-command tool result as Markdown, so tool
// output that is text (file contents, query results, notes) keeps headings,
// lists and code formatting instead of looking like a raw dump.
func aiToolTextBubble(c *ui.Context, msg convstore.Message, t *ui.Theme) {
	name := msg.ToolName
	if name == "" {
		name = i18n.T("ai.tool.output")
	}
	empty := strings.TrimSpace(msg.Content) == ""
	ui.Column(c).AlignItems(ui.Start).Gap(2).Children(func() {
		ui.Column(c).MaxWidth(720).AlignItems(ui.Stretch).Padding(8, 10).Radius(8).
			Background(t.Surface).Border(1, t.Border).Children(func() {
			ui.Text(c, name).FontSize(style.FontSize(11)).TextColor(t.TextMuted)
			if empty {
				ui.Text(c, i18n.T("ai.tool.noOutput")).FontSize(style.FontSize(12)).
					TextColor(t.TextMuted)
				return
			}
			markdown.View(c, msg.Content)
		})
		if !empty {
			ui.Row(c).Justify(ui.Start).Children(func() {
				aiBubbleToolbar(c, t.TextMuted, aiCopyAction(msg.Content))
			})
		}
	})
}

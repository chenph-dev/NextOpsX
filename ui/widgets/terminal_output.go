package widgets

import (
	"strings"

	"github.com/egoist/mygo/ui"

	"nextopsx/i18n"
	"nextopsx/ui/ansi"
	"nextopsx/ui/style"
)

// TerminalOutput renders command output in a scrollable monospace area with
// ANSI colors preserved.
func TerminalOutput(c *ui.Context, output string, maxHeight float32) {
	if strings.TrimSpace(output) == "" {
		ui.Text(c, i18n.T("ai.tool.noOutput")).Font("monospace").
			FontSize(style.FontSize(12)).TextColor(ansi.Muted)
		return
	}
	ui.Scroll(c).MaxHeight(maxHeight).Children(func() {
		ui.RichText(c, ansi.Spans(output)...).Font("monospace").
			FontSize(style.FontSize(12)).TextColor(ansi.Fg).Wrap()
	})
}

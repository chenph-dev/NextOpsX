// Package widgets holds small view pieces shared by feature panels.
package widgets

import (
	"github.com/egoist/mygo/ui"

	"nextopsx/ui/style"
)

// ErrorText renders a wrapped error line.
func ErrorText(c *ui.Context, msg string, t *ui.Theme) {
	ui.Text(c, msg).FontSize(style.FontSize(13)).TextColor(t.Danger).Wrap()
}

// Hint renders a muted one-line hint.
func Hint(c *ui.Context, text string, t *ui.Theme) {
	ui.Text(c, text).FontSize(style.FontSize(13)).TextColor(t.TextMuted)
}

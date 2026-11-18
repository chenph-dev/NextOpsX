package result

import (
	"unicode/utf8"

	"github.com/egoist/mygo/ui"

	"nextopsx/domain/dbexec"
	"nextopsx/i18n"
	"nextopsx/ui/style"
	"nextopsx/ui/workspace"
)

// GridView renders a structured query result in a virtualized table.
// A zero-row result gets one muted placeholder row so wide headers still
// have a horizontal scroll extent (MyGo's Table needs a row for that).
func GridView(c *ui.Context, grid *ui.ListState, res *dbexec.Result, t *ui.Theme) {
	if res == nil || len(res.Columns) == 0 {
		return
	}
	cols := make([]ui.TableColumn, len(res.Columns))
	for i, name := range res.Columns {
		if name == "" {
			name = i18n.Tf("db.result.column", i+1)
		}
		cols[i] = ui.TableColumn{Title: name, MinWidth: resultColumnMinWidth(name)}
	}
	rows := len(res.Rows)
	placeholder := rows == 0
	if placeholder {
		rows = 1
	}
	ui.Table(c, grid, cols, rows, func(row, col int) {
		if placeholder {
			if col == 0 {
				ui.Text(c, i18n.T("db.result.empty")).FontSize(style.FontSize(12)).TextColor(t.TextMuted).SingleLine()
			} else {
				ui.Text(c, "").SingleLine()
			}
			return
		}
		if row < 0 || row >= len(res.Rows) || col < 0 || col >= len(res.Columns) {
			return
		}
		text := "NULL"
		if col < len(res.Rows[row]) && res.Rows[row][col] != nil {
			text = workspace.DBCellText(res.Rows[row][col])
		}
		ui.Text(c, text).FontSize(style.FontSize(12)).SingleLine()
	}).Grow(1).MinWidth(0)
}

func resultColumnMinWidth(name string) float32 {
	w := float32(utf8.RuneCountInString(name))*8 + 28
	if w < 90 {
		return 90
	}
	if w > 240 {
		return 240
	}
	return w
}

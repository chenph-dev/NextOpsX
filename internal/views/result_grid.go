package views

import (
	"fmt"
	"unicode/utf8"

	"github.com/egoist/mygo/ui"

	"nextopsx/internal/app"
	"nextopsx/internal/dbexec"
)

// ResultGridView renders a structured query result in a virtualized table.
// A zero-row result gets one muted placeholder row so wide headers still
// have a horizontal scroll extent (MyGo's Table needs a row for that).
func ResultGridView(c *ui.Context, grid *ui.ListState, res *dbexec.Result, t *ui.Theme) {
	if res == nil || len(res.Columns) == 0 {
		return
	}
	cols := make([]ui.TableColumn, len(res.Columns))
	for i, name := range res.Columns {
		if name == "" {
			name = fmt.Sprintf("列 %d", i+1)
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
				ui.Text(c, "(无数据)").FontSize(11).TextColor(t.TextMuted).SingleLine()
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
			text = app.DBCellText(res.Rows[row][col])
		}
		ui.Text(c, text).FontSize(11).SingleLine()
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

package views

import (
	"fmt"
	"unicode/utf8"

	"github.com/egoist/mygo/ui"
	"github.com/egoist/mygo/ui/editor"

	"nextopsx/internal/app"
)

// DBQueryTab is the SQL editor above its result grid.
func DBQueryTab(c *ui.Context, a *app.App, tab *app.DBTab, t *ui.Theme) {
	if a.DB.Split <= 0 {
		a.DB.Split = 190
	}
	ui.SplitVertical(c, &a.DB.Split, func() {
		ui.Column(c).Fill().Background(t.Background).Children(func() {
			editor.View(c, tab.Editor).Fill()
		})
	}, func() {
		ui.Column(c).Fill().MinWidth(0).Padding(6, 8).Gap(6).Children(func() {
			if tab.Running {
				ui.Row(c).Padding(20).Center().Children(func() {
					ui.Spinner(c)
					ui.Text(c, "执行中…").FontSize(11).TextColor(t.TextMuted)
				})
				return
			}
			if tab.Err != "" {
				dbErrorText(c, tab.Err, t)
				return
			}
			if tab.Result == nil {
				dbHint(c, "执行查询后在此显示结果。", t)
				return
			}
			dbResultMeta(c, tab, t)
			dbResultGrid(c, tab, t)
		})
	}).Fill()
}

// DBTableTab browses one table or view, 100 rows per page.
func DBTableTab(c *ui.Context, a *app.App, tab *app.DBTab, t *ui.Theme) {
	ui.Column(c).Fill().MinWidth(0).Padding(8).Gap(6).Children(func() {
		ui.Row(c).MinWidth(0).Gap(8).AlignItems(ui.Center).Children(func() {
			ui.Text(c, tab.TableName).FontSize(13).Bold().SingleLine()
			if tab.HasTotal {
				ui.Text(c, fmt.Sprintf("共 %d 行", tab.Total)).FontSize(11).TextColor(t.TextMuted)
			}
			if tab.DurationMS > 0 {
				ui.Text(c, fmt.Sprintf("%dms", tab.DurationMS)).FontSize(11).TextColor(t.TextMuted)
			}
			ui.Spacer(c)

			ui.Row(c).Gap(4).AlignItems(ui.Center).Shrink(0).Children(func() {
				prev := ui.Button(c, "上一页")
				if tab.Running || tab.Page <= 0 {
					prev = prev.Disabled(true)
				}
				prev.OnClick(func() { a.DBTablePage(-1) })
				ui.Text(c, dbPageLabel(tab.Page, tab.PageSize, tab.HasTotal, tab.Total)).FontSize(11)
				next := ui.Button(c, "下一页")
				if !dbTableCanNext(tab) {
					next = next.Disabled(true)
				}
				next.OnClick(func() { a.DBTablePage(1) })
				ui.Button(c, "刷新").OnClick(a.RefreshActiveDBTable)
			})
		})

		switch {
		case tab.Running:
			ui.Row(c).Padding(20).Center().Children(func() {
				ui.Spinner(c)
				ui.Text(c, "加载中…").FontSize(11).TextColor(t.TextMuted)
			})
		case tab.Err != "":
			dbErrorText(c, tab.Err, t)
		case tab.Result != nil:
			dbResultGrid(c, tab, t)
		default:
			dbHint(c, "加载表数据…", t)
		}
	})
}

func dbColumnMinWidth(name string) float32 {
	w := float32(utf8.RuneCountInString(name))*8 + 28
	if w < 90 {
		return 90
	}
	if w > 240 {
		return 240
	}
	return w
}

func dbResultGrid(c *ui.Context, tab *app.DBTab, t *ui.Theme) {
	res := tab.Result
	if res == nil || len(res.Columns) == 0 {
		return
	}
	cols := make([]ui.TableColumn, len(res.Columns))
	for i, name := range res.Columns {
		if name == "" {
			name = fmt.Sprintf("列 %d", i+1)
		}
		cols[i] = ui.TableColumn{Title: name, MinWidth: dbColumnMinWidth(name)}
	}
	rows := len(res.Rows)
	placeholder := rows == 0
	if placeholder {
		// A Table only gains a horizontal scroll extent once it has a
		// row; one muted placeholder keeps wide headers scrollable.
		rows = 1
	}
	ui.Table(c, &tab.Grid, cols, rows, func(row, col int) {
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

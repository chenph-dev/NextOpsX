package views

import (
	"fmt"

	"github.com/egoist/mygo/ui"
	"github.com/egoist/mygo/ui/editor"

	"nextopsx/i18n"
	"nextopsx/ui/app"
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
					ui.Text(c, i18n.T("db.query.running")).FontSize(fontSize(12)).TextColor(t.TextMuted)
				})
				return
			}
			if tab.Err != "" {
				dbErrorText(c, tab.Err, t)
				return
			}
			if tab.Result == nil {
				dbHint(c, i18n.T("db.query.empty"), t)
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
			ui.Text(c, tab.TableName).FontSize(fontSize(13)).Bold().SingleLine()
			if tab.HasTotal {
				ui.Text(c, i18n.Tf("db.table.total", tab.Total)).FontSize(fontSize(12)).TextColor(t.TextMuted)
			}
			if tab.DurationMS > 0 {
				ui.Text(c, fmt.Sprintf("%dms", tab.DurationMS)).FontSize(fontSize(12)).TextColor(t.TextMuted)
			}
			ui.Spacer(c)

			ui.Row(c).Gap(4).AlignItems(ui.Center).Shrink(0).Children(func() {
				prev := ui.Button(c, i18n.T("db.table.prev"))
				if tab.Running || tab.Page <= 0 {
					prev = prev.Disabled(true)
				}
				prev.OnClick(func() { a.DBTablePage(-1) })
				ui.Text(c, dbPageLabel(tab.Page, tab.PageSize, tab.HasTotal, tab.Total)).FontSize(fontSize(12))
				next := ui.Button(c, i18n.T("db.table.next"))
				if !dbTableCanNext(tab) {
					next = next.Disabled(true)
				}
				next.OnClick(func() { a.DBTablePage(1) })
				ui.Button(c, i18n.T("db.table.refresh")).OnClick(a.RefreshActiveDBTable)
			})
		})

		switch {
		case tab.Running:
			ui.Row(c).Padding(20).Center().Children(func() {
				ui.Spinner(c)
				ui.Text(c, i18n.T("db.table.loading")).FontSize(fontSize(12)).TextColor(t.TextMuted)
			})
		case tab.Err != "":
			dbErrorText(c, tab.Err, t)
		case tab.Result != nil:
			dbResultGrid(c, tab, t)
		default:
			dbHint(c, i18n.T("db.table.loadingHint"), t)
		}
	})
}

func dbResultGrid(c *ui.Context, tab *app.DBTab, t *ui.Theme) {
	ResultGridView(c, &tab.Grid, tab.Result, t)
}

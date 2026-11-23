package db

import (
	"fmt"

	"github.com/egoist/mygo/ui"
	"github.com/egoist/mygo/ui/editor"

	"nextopsx/i18n"
	"nextopsx/ui/style"
	"nextopsx/ui/views/result"
	"nextopsx/ui/widgets"
	"nextopsx/ui/workspace"
)

// QueryTab is the SQL editor above its result grid.
func QueryTab(c *ui.Context, a *workspace.State, tab *workspace.DBTab, t *ui.Theme) {
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
					ui.Text(c, i18n.T("db.query.running")).FontSize(style.FontSize(12)).TextColor(t.TextMuted)
				})
				return
			}
			if tab.Err != "" {
				widgets.ErrorText(c, tab.Err, t)
				return
			}
			if tab.Result == nil {
				widgets.Hint(c, i18n.T("db.query.empty"), t)
				return
			}
			dbResultMeta(c, tab, t)
			dbResultGrid(c, a, tab, t)
		})
	}).Fill()
}

// TableTab browses one table or view, 100 rows per page.
func TableTab(c *ui.Context, a *workspace.State, tab *workspace.DBTab, t *ui.Theme) {
	ui.Column(c).Fill().MinWidth(0).Padding(8).Gap(6).Children(func() {
		ui.Row(c).MinWidth(0).Gap(8).AlignItems(ui.Center).Children(func() {
			ui.Text(c, tab.TableName).FontSize(style.FontSize(13)).Bold().SingleLine()
			if tab.HasTotal {
				ui.Text(c, i18n.Tf("db.table.total", tab.Total)).FontSize(style.FontSize(12)).TextColor(t.TextMuted)
			}
			if tab.DurationMS > 0 {
				ui.Text(c, fmt.Sprintf("%dms", tab.DurationMS)).FontSize(style.FontSize(12)).TextColor(t.TextMuted)
			}
			ui.Spacer(c)

			ui.Row(c).Gap(4).AlignItems(ui.Center).Shrink(0).Children(func() {
				prev := ui.Button(c, i18n.T("db.table.prev"))
				if tab.Running || tab.Page <= 0 {
					prev = prev.Disabled(true)
				}
				prev.OnClick(func() { a.DBTablePage(-1) })
				ui.Text(c, dbPageLabel(tab.Page, tab.PageSize, tab.HasTotal, tab.Total)).FontSize(style.FontSize(12))
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
				ui.Text(c, i18n.T("db.table.loading")).FontSize(style.FontSize(12)).TextColor(t.TextMuted)
			})
		case tab.Err != "":
			widgets.ErrorText(c, tab.Err, t)
		case tab.Result != nil:
			dbResultGrid(c, a, tab, t)
		default:
			widgets.Hint(c, i18n.T("db.table.loadingHint"), t)
		}
	})
}

func dbResultGrid(c *ui.Context, a *workspace.State, tab *workspace.DBTab, t *ui.Theme) {
	exportLabel := i18n.T("db.export.menu.csv")
	if tab.Kind == workspace.DBTabTable {
		exportLabel = i18n.T("db.export.menu.page")
	}
	result.GridViewMenu(c, &tab.Grid, tab.Result, t, func(m *ui.Menu) {
		item := m.Item(exportLabel)
		item.Disabled(dbExportDisabled(tab))
		if item.Chosen() {
			a.ExportDBResult(tab)
		}
	})
}

// dbExportDisabled reports whether tab's result cannot be exported yet.
func dbExportDisabled(tab *workspace.DBTab) bool {
	return tab == nil || tab.Result == nil || len(tab.Result.Columns) == 0 ||
		tab.Running ||
		(tab.Kind == workspace.DBTabTable && tab.ResultPage != tab.Page)
}

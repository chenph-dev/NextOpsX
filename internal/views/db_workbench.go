package views

import (
	"fmt"

	"github.com/egoist/mygo/ui"

	"nextopsx/internal/app"
	"nextopsx/resources"
)

// DBWorkbench is the native database workbench: schema tree, query tabs and
// a status bar.
func DBWorkbench(c *ui.Context, a *app.App) {
	t := c.Theme()
	// Ctrl/Cmd+Enter runs the active query when the editor does not consume
	// the key first; the toolbar button is always available.
	c.OnShortcut(ui.Ctrl, ui.KeyEnter, a.RunActiveDBTab)
	c.OnShortcut(ui.Super, ui.KeyEnter, a.RunActiveDBTab)

	ui.Column(c).Fill().Children(func() {
		dbToolbar(c, a, t)
		ui.Divider(c)
		ui.Row(c).Grow(1).FillHeight().MinWidth(0).Children(func() {
			DBObjectTree(c, a, t)
			ui.Column(c).Grow(1).FillHeight().MinWidth(0).Children(func() {
				dbTabStrip(c, a, t)
				ui.Divider(c)
				tab := a.ActiveDBTab()
				if tab == nil {
					ui.Column(c).Fill().Center().Gap(6).Children(func() {
						ui.Text(c, "没有打开的标签").FontSize(12).TextColor(t.TextMuted)
						ui.Text(c, "点击标签栏右侧的 + 新建查询。").
							FontSize(11).TextColor(t.TextMuted)
					})
					return
				}
				if tab.Kind == app.DBTabTable {
					DBTableTab(c, a, tab, t)
					return
				}
				DBQueryTab(c, a, tab, t)
			})
		})
		dbStatusBar(c, a, t)
	})
}

func dbToolbar(c *ui.Context, a *app.App, t *ui.Theme) {
	active := a.ActiveDBTab()
	ui.Row(c).Padding(6, 10).Gap(6).AlignItems(ui.Center).Children(func() {
		run := ui.PrimaryButton(c, "执行")
		if active == nil || active.Kind != app.DBTabQuery || active.Running {
			run = run.Disabled(true)
		}
		run.OnClick(a.RunActiveDBTab)

		ui.Button(c, "刷新结构").OnClick(a.RefreshDBSchema)

		exportLabel := "导出 CSV"
		if active != nil && active.Kind == app.DBTabTable {
			exportLabel = "导出本页"
		}
		export := ui.Button(c, exportLabel)
		if active == nil || active.Result == nil || len(active.Result.Columns) == 0 ||
			active.Running ||
			(active.Kind == app.DBTabTable && active.ResultPage != active.Page) {
			export = export.Disabled(true)
		}
		export.OnClick(a.ExportActiveDBResult)

		ui.Spacer(c)
		ui.Text(c, dbConnectionLabel(a)).FontSize(12).Bold().SingleLine()
	})
}

func dbTabStrip(c *ui.Context, a *app.App, t *ui.Theme) {
	ui.Row(c).Padding(4, 6, 0, 6).Gap(6).AlignItems(ui.End).
		Background(t.Surface).Children(func() {
		ui.ScrollHorizontal(c).Grow(1).Height(32).Children(func() {
			ui.Row(c).Gap(2).AlignItems(ui.End).Children(func() {
				for i, tab := range a.DB.Tabs {
					i, tab := i, tab
					active := i == a.DB.Active
					box := ui.Row(c).Height(30).Radius(6, 6, 0, 0).
						AlignItems(ui.Center)
					if active {
						box = box.Background(t.Background).
							BorderWidth(1, 1, 0, 1).BorderColor(t.Border)
					}
					box.Children(func() {
						title := ui.ButtonBase(c).Grow(1).Height(30).
							Padding(4, 8).Radius(6, 6, 0, 0)
						title.Children(func() {
							if tab.Kind == app.DBTabTable {
								if s := resources.ObjectSVG("table"); s != nil {
									ui.Icon(c, s).Size(12, 12)
								}
							}
							label := ui.Text(c, tab.Title).FontSize(12).SingleLine().
								MaxWidth(160)
							if active {
								label.Bold()
							} else {
								label.TextColor(t.TextMuted)
							}
						})
						title.OnClick(func() { a.ActivateDBTab(i) })

						close := ui.ButtonBase(c).Size(18, 18).Radius(4).
							Tooltip("关闭标签")
						close.Children(func() {
							ui.Text(c, "×").FontSize(12).TextColor(t.TextMuted)
						})
						close.OnClick(func() { a.CloseDBTab(tab.Key) })
					})
				}

				plus := ui.ButtonBase(c).Size(28, 28).Radius(6).
					Tooltip("新建查询")
				plus.Children(func() {
					ui.Text(c, "+").FontSize(17).TextColor(t.TextMuted)
				})
				plus.OnClick(a.NewDBQueryTab)
			})
		})
	})
}

func dbStatusBar(c *ui.Context, a *app.App, t *ui.Theme) {
	ui.Divider(c)
	ui.Row(c).Padding(3, 10).Gap(12).AlignItems(ui.Center).Children(func() {
		if a.DB.HostID != "" {
			ui.Text(c, dbConnectionLabel(a)).FontSize(11).TextColor(t.TextMuted).SingleLine()
		}
		ui.Spacer(c)
		if a.DB.Err != "" {
			ui.Text(c, a.DB.Err).FontSize(11).TextColor(t.Danger).SingleLine()
			return
		}
		text := a.DB.Notice
		if text == "" {
			text = "就绪"
		}
		ui.Text(c, text).FontSize(11).TextColor(t.TextMuted).SingleLine()
	})
}

func dbResultMeta(c *ui.Context, tab *app.DBTab, t *ui.Theme) {
	ui.Row(c).Gap(10).AlignItems(ui.Center).Children(func() {
		if tab.Result == nil {
			return
		}
		if len(tab.Result.Columns) > 0 {
			ui.Text(c, fmt.Sprintf("%d 行 · %dms", len(tab.Result.Rows), tab.DurationMS)).
				FontSize(11).TextColor(t.TextMuted)
			if tab.Result.Truncated {
				ui.Text(c, "结果已截断").FontSize(11).TextColor(t.Warning)
			}
			return
		}
		ui.Text(c, fmt.Sprintf("影响 %d 行 · %dms", tab.Result.RowsAffected, tab.DurationMS)).
			FontSize(11).TextColor(t.TextMuted)
	})
}

func dbConnectionLabel(a *app.App) string {
	label := a.DB.HostName
	if db := a.DBDatabaseLabel(); db != "" {
		return label + " · " + db
	}
	if a.DB.Protocol != "" {
		return label + " · " + a.DB.Protocol
	}
	return label
}

func dbErrorText(c *ui.Context, msg string, t *ui.Theme) {
	ui.Text(c, msg).FontSize(12).TextColor(t.Danger).Wrap()
}

func dbHint(c *ui.Context, text string, t *ui.Theme) {
	ui.Text(c, text).FontSize(12).TextColor(t.TextMuted)
}

func dbPageLabel(page, size int, hasTotal bool, total int64) string {
	if hasTotal {
		pages := int((total + int64(size) - 1) / int64(size))
		if pages < 1 {
			pages = 1
		}
		return fmt.Sprintf("第 %d / %d 页", page+1, pages)
	}
	return fmt.Sprintf("第 %d 页", page+1)
}

func dbTableCanNext(tab *app.DBTab) bool {
	if tab.Running {
		return false
	}
	if tab.HasTotal {
		return int64((tab.Page+1)*tab.PageSize) < tab.Total
	}
	return tab.Result != nil && len(tab.Result.Rows) >= tab.PageSize
}

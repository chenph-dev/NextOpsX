package db

import (
	"github.com/egoist/mygo/ui"

	"nextopsx/i18n"
	"nextopsx/ui/assets"
	"nextopsx/ui/style"
	"nextopsx/ui/workspace"
)

// Workbench is the native database workbench: schema tree and query tabs.
func Workbench(c *ui.Context, a *workspace.State) {
	t := c.Theme()
	// Ctrl+Enter runs the active query when the editor does not consume
	// the key first; the toolbar button is always available.
	c.OnShortcut(ui.Ctrl, ui.KeyEnter, a.RunActiveDBTab)
	c.OnShortcut(ui.Super, ui.KeyEnter, a.RunActiveDBTab)

	ui.Column(c).Fill().Children(func() {
		ui.Row(c).Grow(1).FillHeight().MinWidth(0).Children(func() {
			ObjectTree(c, a, t)
			ui.Column(c).Grow(1).FillHeight().MinWidth(0).Children(func() {
				dbTabStrip(c, a, t)
				ui.Divider(c)
				tab := a.ActiveDBTab()
				if tab == nil {
					ui.Column(c).Fill().Center().Gap(6).Children(func() {
						ui.Text(c, i18n.T("db.empty.title")).FontSize(style.FontSize(13)).TextColor(t.TextMuted)
						ui.Text(c, i18n.T("db.empty.hint")).
							FontSize(style.FontSize(12)).TextColor(t.TextMuted)
					})
					return
				}
				if tab.Kind == workspace.DBTabTable {
					TableTab(c, a, tab, t)
					return
				}
				QueryTab(c, a, tab, t)
			})
		})
		dbNoticeBar(c, a, t)
	})
}

func dbTabStrip(c *ui.Context, a *workspace.State, t *ui.Theme) {
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
							if tab.Kind == workspace.DBTabTable {
								if s := assets.ObjectSVG("table"); s != nil {
									ui.Icon(c, s).Size(12, 12)
								}
							}
							label := ui.Text(c, tab.Title).FontSize(style.FontSize(13)).SingleLine().
								MaxWidth(160)
							if active {
								label.Bold()
							} else {
								label.TextColor(t.TextMuted)
							}
						})
						title.OnClick(func() { a.ActivateDBTab(i) })

						close := ui.ButtonBase(c).Size(18, 18).Radius(4).
							Tooltip(i18n.T("db.tab.close"))
						close.Children(func() {
							ui.Text(c, "×").FontSize(style.FontSize(13)).TextColor(t.TextMuted)
						})
						close.OnClick(func() { a.CloseDBTab(tab.Key) })
					})
				}

				plus := ui.ButtonBase(c).Size(28, 28).Radius(6).
					Tooltip(i18n.T("db.tab.newQuery"))
				plus.Children(func() {
					ui.Text(c, "+").FontSize(style.FontSize(17)).TextColor(t.TextMuted)
				})
				plus.OnClick(a.NewDBQueryTab)
			})
		})
	})
}

// dbNoticeBar reports export errors and notices, and takes no room when
// there is nothing to say.
func dbNoticeBar(c *ui.Context, a *workspace.State, t *ui.Theme) {
	text, color := a.DB.Notice, t.TextMuted
	if a.DB.Err != "" {
		text, color = a.DB.Err, t.Danger
	}
	if text == "" {
		return
	}
	ui.Divider(c)
	ui.Row(c).Padding(3, 10).AlignItems(ui.Center).Children(func() {
		ui.Text(c, text).FontSize(style.FontSize(12)).TextColor(color).SingleLine()
	})
}

func dbResultMeta(c *ui.Context, tab *workspace.DBTab, t *ui.Theme) {
	ui.Row(c).Gap(10).AlignItems(ui.Center).Children(func() {
		if tab.Result == nil {
			return
		}
		if len(tab.Result.Columns) > 0 {
			ui.Text(c, i18n.Tf("db.result.rows", len(tab.Result.Rows), tab.DurationMS)).
				FontSize(style.FontSize(12)).TextColor(t.TextMuted)
			if tab.Result.Truncated {
				ui.Text(c, i18n.T("db.result.truncated")).FontSize(style.FontSize(12)).TextColor(t.Warning)
			}
			return
		}
		ui.Text(c, i18n.Tf("db.result.affected", tab.Result.RowsAffected, tab.DurationMS)).
			FontSize(style.FontSize(12)).TextColor(t.TextMuted)
	})
}

func dbPageLabel(page, size int, hasTotal bool, total int64) string {
	if hasTotal {
		pages := int((total + int64(size) - 1) / int64(size))
		if pages < 1 {
			pages = 1
		}
		return i18n.Tf("db.page.of", page+1, pages)
	}
	return i18n.Tf("db.page.current", page+1)
}

func dbTableCanNext(tab *workspace.DBTab) bool {
	if tab.Running {
		return false
	}
	if tab.HasTotal {
		return int64((tab.Page+1)*tab.PageSize) < tab.Total
	}
	return tab.Result != nil && len(tab.Result.Rows) >= tab.PageSize
}

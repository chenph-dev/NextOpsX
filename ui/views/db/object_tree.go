package db

import (
	"fmt"
	"strings"

	"github.com/egoist/mygo/ui"

	"nextopsx/domain/dbexec"
	"nextopsx/i18n"
	"nextopsx/ui/assets"
	"nextopsx/ui/style"
	"nextopsx/ui/widgets"
	"nextopsx/ui/workspace"
)

// ObjectTree is the schema tree: tables and views with their columns.
func ObjectTree(c *ui.Context, a *workspace.State, t *ui.Theme) {
	ui.Column(c).Width(240).FillHeight().Background(t.Surface).
		Border(1, t.Border).Children(func() {
		ui.Row(c).Padding(6, 8).Gap(4).AlignItems(ui.Center).Children(func() {
			ui.Text(c, i18n.T("db.tree.objects")).FontSize(style.FontSize(13)).Bold()
			ui.Spacer(c)
			ui.Button(c, i18n.T("db.tree.expand")).FontSize(style.FontSize(11)).OnClick(func() { a.SetDBTreeAll(true) })
			ui.Button(c, i18n.T("db.tree.collapse")).FontSize(style.FontSize(11)).OnClick(func() { a.SetDBTreeAll(false) })
		})
		ui.Divider(c)
		ui.Row(c).Padding(4, 6).Gap(4).AlignItems(ui.Center).Children(func() {
			c.Key("db-tree-filter")
			ui.TextInput(c, &a.DB.Filter).Grow(1).FontSize(style.FontSize(12)).
				Placeholder(i18n.T("db.tree.filterPlaceholder"))
			if a.DB.Filter != "" {
				ui.Button(c, i18n.T("db.tree.clear")).FontSize(style.FontSize(11)).OnClick(func() { a.DB.Filter = "" })
			}
		})
		ui.Divider(c)
		switch {
		case a.DB.SchemaLoading:
			ui.Row(c).Padding(16).Center().Children(func() {
				ui.Spinner(c)
				ui.Text(c, i18n.T("db.tree.loading")).FontSize(style.FontSize(12)).TextColor(t.TextMuted)
			})
		case a.DB.SchemaErr != "":
			ui.Column(c).Padding(8).Children(func() {
				widgets.ErrorText(c, a.DB.SchemaErr, t)
			})
		default:
			tables, views := dbSchemaGroups(a.DB.Schema, a.DB.Filter)
			if len(tables) == 0 && len(views) == 0 {
				ui.Column(c).Padding(8).Gap(6).Children(func() {
					if a.DB.Filter != "" {
						ui.Text(c, i18n.Tf("db.tree.noMatch", a.DB.Filter)).
							FontSize(style.FontSize(12)).TextColor(t.TextMuted).Wrap()
						ui.Button(c, i18n.T("db.tree.clearFilter")).OnClick(func() { a.DB.Filter = "" })
						return
					}
					ui.Text(c, dbEmptySchemaHint(a, a.DB.Protocol)).
						FontSize(style.FontSize(12)).TextColor(t.TextMuted).Wrap()
					ui.Button(c, i18n.T("db.toolbar.refreshSchema")).OnClick(a.RefreshDBSchema)
				})
				return
			}
			ui.Scroll(c).Grow(1).Children(func() {
				ui.Tree(c, func() {
					dbTableGroup(c, a, i18n.T("db.tree.group.tables"), "table", tables)
					dbTableGroup(c, a, i18n.T("db.tree.group.views"), "view", views)
				})
			})
		}
	})
}

func dbSchemaGroups(schema *dbexec.Schema, filter string) (tables, views []dbexec.Table) {
	if schema == nil {
		return nil, nil
	}
	filter = strings.ToLower(strings.TrimSpace(filter))
	for _, tb := range schema.Tables {
		if filter != "" && !strings.Contains(strings.ToLower(tb.Name), filter) {
			continue
		}
		if tb.Type == "view" {
			views = append(views, tb)
		} else {
			tables = append(tables, tb)
		}
	}
	return tables, views
}

func dbEmptySchemaHint(a *workspace.State, _ string) string {
	if db := a.DBDatabaseLabel(); db != "" {
		return i18n.Tf("db.tree.empty.current", db)
	}
	return i18n.T("db.tree.empty.hint")
}

func dbTableGroup(c *ui.Context, a *workspace.State, label, key string, tables []dbexec.Table) {
	item := ui.TreeItem(c, "", a.DBTreeOpen(key), func() {
		for i := range tables {
			dbTableNode(c, a, tables[i])
		}
	})
	item.Children(func() {
		dbTreeIcon(c, key)
		ui.Text(c, fmt.Sprintf("%s (%d)", label, len(tables))).SingleLine()
	})
}

func dbTableNode(c *ui.Context, a *workspace.State, tb dbexec.Table) {
	icon := "table"
	if tb.Type == "view" {
		icon = "view"
	}
	item := ui.TreeItem(c, "", a.DBTreeOpen(workspace.DBTableKey(tb)), func() {
		for _, col := range tb.Columns {
			label := fmt.Sprintf("%s : %s", col.Name, col.DataType)
			colItem := ui.TreeItem(c, "", nil, nil)
			colItem.Children(func() {
				dbTreeIcon(c, "field")
				ui.Text(c, label).SingleLine()
			})
			colItem.Tooltip(label)
		}
	})
	item.Children(func() {
		dbTreeIcon(c, icon)
		ui.Text(c, tb.Name).SingleLine()
	})
	item = item.Tooltip(i18n.T("db.tree.tooltip.browse"))
	item.ContextMenu(func(m *ui.Menu) {
		if m.Item(i18n.T("db.tree.menu.browse")).Chosen() {
			a.OpenDBTable(tb.Name)
		}
		if m.Item(i18n.T("db.tree.menu.generateQuery")).Chosen() {
			a.NewDBQueryTab()
			if tab := a.ActiveDBTab(); tab != nil && tab.Editor != nil {
				tab.Editor.Buffer.Reset(a.DBPreviewQuery(tb.Name))
			}
		}
	})
	if item.DoubleClicked() {
		a.OpenDBTable(tb.Name)
	}
}

func dbTreeIcon(c *ui.Context, name string) {
	if s := assets.ObjectSVG(name); s != nil {
		ui.Icon(c, s).Size(13, 13)
	}
}

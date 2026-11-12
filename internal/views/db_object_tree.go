package views

import (
	"fmt"
	"strings"

	"github.com/egoist/mygo/ui"

	"nextopsx/internal/app"
	"nextopsx/internal/dbexec"
	"nextopsx/resources"
)

// DBObjectTree is the schema tree: tables and views with their columns.
func DBObjectTree(c *ui.Context, a *app.App, t *ui.Theme) {
	ui.Column(c).Width(240).FillHeight().Background(t.Surface).
		Border(1, t.Border).Children(func() {
		ui.Row(c).Padding(6, 8).Gap(4).AlignItems(ui.Center).Children(func() {
			ui.Text(c, "对象").FontSize(12).Bold()
			ui.Spacer(c)
			ui.Button(c, "展开").FontSize(10).OnClick(func() { a.SetDBTreeAll(true) })
			ui.Button(c, "收起").FontSize(10).OnClick(func() { a.SetDBTreeAll(false) })
		})
		ui.Divider(c)
		ui.Row(c).Padding(4, 6).Gap(4).AlignItems(ui.Center).Children(func() {
			c.Key("db-tree-filter")
			ui.TextInput(c, &a.DB.Filter).Grow(1).FontSize(11).
				Placeholder("筛选表/视图")
			if a.DB.Filter != "" {
				ui.Button(c, "清除").FontSize(10).OnClick(func() { a.DB.Filter = "" })
			}
		})
		ui.Divider(c)
		switch {
		case a.DB.SchemaLoading:
			ui.Row(c).Padding(16).Center().Children(func() {
				ui.Spinner(c)
				ui.Text(c, "正在加载结构…").FontSize(11).TextColor(t.TextMuted)
			})
		case a.DB.SchemaErr != "":
			ui.Column(c).Padding(8).Children(func() {
				dbErrorText(c, a.DB.SchemaErr, t)
			})
		default:
			tables, views := dbSchemaGroups(a.DB.Schema, a.DB.Filter)
			if len(tables) == 0 && len(views) == 0 {
				ui.Column(c).Padding(8).Gap(6).Children(func() {
					if a.DB.Filter != "" {
						ui.Text(c, "没有匹配「"+a.DB.Filter+"」的表或视图。").
							FontSize(11).TextColor(t.TextMuted).Wrap()
						ui.Button(c, "清除筛选").OnClick(func() { a.DB.Filter = "" })
						return
					}
					ui.Text(c, dbEmptySchemaHint(a, a.DB.Protocol)).
						FontSize(11).TextColor(t.TextMuted).Wrap()
					ui.Button(c, "刷新结构").OnClick(a.RefreshDBSchema)
				})
				return
			}
			ui.Scroll(c).Grow(1).Children(func() {
				ui.Tree(c, func() {
					dbTableGroup(c, a, "表", "table", tables)
					dbTableGroup(c, a, "视图", "view", views)
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

func dbEmptySchemaHint(a *app.App, _ string) string {
	if db := a.DBDatabaseLabel(); db != "" {
		return "当前库 " + db + " 没有可浏览的表或视图。"
	}
	return "没有可浏览的表或视图；请确认资产已填写数据库名，或点击刷新结构。"
}

func dbTableGroup(c *ui.Context, a *app.App, label, key string, tables []dbexec.Table) {
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

func dbTableNode(c *ui.Context, a *app.App, tb dbexec.Table) {
	icon := "table"
	if tb.Type == "view" {
		icon = "view"
	}
	item := ui.TreeItem(c, "", a.DBTreeOpen(app.DBTableKey(tb)), func() {
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
	item = item.Tooltip("双击浏览数据")
	item.ContextMenu(func(m *ui.Menu) {
		if m.Item("浏览数据").Chosen() {
			a.OpenDBTable(tb.Name)
		}
		if m.Item("生成查询").Chosen() {
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
	if s := resources.ObjectSVG(name); s != nil {
		ui.Icon(c, s).Size(13, 13)
	}
}

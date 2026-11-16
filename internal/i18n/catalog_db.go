package i18n

// Database workbench strings: schema tree, query tabs, result grid and CSV
// export. The two maps below are kept split so each key appears in both
// languages side by side, which makes missing translations easy to spot.
var (
	dbZh = map[string]string{
		// Empty workbench.
		"db.empty.title": "没有打开的标签",
		"db.empty.hint":  "点击标签栏右侧的 + 新建查询。",

		// Toolbar and tabs.
		"db.toolbar.run":           "执行",
		"db.toolbar.refreshSchema": "刷新结构",
		"db.toolbar.exportCSV":     "导出 CSV",
		"db.toolbar.exportPage":    "导出本页",
		"db.tab.close":             "关闭标签",
		"db.tab.newQuery":          "新建查询",
		"db.status.ready":          "就绪",

		// Result summary.
		"db.result.rows":      "%d 行 · %dms",
		"db.result.truncated": "结果已截断",
		"db.result.affected":  "影响 %d 行 · %dms",
		"db.result.column":    "列 %d",
		"db.result.empty":     "(无数据)",
		"db.page.of":          "第 %d / %d 页",
		"db.page.current":     "第 %d 页",

		// Object tree.
		"db.tree.objects":            "对象",
		"db.tree.expand":             "展开",
		"db.tree.collapse":           "收起",
		"db.tree.filterPlaceholder":  "筛选表/视图",
		"db.tree.clear":              "清除",
		"db.tree.loading":            "正在加载结构…",
		"db.tree.noMatch":            "没有匹配「%s」的表或视图。",
		"db.tree.clearFilter":        "清除筛选",
		"db.tree.group.tables":       "表",
		"db.tree.group.views":        "视图",
		"db.tree.empty.current":      "当前库 %s 没有可浏览的表或视图。",
		"db.tree.empty.hint":         "没有可浏览的表或视图；请确认资产已填写数据库名，或点击刷新结构。",
		"db.tree.tooltip.browse":     "双击浏览数据",
		"db.tree.menu.browse":        "浏览数据",
		"db.tree.menu.generateQuery": "生成查询",

		// Query and table tabs.
		"db.query.running":      "执行中…",
		"db.query.empty":        "执行查询后在此显示结果。",
		"db.table.total":        "共 %d 行",
		"db.table.prev":         "上一页",
		"db.table.next":         "下一页",
		"db.table.refresh":      "刷新",
		"db.table.loading":      "加载中…",
		"db.table.loadingHint":  "加载表数据…",
		"db.editor.label":       "SQL 查询",
		"db.editor.placeholder": "输入 SQL，点击「执行」运行",

		// Tab titles and backend errors.
		"db.tab.query":         "查询",
		"db.tab.queryNumbered": "查询 %d",
		"db.backendNotReady":   "后端未初始化",
		"db.schema.loadFailed": "加载结构失败：%s",

		// CSV export.
		"db.export.empty":            "当前没有可导出的查询结果",
		"db.export.loading":          "数据正在加载，请稍后再导出",
		"db.export.dialog.title":     "导出 CSV",
		"db.export.dialog.button":    "导出",
		"db.export.failed":           "导出失败：%s",
		"db.export.csvFailed":        "生成 CSV 失败：%s",
		"db.export.writeFailed":      "写入文件失败：%s",
		"db.export.notice.rows":      "已导出 %d 行",
		"db.export.notice.page":      "已导出第 %d 页 %d 行",
		"db.export.notice.total":     "（共 %d 行，仅当前页）",
		"db.export.notice.truncated": "；结果已截断",
		"db.notice.separator":        "：",
	}

	dbEn = map[string]string{
		// Empty workbench.
		"db.empty.title": "No open tabs",
		"db.empty.hint":  "Click + on the right of the tab bar to create a query.",

		// Toolbar and tabs.
		"db.toolbar.run":           "Run",
		"db.toolbar.refreshSchema": "Refresh schema",
		"db.toolbar.exportCSV":     "Export CSV",
		"db.toolbar.exportPage":    "Export page",
		"db.tab.close":             "Close tab",
		"db.tab.newQuery":          "New query",
		"db.status.ready":          "Ready",

		// Result summary.
		"db.result.rows":      "%d rows · %dms",
		"db.result.truncated": "Results truncated",
		"db.result.affected":  "%d rows affected · %dms",
		"db.result.column":    "Column %d",
		"db.result.empty":     "(No data)",
		"db.page.of":          "Page %d / %d",
		"db.page.current":     "Page %d",

		// Object tree.
		"db.tree.objects":            "Objects",
		"db.tree.expand":             "Expand",
		"db.tree.collapse":           "Collapse",
		"db.tree.filterPlaceholder":  "Filter tables/views",
		"db.tree.clear":              "Clear",
		"db.tree.loading":            "Loading schema…",
		"db.tree.noMatch":            "No tables or views match \"%s\".",
		"db.tree.clearFilter":        "Clear filter",
		"db.tree.group.tables":       "Tables",
		"db.tree.group.views":        "Views",
		"db.tree.empty.current":      "Database %s has no browsable tables or views.",
		"db.tree.empty.hint":         "No browsable tables or views; make sure the asset has a database name, or click Refresh schema.",
		"db.tree.tooltip.browse":     "Double-click to browse data",
		"db.tree.menu.browse":        "Browse data",
		"db.tree.menu.generateQuery": "Generate query",

		// Query and table tabs.
		"db.query.running":      "Running…",
		"db.query.empty":        "Results appear here after you run a query.",
		"db.table.total":        "%d rows",
		"db.table.prev":         "Previous",
		"db.table.next":         "Next",
		"db.table.refresh":      "Refresh",
		"db.table.loading":      "Loading…",
		"db.table.loadingHint":  "Loading table data…",
		"db.editor.label":       "SQL query",
		"db.editor.placeholder": "Enter SQL and click Run",

		// Tab titles and backend errors.
		"db.tab.query":         "Query",
		"db.tab.queryNumbered": "Query %d",
		"db.backendNotReady":   "Backend not initialized",
		"db.schema.loadFailed": "Failed to load schema: %s",

		// CSV export.
		"db.export.empty":            "No query results to export",
		"db.export.loading":          "Data is still loading; try exporting again shortly",
		"db.export.dialog.title":     "Export CSV",
		"db.export.dialog.button":    "Export",
		"db.export.failed":           "Export failed: %s",
		"db.export.csvFailed":        "Failed to build CSV: %s",
		"db.export.writeFailed":      "Failed to write file: %s",
		"db.export.notice.rows":      "Exported %d rows",
		"db.export.notice.page":      "Exported page %d (%d rows)",
		"db.export.notice.total":     " (%d rows total; current page only)",
		"db.export.notice.truncated": "; results truncated",
		"db.notice.separator":        ": ",
	}
)

func init() {
	add(ZhCN, dbZh)
	add(EnUS, dbEn)
}

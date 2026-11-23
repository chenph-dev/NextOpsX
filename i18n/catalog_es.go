package i18n

// Elasticsearch workbench UI strings: index browser, DSL query, cluster info,
// REST command terminal and app-layer ES errors/notices.
var (
	esZh = map[string]string{
		// Tabs and toolbar.
		"es.tab.query":             "DSL 查询",
		"es.tab.info":              "集群信息",
		"es.action.refreshIndices": "刷新索引",
		"es.action.refresh":        "刷新",
		"es.action.retry":          "重试",
		"es.action.runQuery":       "执行查询",
		"es.action.execute":        "执行",

		// Index browser.
		"es.indices.title":   "索引 (%d)",
		"es.indices.loading": "读取索引…",
		"es.indices.empty":   "没有索引",

		// DSL query tab.
		"es.query.noIndex": "索引：未选择",
		"es.query.index":   "索引：%s",
		"es.query.running": "查询中…",
		"es.query.hint":    "选择索引并输入 Query DSL 后执行。",

		// Cluster info tab.
		"es.info.title":   "集群信息",
		"es.info.loading": "读取集群信息…",
		"es.info.empty":   "没有可展示的集群指标。",

		// REST command terminal.
		"es.command.title":        "REST 命令终端",
		"es.command.historyCount": "历史 (%d)",
		"es.command.running":      "执行中…",
		"es.command.hint":         "输入 REST 命令；首行为路径，后续行为 JSON body。",
		"es.command.placeholder":  "_cat/indices?format=json\n或 logs-*/_search\n{\"query\":{\"match_all\":{}}}",
		"es.command.historyTitle": "命令历史",
		"es.command.historyEmpty": "暂无历史",

		// App-layer errors and notices.
		"es.error.backendUninitialized": "后端未初始化",
		"es.error.indicesFailed":        "读取索引失败：%s",
		"es.error.selectIndex":          "请先选择索引",
		"es.error.enterDSL":             "请输入 Query DSL",
		"es.error.infoFailed":           "读取集群信息失败：%s",
	}

	esEn = map[string]string{
		// Tabs and toolbar.
		"es.tab.query":             "DSL query",
		"es.tab.info":              "Cluster info",
		"es.action.refreshIndices": "Refresh indices",
		"es.action.refresh":        "Refresh",
		"es.action.retry":          "Retry",
		"es.action.runQuery":       "Run query",
		"es.action.execute":        "Run",

		// Index browser.
		"es.indices.title":   "Indices (%d)",
		"es.indices.loading": "Loading indices…",
		"es.indices.empty":   "No indices",

		// DSL query tab.
		"es.query.noIndex": "Index: none selected",
		"es.query.index":   "Index: %s",
		"es.query.running": "Querying…",
		"es.query.hint":    "Select an index, enter a Query DSL and run it.",

		// Cluster info tab.
		"es.info.title":   "Cluster info",
		"es.info.loading": "Loading cluster info…",
		"es.info.empty":   "No cluster metrics to show.",

		// REST command terminal.
		"es.command.title":        "REST command terminal",
		"es.command.historyCount": "History (%d)",
		"es.command.running":      "Running…",
		"es.command.hint":         "Enter a REST command; the first line is the path, following lines are the JSON body.",
		"es.command.placeholder":  "_cat/indices?format=json\nor logs-*/_search\n{\"query\":{\"match_all\":{}}}",
		"es.command.historyTitle": "Command history",
		"es.command.historyEmpty": "No history",

		// App-layer errors and notices.
		"es.error.backendUninitialized": "Backend not initialized",
		"es.error.indicesFailed":        "Failed to read indices: %s",
		"es.error.selectIndex":          "Select an index first",
		"es.error.enterDSL":             "Enter a Query DSL",
		"es.error.infoFailed":           "Failed to read cluster info: %s",
	}
)

func init() {
	add(ZhCN, esZh)
	add(EnUS, esEn)
}

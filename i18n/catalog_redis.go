package i18n

// Redis workbench UI strings: key browser, detail/info tabs, command
// terminal, dialogs and app-layer Redis errors/notices.
var (
	redisZh = map[string]string{
		// Tabs and toolbar.
		"redis.tab.detail":           "键详情",
		"redis.tab.info":             "服务器信息",
		"redis.action.refreshKeys":   "刷新键",
		"redis.action.refresh":       "刷新",
		"redis.action.search":        "搜索",
		"redis.action.prevPage":      "上一页",
		"redis.action.nextPage":      "下一页",
		"redis.action.retry":         "重试",
		"redis.action.execute":       "执行",
		"redis.action.editValue":     "编辑值",
		"redis.action.setTTL":        "设置 TTL",
		"redis.action.delete":        "删除",
		"redis.action.cancel":        "取消",
		"redis.action.save":          "保存",
		"redis.action.confirmDelete": "确认删除",

		// Key browser.
		"redis.keys.title":              "键 (%d)",
		"redis.keys.cursor":             "游标 %s",
		"redis.keys.patternPlaceholder": "匹配模式，如 user:*",
		"redis.keys.scanning":           "扫描中…",
		"redis.keys.empty":              "没有匹配的键",
		"redis.pager.page":              "第 %d 页",
		"redis.pager.pageSize":          "每页 %d",

		// Key detail.
		"redis.detail.empty":       "选择左侧的键查看详情",
		"redis.detail.type":        "类型 %s",
		"redis.detail.truncated":   "仅显示前 200 项",
		"redis.detail.loading":     "读取键…",
		"redis.detail.unsupported": "该类型暂不支持预览。",
		"redis.ttl.permanent":      "永久",
		"redis.ttl.missing":        "不存在",

		// Server info tab.
		"redis.info.title":   "服务器信息",
		"redis.info.loading": "读取 INFO…",
		"redis.info.empty":   "没有可展示的 INFO 指标。",

		// Command terminal.
		"redis.command.title":        "命令终端",
		"redis.command.historyCount": "历史 (%d)",
		"redis.command.running":      "执行中…",
		"redis.command.hint":         "输入 Redis 命令，Enter 执行。",
		"redis.command.historyTitle": "命令历史",
		"redis.command.historyEmpty": "暂无历史",

		// Dialogs.
		"redis.dialog.editTitle":   "编辑字符串值",
		"redis.dialog.ttlTitle":    "设置 TTL",
		"redis.dialog.ttlSeconds":  "秒数",
		"redis.dialog.ttlHint":     "-1 表示永久",
		"redis.dialog.deleteTitle": "删除键",
		"redis.dialog.deleteBody":  "确定删除「%s」吗？此操作无法撤销。",

		// App-layer errors and notices.
		"redis.error.backendUninitialized": "后端未初始化",
		"redis.error.scanFailed":           "扫描键失败：%s",
		"redis.error.deleteFailed":         "删除失败：%s",
		"redis.error.saveFailed":           "保存失败：%s",
		"redis.error.keyNotReady":          "键信息尚未加载完成",
		"redis.error.ttlNotInteger":        "TTL 需要是整数秒",
		"redis.error.ttlZero":              "TTL 0 会立即删除键；请使用删除按钮，或输入正数/-1",
		"redis.error.ttlFailed":            "设置 TTL 失败：%s",
		"redis.error.infoFailed":           "读取 INFO 失败：%s",
		"redis.notice.deleted":             "已删除 %s",
		"redis.notice.saved":               "已保存 %s",
		"redis.notice.ttlUpdated":          "已更新 TTL：%s",
	}

	redisEn = map[string]string{
		// Tabs and toolbar.
		"redis.tab.detail":           "Key detail",
		"redis.tab.info":             "Server info",
		"redis.action.refreshKeys":   "Refresh keys",
		"redis.action.refresh":       "Refresh",
		"redis.action.search":        "Search",
		"redis.action.prevPage":      "Previous page",
		"redis.action.nextPage":      "Next page",
		"redis.action.retry":         "Retry",
		"redis.action.execute":       "Run",
		"redis.action.editValue":     "Edit value",
		"redis.action.setTTL":        "Set TTL",
		"redis.action.delete":        "Delete",
		"redis.action.cancel":        "Cancel",
		"redis.action.save":          "Save",
		"redis.action.confirmDelete": "Confirm delete",

		// Key browser.
		"redis.keys.title":              "Keys (%d)",
		"redis.keys.cursor":             "Cursor %s",
		"redis.keys.patternPlaceholder": "Match pattern, e.g. user:*",
		"redis.keys.scanning":           "Scanning…",
		"redis.keys.empty":              "No matching keys",
		"redis.pager.page":              "Page %d",
		"redis.pager.pageSize":          "%d per page",

		// Key detail.
		"redis.detail.empty":       "Select a key on the left to see its details",
		"redis.detail.type":        "Type %s",
		"redis.detail.truncated":   "Showing the first 200 items only",
		"redis.detail.loading":     "Reading key…",
		"redis.detail.unsupported": "Preview is not supported for this type.",
		"redis.ttl.permanent":      "Permanent",
		"redis.ttl.missing":        "Missing",

		// Server info tab.
		"redis.info.title":   "Server info",
		"redis.info.loading": "Reading INFO…",
		"redis.info.empty":   "No INFO metrics to show.",

		// Command terminal.
		"redis.command.title":        "Command terminal",
		"redis.command.historyCount": "History (%d)",
		"redis.command.running":      "Running…",
		"redis.command.hint":         "Enter a Redis command and press Enter to run.",
		"redis.command.historyTitle": "Command history",
		"redis.command.historyEmpty": "No history",

		// Dialogs.
		"redis.dialog.editTitle":   "Edit string value",
		"redis.dialog.ttlTitle":    "Set TTL",
		"redis.dialog.ttlSeconds":  "Seconds",
		"redis.dialog.ttlHint":     "-1 means permanent",
		"redis.dialog.deleteTitle": "Delete key",
		"redis.dialog.deleteBody":  "Delete \"%s\"? This cannot be undone.",

		// App-layer errors and notices.
		"redis.error.backendUninitialized": "Backend not initialized",
		"redis.error.scanFailed":           "Failed to scan keys: %s",
		"redis.error.deleteFailed":         "Delete failed: %s",
		"redis.error.saveFailed":           "Save failed: %s",
		"redis.error.keyNotReady":          "Key information has not loaded yet",
		"redis.error.ttlNotInteger":        "TTL must be an integer number of seconds",
		"redis.error.ttlZero":              "TTL 0 deletes the key immediately; use the delete button, or enter a positive number/-1",
		"redis.error.ttlFailed":            "Failed to set TTL: %s",
		"redis.error.infoFailed":           "Failed to read INFO: %s",
		"redis.notice.deleted":             "Deleted %s",
		"redis.notice.saved":               "Saved %s",
		"redis.notice.ttlUpdated":          "TTL updated: %s",
	}
)

func init() {
	add(ZhCN, redisZh)
	add(EnUS, redisEn)
}

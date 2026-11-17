package i18n

// Command history: terminal picker, per-asset history entries and favorites.
var (
	historyZh = map[string]string{
		"history.title":        "命令历史",
		"history.tab":          "历史",
		"history.favoritesTab": "收藏",
		"history.search":       "搜索命令、名称或说明",
		"history.emptyTitle":   "还没有命令历史",
		"history.emptyHint":    "智能运维执行命令或运行收藏后，会记录在这里。",
		"history.noMatch":      "没有匹配的命令",
		"history.output":       "输出",
		"history.hideOutput":   "收起输出",
		"history.insert":       "插入",
		"history.run":          "执行",
		"history.favorite":     "收藏",
		"history.retry":        "重试",

		"history.source.history":  "历史",
		"history.source.favorite": "收藏",
		"history.source.both":     "收藏 · 历史",

		"history.err.load": "读取命令历史失败：%s",
		"history.err.save": "记录命令历史失败：%s",

		"history.favoriteGroup":       "来自历史",
		"history.favoriteDescription": "从命令历史收藏",
		"history.favoriteFallback":    "命令",
		"history.suggestTitle":        "命令建议",
		"history.suggestCount":        "%d 条",
		"history.bubbleHint":          "↑↓ 选择 · Tab 填入 · Esc 关闭",
		"history.pickerHint":          "Esc 关闭 · 点击命令插入",
		"history.close":               "关闭",
		"history.filterAll":           "全部",
		"history.filterFavorites":     "收藏",
		"history.filterHistory":       "历史",
		"history.groupFavorites":      "收藏",
		"history.groupToday":          "今天",
		"history.groupYesterday":      "昨天",
		"history.groupEarlier":        "更早",
		"history.justNow":             "刚刚",
		"history.minutesAgo":          "%d 分钟前",
		"history.todayAt":             "今天 %s",
		"history.yesterdayAt":         "昨天 %s",
		"history.currentHost":         "仅当前主机 · %s",
	}

	historyEn = map[string]string{
		"history.title":        "Command history",
		"history.tab":          "History",
		"history.favoritesTab": "Favorites",
		"history.search":       "Search commands, names or notes",
		"history.emptyTitle":   "No command history yet",
		"history.emptyHint":    "Commands run by AI or from favorites appear here.",
		"history.noMatch":      "No matching commands",
		"history.output":       "Output",
		"history.hideOutput":   "Hide output",
		"history.insert":       "Insert",
		"history.run":          "Run",
		"history.favorite":     "Favorite",
		"history.retry":        "Retry",

		"history.source.history":  "History",
		"history.source.favorite": "Favorite",
		"history.source.both":     "Favorite · History",

		"history.err.load": "Failed to load command history: %s",
		"history.err.save": "Failed to record command history: %s",

		"history.favoriteGroup":       "From history",
		"history.favoriteDescription": "Favorited from command history",
		"history.favoriteFallback":    "Command",
		"history.suggestTitle":        "Suggestions",
		"history.suggestCount":        "%d matches",
		"history.bubbleHint":          "↑↓ select · Tab insert · Esc close",
		"history.pickerHint":          "Esc close · click a command to insert",
		"history.close":               "Close",
		"history.filterAll":           "All",
		"history.filterFavorites":     "Favorites",
		"history.filterHistory":       "History",
		"history.groupFavorites":      "Favorites",
		"history.groupToday":          "Today",
		"history.groupYesterday":      "Yesterday",
		"history.groupEarlier":        "Earlier",
		"history.justNow":             "just now",
		"history.minutesAgo":          "%d min ago",
		"history.todayAt":             "Today %s",
		"history.yesterdayAt":         "Yesterday %s",
		"history.currentHost":         "Current host · %s",
	}
)

func init() {
	add(ZhCN, historyZh)
	add(EnUS, historyEn)
}

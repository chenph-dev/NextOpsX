package i18n

// Command history: terminal picker, per-asset history entries and favorites.
var (
	historyZh = map[string]string{
		"history.title":        "命令历史",
		"history.tab":          "历史",
		"history.favoritesTab": "收藏",
		"history.search":       "搜索历史命令",
		"history.emptyTitle":   "还没有命令历史",
		"history.emptyHint":    "AI 执行命令或运行收藏后，会记录在这里。",
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
	}

	historyEn = map[string]string{
		"history.title":        "Command history",
		"history.tab":          "History",
		"history.favoritesTab": "Favorites",
		"history.search":       "Search command history",
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
	}
)

func init() {
	add(ZhCN, historyZh)
	add(EnUS, historyEn)
}

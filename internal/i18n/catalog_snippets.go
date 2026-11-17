package i18n

// Command snippet library: settings section, editor and terminal picker.
var (
	snippetsZh = map[string]string{
		"settings.nav.snippets": "收藏",
		"shell.history":         "命令",

		"snippets.title":      "收藏命令",
		"snippets.new":        "新建收藏",
		"snippets.count":      "共 %d 条收藏",
		"snippets.search":     "搜索收藏",
		"snippets.emptyTitle": "还没有收藏命令",
		"snippets.emptyHint":  "从命令历史收藏，或手动新建一条。",
		"snippets.selectHint": "选择左侧收藏命令进行编辑，或新建一条。",

		"snippets.field.name":                   "名称",
		"snippets.field.group":                  "分组",
		"snippets.field.command":                "命令",
		"snippets.field.description":            "说明",
		"snippets.field.namePlaceholder":        "如：查看磁盘占用",
		"snippets.field.groupPlaceholder":       "如：排障",
		"snippets.field.commandPlaceholder":     "如：df -h",
		"snippets.field.descriptionPlaceholder": "可选说明",
		"snippets.variables":                    "可用变量：{{name}}、{{host}}、{{user}}、{{port}}、{{protocol}}",
		"snippets.preview":                      "预览：%s",

		"snippets.editTitle":     "编辑收藏",
		"snippets.newTitle":      "新建收藏",
		"snippets.save":          "保存",
		"snippets.cancel":        "取消",
		"snippets.delete":        "删除",
		"snippets.confirmDelete": "确认删除",
		"snippets.deleteTitle":   "删除收藏命令",
		"snippets.deleteConfirm": "确定删除收藏「%s」吗？",

		"snippets.err.loadFailed":   "读取命令片段失败：%s",
		"snippets.err.saveFailed":   "保存片段失败：%s",
		"snippets.err.deleteFailed": "删除片段失败：%s",
		"snippets.err.required":     "请填写名称和命令",
		"snippets.err.notConnected": "请先连接 SSH 终端",

		"snippets.picker.title":        "收藏",
		"snippets.picker.manage":       "管理收藏",
		"snippets.picker.insert":       "插入",
		"snippets.picker.run":          "执行",
		"snippets.picker.empty":        "没有匹配的收藏",
		"snippets.picker.noGroup":      "未分组",
		"snippets.picker.connectFirst": "请先连接 SSH 终端",
	}

	snippetsEn = map[string]string{
		"settings.nav.snippets": "Favorites",
		"shell.history":         "Commands",

		"snippets.title":      "Favorite commands",
		"snippets.new":        "New favorite",
		"snippets.count":      "%d favorites",
		"snippets.search":     "Search favorites",
		"snippets.emptyTitle": "No favorite commands yet",
		"snippets.emptyHint":  "Favorite commands from history, or create one manually.",
		"snippets.selectHint": "Select a favorite command to edit, or create one.",

		"snippets.field.name":                   "Name",
		"snippets.field.group":                  "Group",
		"snippets.field.command":                "Command",
		"snippets.field.description":            "Description",
		"snippets.field.namePlaceholder":        "e.g. Check disk usage",
		"snippets.field.groupPlaceholder":       "e.g. Troubleshooting",
		"snippets.field.commandPlaceholder":     "e.g. df -h",
		"snippets.field.descriptionPlaceholder": "Optional description",
		"snippets.variables":                    "Variables: {{name}}, {{host}}, {{user}}, {{port}}, {{protocol}}",
		"snippets.preview":                      "Preview: %s",

		"snippets.editTitle":     "Edit favorite",
		"snippets.newTitle":      "New favorite",
		"snippets.save":          "Save",
		"snippets.cancel":        "Cancel",
		"snippets.delete":        "Delete",
		"snippets.confirmDelete": "Delete",
		"snippets.deleteTitle":   "Delete favorite command",
		"snippets.deleteConfirm": "Delete favorite \"%s\"?",

		"snippets.err.loadFailed":   "Failed to load snippets: %s",
		"snippets.err.saveFailed":   "Failed to save snippet: %s",
		"snippets.err.deleteFailed": "Failed to delete snippet: %s",
		"snippets.err.required":     "Name and command are required",
		"snippets.err.notConnected": "Connect an SSH terminal first",

		"snippets.picker.title":        "Favorites",
		"snippets.picker.manage":       "Manage favorites",
		"snippets.picker.insert":       "Insert",
		"snippets.picker.run":          "Run",
		"snippets.picker.empty":        "No matching favorites",
		"snippets.picker.noGroup":      "Ungrouped",
		"snippets.picker.connectFirst": "Connect an SSH terminal first",
	}
)

func init() {
	add(ZhCN, snippetsZh)
	add(EnUS, snippetsEn)
}

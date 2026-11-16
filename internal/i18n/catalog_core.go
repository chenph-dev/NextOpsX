package i18n

// Core UI strings: window shell, menus, tray and general settings.
// The two maps below are kept split so each key appears in both languages
// side by side, which makes missing translations easy to spot.
var (
	coreZh = map[string]string{
		// Shell.
		"shell.assets":        "资产",
		"shell.newAsset":      "＋资产",
		"shell.newFolder":     "＋目录",
		"shell.newAssetMenu":  "新建资产",
		"shell.newFolderMenu": "新建目录",
		"shell.refresh":       "刷新",
		"shell.settings":      "设置",
		"shell.remoteDesktop": "远程桌面",
		"shell.tab.terminal":  "终端",
		"shell.tab.ai":        "AI 对话",
		"shell.tab.database":  "数据库",
		"shell.tab.overview":  "概览",
		"shell.folder":        "目录",

		// Settings navigation and general section.
		"settings.title":              "设置",
		"settings.back":               "返回",
		"settings.nav.general":        "通用",
		"settings.nav.model":          "模型",
		"settings.nav.skills":         "技能",
		"settings.nav.policy":         "审批",
		"settings.nav.audit":          "审计",
		"settings.general.appearance": "外观",
		"settings.general.language":   "界面语言",
		"settings.general.theme":      "主题",
		"settings.theme.system":       "跟随系统",
		"settings.theme.light":        "浅色",
		"settings.theme.dark":         "深色",
		"settings.general.scale":      "界面缩放",
		"settings.general.hint":       "默认跟随系统外观；界面缩放会同时放大文字与控件（90%–125%）。",
		"settings.general.data":       "数据",
		"settings.general.dataDir":    "数据目录",
		"settings.general.dataHint":   "数据库、技能文件与密钥保存在此目录。",
		"settings.general.saveFailed": "保存界面设置失败：%s",

		// Menus.
		"menu.file":             "文件",
		"menu.newAsset":         "新建资产",
		"menu.newFolder":        "新建目录",
		"menu.settings":         "设置…",
		"menu.view":             "视图",
		"menu.assets":           "资产",
		"menu.settings.general": "设置 · 常规",
		"menu.settings.model":   "设置 · 模型",
		"menu.settings.skills":  "设置 · 技能",
		"menu.settings.policy":  "设置 · 审批",
		"menu.settings.audit":   "设置 · 审计",

		// Tray.
		"tray.open":     "打开 NextOpsX",
		"tray.settings": "设置…",

		// App-level errors shown in the shell.
		"app.readUISettingsFailed": "读取界面设置失败：%s",
		"app.connectFailed":        "连接失败: %s",
		"app.terminalFailed":       "创建终端失败: %s",
	}

	coreEn = map[string]string{
		// Shell.
		"shell.assets":        "Assets",
		"shell.newAsset":      "+Asset",
		"shell.newFolder":     "+Folder",
		"shell.newAssetMenu":  "New asset",
		"shell.newFolderMenu": "New folder",
		"shell.refresh":       "Refresh",
		"shell.settings":      "Settings",
		"shell.remoteDesktop": "Remote desktop",
		"shell.tab.terminal":  "Terminal",
		"shell.tab.ai":        "AI chat",
		"shell.tab.database":  "Database",
		"shell.tab.overview":  "Overview",
		"shell.folder":        "Folder",

		// Settings navigation and general section.
		"settings.title":              "Settings",
		"settings.back":               "Back",
		"settings.nav.general":        "General",
		"settings.nav.model":          "Models",
		"settings.nav.skills":         "Skills",
		"settings.nav.policy":         "Approval",
		"settings.nav.audit":          "Audit",
		"settings.general.appearance": "Appearance",
		"settings.general.language":   "Language",
		"settings.general.theme":      "Theme",
		"settings.theme.system":       "System",
		"settings.theme.light":        "Light",
		"settings.theme.dark":         "Dark",
		"settings.general.scale":      "UI scale",
		"settings.general.hint":       "Follows the system appearance by default; UI scale enlarges text and controls together (90%–125%).",
		"settings.general.data":       "Data",
		"settings.general.dataDir":    "Data directory",
		"settings.general.dataHint":   "The database, skill files and keys are stored in this directory.",
		"settings.general.saveFailed": "Failed to save interface settings: %s",

		// Menus.
		"menu.file":             "File",
		"menu.newAsset":         "New asset",
		"menu.newFolder":        "New folder",
		"menu.settings":         "Settings…",
		"menu.view":             "View",
		"menu.assets":           "Assets",
		"menu.settings.general": "Settings · General",
		"menu.settings.model":   "Settings · Models",
		"menu.settings.skills":  "Settings · Skills",
		"menu.settings.policy":  "Settings · Approval",
		"menu.settings.audit":   "Settings · Audit",

		// Tray.
		"tray.open":     "Open NextOpsX",
		"tray.settings": "Settings…",

		// App-level errors shown in the shell.
		"app.readUISettingsFailed": "Failed to read interface settings: %s",
		"app.connectFailed":        "Connection failed: %s",
		"app.terminalFailed":       "Failed to create terminal: %s",
	}
)

func init() {
	add(ZhCN, coreZh)
	add(EnUS, coreEn)
}

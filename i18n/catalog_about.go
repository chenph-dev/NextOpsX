package i18n

// About section and automatic-update strings.
var (
	aboutZh = map[string]string{
		"settings.nav.about": "关于",

		"settings.about.title":         "关于",
		"settings.about.version":       "版本",
		"settings.about.devVersion":    "开发版",
		"settings.about.updates":       "自动更新",
		"settings.about.check":         "检查更新",
		"settings.about.checking":      "正在检查…",
		"settings.about.upToDate":      "已是最新版本",
		"settings.about.available":     "发现新版本 %s",
		"settings.about.notes":         "更新说明",
		"settings.about.install":       "下载并安装",
		"settings.about.installing":    "正在下载…",
		"settings.about.progress":      "已下载 %d%%",
		"settings.about.installed":     "更新已安装，重启后生效",
		"settings.about.relaunch":      "重启并应用更新",
		"settings.about.disabled":      "当前构建或安装方式不支持自动更新",
		"settings.about.disabledHint":  "开发版、由系统包管理器安装的版本或未配置更新渠道的构建无法自更新；安装到用户目录的版本可从 GitHub Releases 更新。",
		"settings.about.checkFailed":   "检查更新失败：%s",
		"settings.about.installFailed": "安装更新失败：%s",

		"menu.settings.about": "设置 · 关于",
	}

	aboutEn = map[string]string{
		"settings.nav.about": "About",

		"settings.about.title":         "About",
		"settings.about.version":       "Version",
		"settings.about.devVersion":    "dev",
		"settings.about.updates":       "Automatic updates",
		"settings.about.check":         "Check for updates",
		"settings.about.checking":      "Checking…",
		"settings.about.upToDate":      "You have the latest version",
		"settings.about.available":     "Version %s is available",
		"settings.about.notes":         "Release notes",
		"settings.about.install":       "Download and install",
		"settings.about.installing":    "Downloading…",
		"settings.about.progress":      "Downloaded %d%%",
		"settings.about.installed":     "Update installed; restart to apply it",
		"settings.about.relaunch":      "Restart and apply update",
		"settings.about.disabled":      "This build or installation cannot update itself",
		"settings.about.disabledHint":  "Development builds, versions installed by a system package manager, and builds without an update feed cannot update themselves; versions installed in your user directory update from GitHub Releases.",
		"settings.about.checkFailed":   "Failed to check for updates: %s",
		"settings.about.installFailed": "Failed to install the update: %s",

		"menu.settings.about": "Settings · About",
	}
)

func init() {
	add(ZhCN, aboutZh)
	add(EnUS, aboutEn)
}

package main

import (
	_ "embed"
	"log"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"

	"nextopsx/internal/app"
	"nextopsx/internal/backend"
	"nextopsx/internal/i18n"
	"nextopsx/internal/views"
)

//go:embed resources/tray.png
var trayIcon []byte

// appTray keeps the tray icon alive for the app's lifetime.
var appTray *mygo.Tray

func main() {
	b, err := backend.New()
	if err != nil {
		log.Fatal(err)
	}
	a, err := app.New(b)
	if err != nil {
		b.Close()
		log.Fatal(err)
	}

	installMenu(a)

	mygo.App.WhenReady(func() {
		win := mygo.NewWindow(mygo.WindowOptions{
			Title:     "NextOpsX",
			Width:     1200,
			Height:    800,
			MinWidth:  900,
			MinHeight: 600,
			StateKey:  "main",
			// 默认隐藏原生菜单栏，Alt / F10 可临时唤出；快捷键始终有效。
			AutoHideMenuBar: true,
			Content:         ui.View(views.Shell(a)),
		})
		a.Attach(win)

		// 桌面通知：窗口在前台时不打扰。
		a.SetNotifier(func(title, body string) {
			if win.IsFocused() || !mygo.NotificationsSupported() {
				return
			}
			n := mygo.NewNotification(mygo.NotificationOptions{Title: title, Body: body})
			if err := n.Show(); err != nil {
				log.Printf("notification: %v", err)
			}
		})

		// 托盘：Linux 需要 libayatana-appindicator3，缺失时只记录错误。
		tray, err := mygo.NewTray(mygo.TrayOptions{
			Icon:    trayIcon,
			ToolTip: "NextOpsX",
			Menu:    trayMenu(a, win),
		})
		if err != nil {
			log.Println("tray:", err)
		} else {
			appTray = tray
		}

		// 语言切换后重建菜单栏与托盘菜单。
		a.SetChromeRebuilder(func() {
			installMenu(a)
			if appTray != nil {
				appTray.SetMenu(trayMenu(a, win))
			}
		})

		win.OnClosed(func() {
			a.Disconnect()
			a.WaitForWrites() // 主题等短写完成后才能关闭数据库
			b.Close()
		})
	})
	if err := mygo.App.Run(); err != nil {
		log.Fatal(err)
	}
}

// installMenu builds the native menu bar. Menu callbacks run on the main
// thread outside a frame, so every action goes through App.Post, which runs
// it on the UI thread and schedules a redraw.
func installMenu(a *app.App) {
	run := func(fn func()) func(*mygo.MenuItem, *mygo.Window) {
		return func(*mygo.MenuItem, *mygo.Window) { a.Post(fn) }
	}
	menu := mygo.NewMenu([]*mygo.MenuItem{
		{Role: mygo.RoleAppMenu},
		{Label: i18n.T("menu.file"), Submenu: []*mygo.MenuItem{
			{Label: i18n.T("menu.newAsset"), Accelerator: "CmdOrCtrl+N", Click: run(a.OpenNewHostForm)},
			{Label: i18n.T("menu.newFolder"), Accelerator: "CmdOrCtrl+Shift+N", Click: run(a.OpenNewFolderForm)},
			mygo.Separator(),
			{Label: i18n.T("menu.settings"), Accelerator: "CmdOrCtrl+,", Click: run(func() {
				a.OpenSettings(a.Settings.Tab)
			})},
			mygo.Separator(),
			{Role: mygo.RoleQuit},
		}},
		{Role: mygo.RoleEditMenu},
		{Label: i18n.T("menu.view"), Submenu: []*mygo.MenuItem{
			{Label: i18n.T("menu.assets"), Accelerator: "CmdOrCtrl+1", Click: run(a.CloseSettings)},
			mygo.Separator(),
			{Label: i18n.T("menu.settings.general"), Click: run(func() { a.OpenSettings(app.SettingsGeneral) })},
			{Label: i18n.T("menu.settings.model"), Click: run(func() { a.OpenSettings(app.SettingsModel) })},
			{Label: i18n.T("menu.settings.skills"), Click: run(func() { a.OpenSettings(app.SettingsSkills) })},
			{Label: i18n.T("menu.settings.snippets"), Click: run(func() { a.OpenSettings(app.SettingsSnippets) })},
			{Label: i18n.T("menu.settings.policy"), Click: run(func() { a.OpenSettings(app.SettingsPolicy) })},
			{Label: i18n.T("menu.settings.audit"), Click: run(func() { a.OpenSettings(app.SettingsAudit) })},
			{Label: i18n.T("menu.settings.about"), Click: run(func() { a.OpenSettings(app.SettingsAbout) })},
			mygo.Separator(),
			{Role: mygo.RoleToggleFullScreen},
		}},
		{Role: mygo.RoleWindowMenu},
	})
	mygo.App.SetMenu(menu)
}

// trayMenu builds the tray menu. It is rebuilt when the UI language changes.
func trayMenu(a *app.App, win *mygo.Window) *mygo.Menu {
	return mygo.NewMenu([]*mygo.MenuItem{
		{Label: i18n.T("tray.open"), Click: func(*mygo.MenuItem, *mygo.Window) {
			win.Show()
			win.Focus()
		}},
		{Label: i18n.T("tray.settings"), Click: func(*mygo.MenuItem, *mygo.Window) {
			a.Post(func() { a.OpenSettings(a.Settings.Tab) })
			win.Show()
			win.Focus()
		}},
		mygo.Separator(),
		{Role: mygo.RoleQuit},
	})
}

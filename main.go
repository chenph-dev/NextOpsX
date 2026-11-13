package main

import (
	"log"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"

	"nextopsx/internal/app"
	"nextopsx/internal/backend"
	"nextopsx/internal/views"
)

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
			Content:   ui.View(views.Shell(a)),
		})
		a.Attach(win)
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
		{Label: "文件", Submenu: []*mygo.MenuItem{
			{Label: "新建资产", Accelerator: "CmdOrCtrl+N", Click: run(a.OpenNewHostForm)},
			{Label: "新建目录", Accelerator: "CmdOrCtrl+Shift+N", Click: run(a.OpenNewFolderForm)},
			mygo.Separator(),
			{Label: "设置…", Accelerator: "CmdOrCtrl+,", Click: run(func() {
				a.OpenSettings(a.Settings.Tab)
			})},
			mygo.Separator(),
			{Role: mygo.RoleQuit},
		}},
		{Role: mygo.RoleEditMenu},
		{Label: "视图", Submenu: []*mygo.MenuItem{
			{Label: "资产", Accelerator: "CmdOrCtrl+1", Click: run(a.CloseSettings)},
			mygo.Separator(),
			{Label: "设置 · 常规", Click: run(func() { a.OpenSettings(app.SettingsGeneral) })},
			{Label: "设置 · 模型", Click: run(func() { a.OpenSettings(app.SettingsModel) })},
			{Label: "设置 · 技能", Click: run(func() { a.OpenSettings(app.SettingsSkills) })},
			{Label: "设置 · 审批", Click: run(func() { a.OpenSettings(app.SettingsPolicy) })},
			{Label: "设置 · 审计", Click: run(func() { a.OpenSettings(app.SettingsAudit) })},
			mygo.Separator(),
			{Role: mygo.RoleToggleFullScreen},
		}},
		{Role: mygo.RoleWindowMenu},
	})
	mygo.App.SetMenu(menu)
}

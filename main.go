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
			b.Close()
		})
	})
	if err := mygo.App.Run(); err != nil {
		log.Fatal(err)
	}
}

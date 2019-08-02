package app

import (
	"github.com/gotk3/gotk3/gtk"

	"github.com/ftl/panacotta/ui"
)

type mainWindow struct {
	window *gtk.ApplicationWindow
}

func newMainWindow(builder *gtk.Builder, application *gtk.Application) *mainWindow {
	result := new(mainWindow)

	result.window = ui.Get(builder, "mainWindow").(*gtk.ApplicationWindow)
	result.window.SetApplication(application)
	result.window.SetDefaultSize(2500, 500)

	return result
}

func (w *mainWindow) Show() {
	w.window.ShowAll()
}

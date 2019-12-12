package app

import (
	"github.com/gotk3/gotk3/gtk"

	"github.com/ftl/panacotta/ui"
)

type mainWindow struct {
	Window *gtk.ApplicationWindow
}

func newMainWindow(builder *gtk.Builder, application *gtk.Application) *mainWindow {
	result := new(mainWindow)

	result.Window = ui.Get(builder, "mainWindow").(*gtk.ApplicationWindow)
	result.Window.SetApplication(application)
	result.Window.SetDefaultSize(2500, 300)

	return result
}

func (w *mainWindow) Show() {
	w.Window.ShowAll()
}

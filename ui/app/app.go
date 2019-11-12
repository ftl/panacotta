package app

import (
	"log"

	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"

	"github.com/ftl/panacotta/ui/panorama"
)

// Run the application
func Run(controller Controller, args []string) {
	var err error
	a := &application{
		id:         "ft.panacotta",
		controller: controller,
	}

	a.app, err = gtk.ApplicationNew(a.id, glib.APPLICATION_FLAGS_NONE)
	if err != nil {
		log.Fatal("Cannot create application: ", err)
	}
	a.app.Connect("startup", a.startup)
	a.app.Connect("activate", a.activate)
	a.app.Connect("shutdown", a.shutdown)

	a.app.Run(args)
}

// Controller of the application
type Controller interface {
	panorama.Controller

	Startup()
	Shutdown()
}

type application struct {
	id         string
	controller Controller
	app        *gtk.Application
	builder    *gtk.Builder
	mainWindow *mainWindow
	done       chan struct{}
}

func (a *application) startup() {
	a.done = make(chan struct{})
}

func (a *application) activate() {
	a.builder = setupBuilder()
	a.mainWindow = newMainWindow(a.builder, a.app)

	a.controller.Startup()
	panorama.New(a.builder, a.controller)

	a.mainWindow.Show()
}

func (a *application) shutdown() {
	close(a.done)
	a.controller.Shutdown()
}

func setupBuilder() *gtk.Builder {
	builder, err := gtk.BuilderNew()
	if err != nil {
		log.Fatal("Cannot create builder: ", err)
	}

	err = builder.AddFromFile("ui/glade/ui.glade")
	// builder.AddFromString(glade.MustAssetString("contest.glade"))
	if err != nil {
		log.Fatal("Cannot load glade resource: ", err)
	}

	return builder
}

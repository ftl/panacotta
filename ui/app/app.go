package app

import (
	"log"

	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"

	coreapp "github.com/ftl/panacotta/core/app"
	"github.com/ftl/panacotta/ui/panorama"
)

// Run the application
func Run(args []string) {
	var err error
	a := &application{id: "ft.panacotta"}
	a.app, err = gtk.ApplicationNew(a.id, glib.APPLICATION_FLAGS_NONE)
	if err != nil {
		log.Fatal("Cannot create application: ", err)
	}

	a.app.Connect("startup", a.startup)
	a.app.Connect("activate", a.activate)
	a.app.Connect("shutdown", a.shutdown)

	a.app.Run(args)
}

type controller interface {
	panorama.Controller

	Startup()
	Shutdown()
	SetPanoramaView(coreapp.PanoramaView)
}

type application struct {
	id         string
	app        *gtk.Application
	builder    *gtk.Builder
	mainWindow *mainWindow
	controller controller
	done       chan struct{}
}

func (a *application) startup() {
	a.done = make(chan struct{})
}

func (a *application) activate() {
	a.builder = setupBuilder()

	a.controller = coreapp.NewController()
	a.mainWindow = newMainWindow(a.builder, a.app)
	a.controller.SetPanoramaView(panorama.New(a.builder, a.controller))

	a.controller.Startup()

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

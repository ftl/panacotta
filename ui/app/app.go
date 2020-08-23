package app

import (
	"log"
	"path/filepath"

	"github.com/ftl/gmtry"
	"github.com/ftl/hamradio/cfg"
	"github.com/gotk3/gotk3/gdk"
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

	gdk.SetAllowedBackends("x11")

	a.app, err = gtk.ApplicationNew(a.id, glib.APPLICATION_FLAGS_NONE)
	if err != nil {
		log.Fatal("Cannot create application: ", err)
	}
	a.app.Connect("startup", a.startup)
	a.app.Connect("activate", a.activate)
	a.app.Connect("shutdown", a.shutdown)

	a.app.Run(args)
	log.Print("app finished")
}

// Controller of the application
type Controller interface {
	panorama.Controller

	Startup()
	Shutdown()
}

type application struct {
	id             string
	controller     Controller
	app            *gtk.Application
	builder        *gtk.Builder
	mainWindow     *mainWindow
	windowGeometry *gmtry.Geometry
	done           chan struct{}
}

func (a *application) startup() {
	a.done = make(chan struct{})

	configDir, err := cfg.Directory("")
	if err != nil {
		log.Fatalf("No access to configuration directory %s: %v", cfg.DefaultDirectory, err)
	}
	filename := filepath.Join(configDir, "panacotta.geometry")
	a.windowGeometry = gmtry.NewGeometry(filename)
}

func (a *application) activate() {
	a.builder = setupBuilder()
	a.mainWindow = newMainWindow(a.builder, a.app)

	a.controller.Startup()
	panorama.New(a.builder, a.controller)

	connectToGeometry(a.windowGeometry, "main", &a.mainWindow.Window.Window)
	err := a.windowGeometry.Restore()
	if err != nil {
		log.Printf("Cannot load window geometry, using defaults instead: %v", err)
	}

	a.mainWindow.Show()
}

func (a *application) shutdown() {
	close(a.done)
	a.controller.Shutdown()

	err := a.windowGeometry.Store()
	if err != nil {
		log.Printf("Cannot store window geometry: %v", err)
	}
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

func connectToGeometry(geometry *gmtry.Geometry, id gmtry.ID, window *gtk.Window) {
	geometry.Add(id, window)

	window.Connect("configure-event", func(_ interface{}, event *gdk.Event) {
		e := gdk.EventConfigureNewFromEvent(event)
		w := geometry.Get(id)
		w.SetPosition(window.GetPosition())
		w.SetSize(e.Width(), e.Height())
	})
	window.Connect("window-state-event", func(_ interface{}, event *gdk.Event) {
		e := gdk.EventWindowStateNewFromEvent(event)
		if e.ChangedMask()&gdk.WINDOW_STATE_MAXIMIZED == gdk.WINDOW_STATE_MAXIMIZED {
			geometry.Get(id).SetMaximized(e.NewWindowState()&gdk.WINDOW_STATE_MAXIMIZED == gdk.WINDOW_STATE_MAXIMIZED)
		}
	})
}

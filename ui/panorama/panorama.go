package panorama

import (
	"log"

	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"

	"github.com/ftl/panacotta/core"
	"github.com/ftl/panacotta/ui"
)

// New returns a new instance of the FFT View, connected to the fftArea accesible through the given builder.
func New(builder *gtk.Builder, controller Controller) *View {
	result := View{
		view:       ui.Get(builder, "panoramaView").(*gtk.DrawingArea),
		controller: controller,
	}
	result.view.Connect("draw", result.onDraw)
	result.view.Connect("configure-event", result.onResize)
	result.connectMouse()
	result.connectKeyboard()

	go result.run()

	return &result
}

// Controller for the panorama view.
type Controller interface {
	Done() chan struct{}
	Panorama() <-chan core.Panorama
	SetPanoramaSize(core.Px, core.Px)
	TuneTo(core.Frequency)
	TuneBy(core.Frequency)
	TuneUp()
	TuneDown()
	ToggleViewMode()
	ZoomIn()
	ZoomOut()
	ResetZoom()
}

// View of the FFT.
type View struct {
	view       *gtk.DrawingArea
	controller Controller

	data core.Panorama

	mouse    mouse
	keyboard keyboard
}

func (v *View) run() {
	for {
		select {
		case data := <-v.controller.Panorama():
			glib.IdleAdd(func() bool {
				v.data = data
				v.view.QueueDraw()
				return true
			})
		case <-v.controller.Done():
			return
		}
	}
	log.Print("View.run done")
}

func (v *View) onResize(widget *gtk.DrawingArea, event *gdk.Event) {
	e := gdk.EventConfigureNewFromEvent(event)
	v.controller.SetPanoramaSize(core.Px(e.Width()), core.Px(e.Height()))
}

func (v *View) deviceToFrequency(x float64) core.Frequency {
	return v.data.ToHz(core.Px(x))
}

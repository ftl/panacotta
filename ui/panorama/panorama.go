package panorama

import (
	"log"
	"sync"

	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/gtk"

	"github.com/ftl/panacotta/core"
	"github.com/ftl/panacotta/ui"
)

// New returns a new instance of the FFT View, connected to the fftArea accesible through the given builder.
func New(builder *gtk.Builder, controller Controller) *View {
	result := View{
		view:       ui.Get(builder, "panoramaView").(*gtk.DrawingArea),
		controller: controller,
		dataLock:   new(sync.RWMutex),
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

	data     core.Panorama
	dataLock *sync.RWMutex

	mouse    mouse
	keyboard keyboard
}

func (v *View) run() {
	for {
		select {
		case data := <-v.controller.Panorama():
			v.updateData(data)
			v.view.QueueDraw()
		case <-v.controller.Done():
			return
		}
	}
	log.Print("View.run done")
}

func (v *View) updateData(data core.Panorama) {
	v.dataLock.Lock()
	defer v.dataLock.Unlock()
	v.data = data
}

func (v *View) onResize(widget *gtk.DrawingArea, event *gdk.Event) {
	e := gdk.EventConfigureNewFromEvent(event)
	v.controller.SetPanoramaSize(core.Px(e.Width()), core.Px(e.Height()))
}

func (v *View) deviceToFrequency(x float64) core.Frequency {
	data := v.currentData()
	return data.ToHz(core.Px(x))
}

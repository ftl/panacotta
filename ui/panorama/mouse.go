package panorama

import (
	"log"
	"math"

	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/gtk"
)

type mouse struct {
	buttonPressed  bool
	doublePressed  bool
	startX, startY float64
	button         uint
	dragThreshold  float64
	dragging       bool
}

func (v *View) connectMouse() {
	log.Print("connect mouse")
	v.mouse = mouse{
		dragThreshold: 10.0,
	}

	v.view.AddEvents(
		int(gdk.BUTTON_PRESS_MASK) |
			int(gdk.BUTTON_RELEASE_MASK) |
			int(gdk.POINTER_MOTION_MASK) |
			int(gdk.SCROLL_MASK))

	v.view.Connect("button-press-event", v.onButtonPress)
	v.view.Connect("button-release-event", v.onButtonRelease)
	v.view.Connect("motion-notify-event", v.onPointerMotion)
	v.view.Connect("scroll-event", v.onScroll)
}

func (v *View) onButtonPress(da *gtk.DrawingArea, e *gdk.Event) {
	buttonEvent := gdk.EventButtonNewFromEvent(e)
	if v.mouse.buttonPressed {
		v.mouse.doublePressed = true
		return
	}

	v.mouse.buttonPressed = true
	v.mouse.startX, v.mouse.startY = buttonEvent.X(), buttonEvent.Y()
	v.mouse.button = buttonEvent.Button()
}

func (v *View) onButtonRelease(da *gtk.DrawingArea, e *gdk.Event) {
	if v.mouse.doublePressed {
		v.onDoubleClick(v.mouse.button)
	} else if v.mouse.dragging {
		log.Printf("drag end")
	} else if v.mouse.buttonPressed {
		v.onClick(v.mouse.button)
	}

	v.mouse.buttonPressed = false
	v.mouse.doublePressed = false
	v.mouse.startX, v.mouse.startY = 0, 0
	v.mouse.button = 0
	v.mouse.dragging = false
}

func (v *View) onClick(button uint) {
	switch button {
	case 1:
		v.controller.Tune(v.deviceToFrequency(v.mouse.startX))
	case 2:
		v.controller.ToggleViewMode()
	default:
		log.Printf("click %d", button)
	}
}

func (v *View) onDoubleClick(button uint) {
	log.Printf("double click %d", button)
}

func (v *View) onPointerMotion(da *gtk.DrawingArea, e *gdk.Event) {
	var x, y float64
	if v.mouse.buttonPressed {
		motionEvent := gdk.EventMotionNewFromEvent(e)
		x, y = motionEvent.MotionVal()

		if math.Abs(v.mouse.startX-x) > v.mouse.dragThreshold {
			v.mouse.dragging = true
		}
	}

	if v.mouse.dragging {
		log.Printf("dragging x %f y %f", x, y)
	}
}

func (v *View) onScroll(da *gtk.DrawingArea, e *gdk.Event) {
	scrollEvent := gdk.EventScrollNewFromEvent(e)
	switch scrollEvent.Direction() {
	case gdk.SCROLL_UP:
		v.controller.TuneUp()
	case gdk.SCROLL_DOWN:
		v.controller.TuneDown()
	default:
		log.Printf("unknown scroll direction %d", scrollEvent.Direction())
	}
}

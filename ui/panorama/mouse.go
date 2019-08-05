package panorama

import (
	"log"
	"math"

	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/gtk"
)

type mouse struct {
	buttonPressed  bool
	startX, startY float64
	button         uint
	dragThreshold  float64
	dragging       bool
}

func (v *View) connectMouse() {
	v.mouse = mouse{
		dragThreshold: 10.0,
	}

	v.view.AddEvents(int(gdk.BUTTON_PRESS_MASK))
	v.view.AddEvents(int(gdk.BUTTON_RELEASE_MASK))
	v.view.AddEvents(int(gdk.POINTER_MOTION_MASK))
	v.view.AddEvents(int(gdk.SCROLL_MASK))
	v.view.Connect("button-press-event", v.onButtonPress)
	v.view.Connect("button-release-event", v.onButtonRelease)
	v.view.Connect("motion-notify-event", v.onPointerMotion)
	v.view.Connect("scroll-event", v.onScroll)
}

func (v *View) onButtonPress(da *gtk.DrawingArea, e *gdk.Event) {
	buttonEvent := gdk.EventButtonNewFromEvent(e)
	if v.mouse.buttonPressed {
		log.Printf("double clock x %f y %f button %d", v.mouse.startX, v.mouse.startY, v.mouse.button)
		switch v.mouse.button {
		case 1:
			v.controller.ToggleViewMode()
		}
		return
	}

	v.mouse.buttonPressed = true
	v.mouse.startX, v.mouse.startY = buttonEvent.X(), buttonEvent.Y()
	v.mouse.button = buttonEvent.Button()

	log.Printf("button press x %f y %f button %d", v.mouse.startX, v.mouse.startY, v.mouse.button)

	switch v.mouse.button {
	case 1:
		v.controller.Tune(v.deviceToFrequency(v.mouse.startX))
	}
}

func (v *View) onButtonRelease(da *gtk.DrawingArea, e *gdk.Event) {
	log.Print("button release")
	v.mouse.buttonPressed = false
	v.mouse.startX, v.mouse.startY = 0, 0
	v.mouse.button = 0
	v.mouse.dragging = false
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
		log.Print("scroll up")
		v.controller.FineTuneUp()
	case gdk.SCROLL_DOWN:
		log.Print("scroll down")
		v.controller.FineTuneDown()
	case gdk.SCROLL_LEFT:
		log.Print("scroll left")
	case gdk.SCROLL_RIGHT:
		log.Print("scroll right")
	default:
		log.Print("direction unknown")
	}
}

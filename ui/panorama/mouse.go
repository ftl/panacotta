package panorama

import (
	"log"
	"math/cmplx"

	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/gtk"

	"github.com/ftl/panacotta/core"
)

type mouse struct {
	buttonPressed  bool
	doublePressed  bool
	x, y           float64
	lastX, lastY   float64
	startX, startY float64
	button         uint
	dragThreshold  float64
}

func (m mouse) moveX() float64 {
	return m.lastX - m.x
}

func (m mouse) moveY() float64 {
	return m.lastY - m.y
}

func (m mouse) moveDistance() float64 {
	return cmplx.Abs(complex(m.moveX(), m.moveY()))
}

func (m mouse) dragX() float64 {
	return m.startX - m.x
}

func (m mouse) dragY() float64 {
	return m.startY - m.y
}

func (m mouse) dragDistance() float64 {
	return cmplx.Abs(complex(m.dragX(), m.dragY()))
}

func (m mouse) dragging() bool {
	return m.buttonPressed && m.dragDistance() > m.dragThreshold
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
	} else if v.mouse.dragging() {
		log.Printf("drag end")
	} else if v.mouse.buttonPressed {
		v.onClick(v.mouse.button)
	}

	v.mouse.buttonPressed = false
	v.mouse.doublePressed = false
	v.mouse.startX, v.mouse.startY = 0, 0
	v.mouse.button = 0
}

func (v *View) onClick(button uint) {
	switch button {
	case 1:
		v.onSingleLeftClick(v.mouse.startX, v.mouse.startY)
	case 2:
		v.controller.ToggleViewMode()
	default:
		log.Printf("click %d", button)
	}
}

func (v *View) onSingleLeftClick(x, y float64) {
	pointer := point{x, y}
	for i, r := range v.geometry.peaks {
		if r.contains(pointer) {
			f := v.data.Peaks[i].MaxFrequency
			v.controller.TuneTo(f)
			return
		}
	}
	if v.geometry.fft.contains(pointer) || v.geometry.waterfall.contains(pointer) {
		v.controller.TuneTo(v.deviceToFrequency(x))
	}
}

func (v *View) onDoubleClick(button uint) {
	switch button {
	case 1:
		v.onDoubleLeftClick(v.mouse.startX, v.mouse.startY)
	default:
		log.Printf("double click %d", button)
	}
}

func (v *View) onDoubleLeftClick(x, y float64) {
	pointer := point{x, y}
	if v.geometry.bandIndicator.contains(pointer) {
		v.controller.ZoomToBand()
	}
}

func (v *View) onPointerMotion(da *gtk.DrawingArea, e *gdk.Event) {
	motionEvent := gdk.EventMotionNewFromEvent(e)
	v.mouse.lastX, v.mouse.lastY = v.mouse.x, v.mouse.y
	v.mouse.x, v.mouse.y = motionEvent.MotionVal()

	if v.mouse.dragging() {
		v.onDrag()
	}
}

func (v *View) onDrag() {
	pointer := point{v.mouse.startX, v.mouse.startY}
	if v.geometry.dbScale.contains(pointer) {
		shift := -v.mouse.moveY() / v.geometry.dbScale.height()
		v.controller.ShiftDynamicRange(core.Frct(shift))
	} else {
		log.Printf("dragging x %f y %f", v.mouse.x, v.mouse.y)
	}
}

func (v *View) onScroll(da *gtk.DrawingArea, e *gdk.Event) {
	scrollEvent := gdk.EventScrollNewFromEvent(e)

	pointer := point{scrollEvent.X(), scrollEvent.Y()}
	if v.geometry.dbScale.contains(pointer) {
		switch scrollEvent.Direction() {
		case gdk.SCROLL_UP:
			v.controller.FinerDynamicRange()
		case gdk.SCROLL_DOWN:
			v.controller.CoarserDynamicRange()
		default:
			log.Printf("unknown scroll direction %d", scrollEvent.Direction())
		}
	} else {
		switch scrollEvent.Direction() {
		case gdk.SCROLL_UP:
			v.controller.TuneUp()
		case gdk.SCROLL_DOWN:
			v.controller.TuneDown()
		default:
			log.Printf("unknown scroll direction %d", scrollEvent.Direction())
		}
	}
}

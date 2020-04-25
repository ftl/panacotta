package panorama

import (
	"log"

	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/gtk"
)

type keyboard map[uint]func()

func (v *View) connectKeyboard() {
	log.Print("connect keyboard")
	v.keyboard = keyboard{
		gdk.KEY_Up:    v.controller.ZoomOut,
		gdk.KEY_Down:  v.controller.ZoomIn,
		gdk.KEY_Left:  v.controller.TuneDown,
		gdk.KEY_Right: v.controller.TuneUp,
		gdk.KEY_d:     v.controller.ToggleSignalDetection,
		gdk.KEY_r:     v.controller.ResetZoom,
		gdk.KEY_v:     v.controller.ToggleViewMode,
	}

	v.view.SetCanFocus(true)
	v.view.AddEvents(int(gdk.KEY_PRESS_MASK) | int(gdk.KEY_RELEASE_MASK))

	v.view.Connect("key-press-event", v.onKeyPress)
	v.view.Connect("key-release-event", v.onKeyRelease)
}

func (v *View) onKeyPress(da *gtk.DrawingArea, event *gdk.Event) bool {
	keyEvent := gdk.EventKeyNewFromEvent(event)
	if action, ok := v.keyboard[keyEvent.KeyVal()]; ok {
		action()
		return true
	}
	return false
}

func (v *View) onKeyRelease(da *gtk.DrawingArea, event *gdk.Event) bool {
	return false
}

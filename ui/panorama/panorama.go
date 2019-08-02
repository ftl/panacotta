package panorama

import (
	"sync"
	"time"

	"github.com/gotk3/gotk3/cairo"
	"github.com/gotk3/gotk3/gtk"

	"github.com/ftl/panacotta/ui"
)

// New returns a new instance of the FFT View, connected to the fftArea accesible through the given builder.
func New(builder *gtk.Builder) *View {
	result := new(View)
	result.view = ui.Get(builder, "panoramaView").(*gtk.DrawingArea)
	result.view.Connect("draw", result.onDraw)

	result.dataLock = new(sync.RWMutex)
	result.redrawInterval = (1 * time.Second) / time.Duration(5)

	return result
}

// View of the FFT.
type View struct {
	view *gtk.DrawingArea

	data           []float64
	dataLock       *sync.RWMutex
	lastRedraw     time.Time
	redrawInterval time.Duration
}

// ShowData shows the given data
func (v *View) ShowData(data []float64) {
	now := time.Now()
	if now.Sub(v.lastRedraw) < v.redrawInterval {
		return
	}

	v.dataLock.Lock()
	defer v.dataLock.Unlock()
	v.data = data

	v.lastRedraw = now
	v.view.QueueDraw()
}

func (v *View) onDraw(da *gtk.DrawingArea, cr *cairo.Context) {
	var data []float64
	func() {
		v.dataLock.RLock()
		defer v.dataLock.RUnlock()
		data = v.data
	}()

	if len(data) == 0 {
		return
	}

	height, width := float64(da.GetAllocatedHeight()), float64(da.GetAllocatedWidth())

	fullWidth := 10000
	scaleX := float64(fullWidth) / float64(len(v.data))
	maxY := 100.0

	scaleX = width / float64(len(data))

	cr.Save()

	m := new(cairo.Matrix)
	m.InitTranslate(0, height)
	m.Scale(scaleX, -height/maxY)
	cr.Transform(m)

	centerX, _ := cr.UserToDeviceDistance(float64(len(data)/2), 0)

	cr.SetSourceRGBA(1.0, 0, 0, 0.3)
	cr.MoveTo(0, 0)
	drawLine(cr, data)
	cr.ClosePath()
	cr.Fill()

	cr.MoveTo(0, 0)
	drawLine(cr, data)

	cr.Restore()

	cr.SetLineWidth(0.5)
	cr.SetSourceRGB(1.0, 0, 0)
	cr.Stroke()

	cr.SetSourceRGB(0, 0, 1.0)
	cr.MoveTo(centerX, 0)
	cr.LineTo(centerX, height)
	cr.Stroke()
}

func drawLine(cr *cairo.Context, line []float64) {
	for i := 0; i < len(line); i++ {
		cr.LineTo(float64(i), line[i])
	}
	cr.LineTo(float64((len(line) - 1)), 0)
}

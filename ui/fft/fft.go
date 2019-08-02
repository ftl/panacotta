package fft

import (
	"math"
	"sync"
	"time"

	"github.com/gotk3/gotk3/cairo"
	"github.com/gotk3/gotk3/gtk"

	"github.com/ftl/panacotta/ui"
)

// New returns a new instance of the FFT View, connected to the fftArea accesible through the given builder.
func New(builder *gtk.Builder) *View {
	result := new(View)
	result.view = ui.Get(builder, "fftArea").(*gtk.DrawingArea)
	result.view.Connect("draw", result.onDraw)

	result.dataLock = new(sync.RWMutex)
	result.redrawInterval = (1 * time.Second) / time.Duration(5)
	result.smoothingBuffer = make([][]complex128, 5)

	return result
}

// View of the FFT.
type View struct {
	view *gtk.DrawingArea

	data           []complex128
	dataLock       *sync.RWMutex
	lastRedraw     time.Time
	redrawInterval time.Duration

	smoothingBuffer [][]complex128
	smoothingIndex  int
}

// ShowData shows the given data
func (v *View) ShowData(data []complex128) {
	v.smoothingBuffer[v.smoothingIndex] = data
	v.smoothingIndex = (v.smoothingIndex + 1) % len(v.smoothingBuffer)

	now := time.Now()
	if now.Sub(v.lastRedraw) < v.redrawInterval {
		return
	}

	average := make([]complex128, len(data))
	for i := 0; i < len(average); i++ {
		var re, im float64
		for j := 0; j < len(v.smoothingBuffer); j++ {
			if len(v.smoothingBuffer[j]) != len(data) {
				continue
			}
			re = math.Max(real(v.smoothingBuffer[j][i]), re)
			im = math.Max(imag(v.smoothingBuffer[j][i]), im)
		}
		average[i] = complex(re, im)
	}

	v.dataLock.Lock()
	defer v.dataLock.Unlock()
	v.data = average

	v.lastRedraw = now
	v.view.QueueDraw()
}

func (v *View) onDraw(da *gtk.DrawingArea, cr *cairo.Context) {
	var data []complex128
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

	line := make([]float64, 2000)
	offset := (len(data) - len(line)) / 2
	for i := 0; i < len(line); i++ {
		var d complex128
		if i < len(line)/2 {
			d = data[len(data)/2+(i+offset)]
		} else {
			d = data[(i+offset)-len(data)/2]
		}
		pwr := (imag(d)*imag(d) + real(d)*real(d))
		line[i] = 10.0*math.Log10(pwr+1.0e-20) + 0.5
	}
	scaleX = width / float64(len(line))

	cr.Save()

	m := new(cairo.Matrix)
	m.InitTranslate(0, height)
	m.Scale(scaleX, -height/maxY)
	cr.Transform(m)

	centerX, _ := cr.UserToDeviceDistance(float64(len(line)/2), 0)

	cr.SetSourceRGBA(1.0, 0, 0, 0.3)
	cr.MoveTo(0, 0)
	drawLine(cr, line)
	cr.ClosePath()
	cr.Fill()

	cr.MoveTo(0, 0)
	drawLine(cr, line)

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

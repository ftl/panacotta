package panorama

import (
	"sync"
	"time"

	"github.com/gotk3/gotk3/cairo"
	"github.com/gotk3/gotk3/gtk"

	"github.com/ftl/panacotta/core"
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

	fftData      []float64
	vfoFrequency core.Frequency
	vfoROI       core.FrequencyRange
	dataLock     *sync.RWMutex

	lastRedraw     time.Time
	redrawInterval time.Duration
}

// SetFFTData sets the current FFT data.
func (v *View) SetFFTData(data []float64) {
	v.dataLock.Lock()
	defer v.dataLock.Unlock()
	v.fftData = data

	v.TriggerRedraw()
}

// SetVFO sets the VFO configuration.
func (v *View) SetVFO(frequency core.Frequency, roi core.FrequencyRange) {
	v.dataLock.Lock()
	defer v.dataLock.Unlock()
	v.vfoFrequency = frequency
	v.vfoROI = roi

	v.TriggerRedraw()
}

// TriggerRedraw triggers a redraw according to the redraw interval.
func (v *View) TriggerRedraw() {
	now := time.Now()
	if now.Sub(v.lastRedraw) < v.redrawInterval {
		return
	}

	v.lastRedraw = now
	v.view.QueueDraw()
}

func (v *View) onDraw(da *gtk.DrawingArea, cr *cairo.Context) {
	data, vfoFrequency, vfoROI := func() ([]float64, core.Frequency, core.FrequencyRange) {
		v.dataLock.RLock()
		defer v.dataLock.RUnlock()
		return v.fftData, v.vfoFrequency, v.vfoROI
	}()

	blockSize := len(data)
	if blockSize == 0 {
		return
	}

	_ = vfoFrequency
	_ = vfoROI
	hzPerBin := float64(vfoROI.Width()) / float64(blockSize)

	height, width := float64(da.GetAllocatedHeight()), float64(da.GetAllocatedWidth())

	scaleX := float64(width) / float64(vfoROI.Width())
	maxY := 50.0

	cr.Save()

	m := new(cairo.Matrix)
	m.InitTranslate(0, height)
	m.Scale(scaleX, -height/maxY)
	cr.Transform(m)

	vfoX, _ := cr.UserToDeviceDistance(float64(vfoFrequency-vfoROI.From), 0)

	cr.SetSourceRGBA(1.0, 0, 0, 0.3)
	cr.MoveTo(0, 0)
	drawLine(cr, data, hzPerBin)
	cr.ClosePath()
	cr.Fill()

	cr.MoveTo(0, 0)
	drawLine(cr, data, hzPerBin)

	cr.Restore()

	cr.SetLineWidth(0.5)
	cr.SetSourceRGB(1.0, 0, 0)
	cr.Stroke()

	cr.SetLineWidth(1.5)
	cr.SetSourceRGB(0, 0, 1.0)
	cr.MoveTo(vfoX, 0)
	cr.LineTo(vfoX, height)
	cr.Stroke()
}

func drawLine(cr *cairo.Context, line []float64, hzPerBin float64) {
	for i := 0; i < len(line); i++ {
		cr.LineTo(float64(i)*hzPerBin, line[i])
	}
	cr.LineTo(float64((len(line)-1))*hzPerBin, 0)
}

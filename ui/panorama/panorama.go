package panorama

import (
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/gotk3/gotk3/cairo"
	"github.com/gotk3/gotk3/gtk"

	"github.com/ftl/panacotta/core"
	"github.com/ftl/panacotta/ui"
)

// New returns a new instance of the FFT View, connected to the fftArea accesible through the given builder.
func New(builder *gtk.Builder, controller Controller) *View {
	result := View{
		view:       ui.Get(builder, "panoramaView").(*gtk.DrawingArea),
		controller: controller,

		dataLock:       new(sync.RWMutex),
		redrawInterval: (1 * time.Second) / time.Duration(5),
	}
	result.view.Connect("draw", result.onDraw)
	result.connectMouse()

	return &result
}

// Controller for the panorama view.
type Controller interface {
	Tune(core.Frequency)
	FineTuneUp()
	FineTuneDown()
}

// View of the FFT.
type View struct {
	view       *gtk.DrawingArea
	controller Controller

	fftData      []float64
	vfoFrequency core.Frequency
	vfoROI       core.FrequencyRange
	dataLock     *sync.RWMutex

	lastRedraw     time.Time
	redrawInterval time.Duration

	mouse mouse
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

func (v *View) deviceToFrequency(x float64) core.Frequency {
	vfoROI := func() core.FrequencyRange {
		v.dataLock.RLock()
		defer v.dataLock.RUnlock()
		return v.vfoROI
	}()
	scaleX := float64(v.view.GetAllocatedWidth()) / float64(vfoROI.Width())

	return vfoROI.From + core.Frequency(x/scaleX)
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

	hzPerBin := float64(vfoROI.Width()) / float64(blockSize)

	height, width := float64(da.GetAllocatedHeight()), float64(da.GetAllocatedWidth())

	scaleX := float64(width) / float64(vfoROI.Width())
	maxY := 100.0

	m := new(cairo.Matrix)
	m.InitTranslate(0, height)
	m.Scale(scaleX, -height/maxY)

	fillBackground(cr, width, height)
	dbmY := calculateInfoCoordinates(cr, m, vfoFrequency, vfoROI, maxY)

	drawDBMScale(cr, dbmY, width)
	drawFreqScale(cr, m, vfoROI, height)
	drawVFO(cr, m, vfoFrequency, vfoROI, height)
	drawFFT(cr, m, data, hzPerBin)
}

func fillBackground(cr *cairo.Context, width, height float64) {
	cr.Save()
	defer cr.Restore()

	cr.SetSourceRGB(0, 0, 0)
	cr.Paint()
}

func calculateInfoCoordinates(cr *cairo.Context, matrix *cairo.Matrix, vfoFrequency core.Frequency, vfoROI core.FrequencyRange, maxY float64) (dbmY []float64) {
	cr.Save()
	defer cr.Restore()
	cr.Transform(matrix)

	dbmY = make([]float64, int(maxY/10)-1)
	for i := range dbmY {
		_, dbmY[i] = cr.UserToDeviceDistance(0, maxY-float64((i+1)*10))
		dbmY[i] *= -1
	}

	return
}

func drawDBMScale(cr *cairo.Context, ys []float64, width float64) {
	cr.Save()
	defer cr.Restore()

	cr.SetSourceRGB(0.8, 0.8, 0.8)
	cr.SetLineWidth(0.5)
	cr.SetDash([]float64{10, 10}, 0)
	for _, y := range ys {
		cr.MoveTo(0, y)
		cr.LineTo(width, y)
		cr.Stroke()
	}
	cr.SetDash([]float64{}, 0)
}

func drawFreqScale(cr *cairo.Context, matrix *cairo.Matrix, vfoROI core.FrequencyRange, height float64) {
	cr.Save()
	defer cr.Restore()

	fMagnitude := int(math.Pow(10, float64(int(math.Log10(float64(vfoROI.Width())))-1)))
	fFactor := fMagnitude
	for int(vfoROI.Width())/fFactor > 15 {
		fFactor *= 2
	}

	type freqUnit struct {
		x float64
		f int
	}

	cr.Save()
	cr.Transform(matrix)
	freqScale := make([]freqUnit, 0, int(vfoROI.Width())/fFactor)
	for f := (int(vfoROI.From) / fFactor) * fFactor; f < int(vfoROI.To); f += fFactor {
		x, _ := cr.UserToDeviceDistance(float64(f)-float64(vfoROI.From), 0)
		unit := freqUnit{
			x: x,
			f: f,
		}
		freqScale = append(freqScale, unit)
	}
	cr.Restore()

	cr.SetSourceRGB(0.8, 0.8, 0.8)
	cr.SetLineWidth(0.5)
	cr.SetDash([]float64{10, 10}, 0)
	for _, u := range freqScale {
		cr.MoveTo(u.x, 0)
		cr.LineTo(u.x, height)
		cr.Stroke()

		freqText := fmt.Sprintf("%d", u.f/fMagnitude)
		extents := cr.TextExtents(freqText)
		cr.SetFontSize(20.0)
		cr.MoveTo(u.x+2, extents.Height+2)
		cr.ShowText(freqText)
	}
	cr.SetDash([]float64{}, 0)
}

func drawFFT(cr *cairo.Context, matrix *cairo.Matrix, data []float64, hzPerBin float64) {
	cr.Save()
	cr.Transform(matrix)

	cr.SetSourceRGBA(1, 1, 1, 0.3)
	cr.MoveTo(0, 0)
	drawFFTLine(cr, data, hzPerBin)
	cr.ClosePath()
	cr.Fill()

	cr.MoveTo(0, 0)
	drawFFTLine(cr, data, hzPerBin)

	cr.Restore()

	cr.SetLineWidth(0.5)
	cr.SetSourceRGB(1, 1, 1)
	cr.Stroke()
}

func drawFFTLine(cr *cairo.Context, line []float64, hzPerBin float64) {
	for i := 0; i < len(line); i++ {
		cr.LineTo(float64(i)*hzPerBin, line[i])
	}
	cr.LineTo(float64((len(line)-1))*hzPerBin, 0)
}

func drawVFO(cr *cairo.Context, matrix *cairo.Matrix, vfoFrequency core.Frequency, vfoROI core.FrequencyRange, height float64) {
	cr.Save()
	defer cr.Restore()

	cr.Save()
	cr.Transform(matrix)
	vfoX, _ := cr.UserToDeviceDistance(float64(vfoFrequency-vfoROI.From), 0)
	cr.Restore()

	cr.SetLineWidth(1.5)
	cr.SetSourceRGB(0.6, 0.9, 1.0)
	cr.MoveTo(vfoX, 0)
	cr.LineTo(vfoX, height)
	cr.Stroke()

	cr.SetFontSize(20.0)
	freqText := fmt.Sprintf("%.2fkHz", vfoFrequency/1000)
	vfoExtents := cr.TextExtents("VFO")
	freqExtents := cr.TextExtents(freqText)
	padding := 4.0

	cr.SetLineWidth(1.5)
	cr.SetSourceRGBA(0, 0, 0, 0.5)
	cr.MoveTo(vfoX, 0)
	cr.LineTo(vfoX+math.Max(vfoExtents.Width, freqExtents.Width)+2*padding, 0)
	cr.LineTo(vfoX+math.Max(vfoExtents.Width, freqExtents.Width)+2*padding, vfoExtents.Height+freqExtents.Height+2*padding)
	cr.LineTo(vfoX, vfoExtents.Height+freqExtents.Height+2*padding)
	cr.ClosePath()
	cr.Fill()

	cr.SetLineWidth(1.5)
	cr.SetSourceRGB(0.6, 0.9, 1.0)
	cr.MoveTo(vfoX, 0)
	cr.LineTo(vfoX+math.Max(vfoExtents.Width, freqExtents.Width)+2*padding, 0)
	cr.LineTo(vfoX+math.Max(vfoExtents.Width, freqExtents.Width)+2*padding, vfoExtents.Height+freqExtents.Height+2*padding)
	cr.LineTo(vfoX, vfoExtents.Height+freqExtents.Height+2*padding)
	cr.ClosePath()
	cr.Stroke()

	cr.MoveTo(vfoX+padding, vfoExtents.Height+padding)
	cr.ShowText("VFO")
	cr.MoveTo(vfoX+padding, vfoExtents.Height+freqExtents.Height+padding)
	cr.ShowText(freqText)
}

package panorama

import (
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/gotk3/gotk3/cairo"
	"github.com/gotk3/gotk3/gtk"

	"github.com/ftl/panacotta/core"
	"github.com/ftl/panacotta/core/bandplan"
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
	ToggleViewMode()
}

// View of the FFT.
type View struct {
	view       *gtk.DrawingArea
	controller Controller

	fftData  []float64
	vfo      vfo
	dataLock *sync.RWMutex

	lastRedraw     time.Time
	redrawInterval time.Duration

	mouse mouse
}

type vfo struct {
	frequency core.Frequency
	band      bandplan.Band
	roi       core.FrequencyRange
	mode      string
	bandwidth core.Frequency
}

// SetFFTData sets the current FFT data.
func (v *View) SetFFTData(data []float64) {
	v.dataLock.Lock()
	defer v.dataLock.Unlock()
	v.fftData = data

	v.TriggerRedraw()
}

// SetVFO sets the VFO configuration.
func (v *View) SetVFO(frequency core.Frequency, band bandplan.Band, roi core.FrequencyRange, mode string, bandwidth core.Frequency) {
	v.dataLock.Lock()
	defer v.dataLock.Unlock()
	v.vfo.frequency = frequency
	v.vfo.band = band
	v.vfo.roi = roi
	v.vfo.mode = mode
	v.vfo.bandwidth = bandwidth

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
		return v.vfo.roi
	}()
	scaleX := float64(v.view.GetAllocatedWidth()) / float64(vfoROI.Width())

	return vfoROI.From + core.Frequency(x/scaleX)
}

func (v *View) onDraw(da *gtk.DrawingArea, cr *cairo.Context) {
	data, vfo := func() ([]float64, vfo) {
		v.dataLock.RLock()
		defer v.dataLock.RUnlock()
		return v.fftData, v.vfo
	}()

	blockSize := len(data)
	if blockSize == 0 {
		return
	}

	hzPerBin := float64(vfo.roi.Width()) / float64(blockSize)

	height, width := float64(da.GetAllocatedHeight()), float64(da.GetAllocatedWidth())

	scaleX := float64(width) / float64(vfo.roi.Width())
	maxY := 100.0

	m := new(cairo.Matrix)
	m.InitTranslate(0, height)
	m.Scale(scaleX, -height/maxY)

	fillBackground(cr, width, height)
	dbmY := calculateInfoCoordinates(cr, m, vfo.frequency, vfo.roi, maxY)

	drawDBMScale(cr, dbmY, width)
	drawModeIndicator(cr, m, vfo.band, vfo.roi, height)
	drawFreqScale(cr, m, vfo.roi, height)
	drawVFO(cr, m, vfo, height)
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

func drawModeIndicator(cr *cairo.Context, matrix *cairo.Matrix, vfoBand bandplan.Band, vfoROI core.FrequencyRange, height float64) {
	cr.Save()
	defer cr.Restore()

	for _, p := range vfoBand.Portions {
		cr.Save()
		cr.Transform(matrix)
		startX, _ := cr.UserToDeviceDistance(float64(p.From-vfoROI.From), 0)
		endX, _ := cr.UserToDeviceDistance(float64(p.To-vfoROI.From), 0)
		cr.Restore()

		var y float64
		lineWidth := 10.0
		switch p.Mode {
		case bandplan.ModeCW:
			cr.SetSourceRGB(0.4, 0, 0.4)
			y = lineWidth * 0.5
		case bandplan.ModePhone:
			cr.SetSourceRGB(0.2, 0.4, 0)
			y = lineWidth * 0.5
		case bandplan.ModeDigital:
			cr.SetSourceRGB(0, 0, 0.6)
			y = lineWidth * 0.5
		case bandplan.ModeBeacon:
			cr.SetSourceRGB(1, 0, 0)
			y = lineWidth * 0.5
		case bandplan.ModeContest:
			cr.SetSourceRGB(0.6, 0.3, 0)
			y = lineWidth * 1.5
		}

		cr.SetLineWidth(lineWidth)
		cr.MoveTo(startX, y)
		cr.LineTo(endX, y)
		cr.Stroke()
	}

}

func drawFreqScale(cr *cairo.Context, matrix *cairo.Matrix, vfoROI core.FrequencyRange, height float64) {
	cr.Save()
	defer cr.Restore()

	fZeros := float64(int(math.Log10(float64(vfoROI.Width()))) - 1)
	fMagnitude := int(math.Pow(10, fZeros))
	fFactor := fMagnitude
	for int(vfoROI.Width())/fFactor > 20 {
		if fFactor%10 == 0 {
			fFactor *= 5
		} else {
			fFactor *= 10
		}
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

		freqText := fmt.Sprintf("%.0fk", float64(u.f)/1000.0)
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

func drawVFO(cr *cairo.Context, matrix *cairo.Matrix, vfo vfo, height float64) {
	cr.Save()
	defer cr.Restore()

	cr.Save()
	cr.Transform(matrix)
	vfoX, _ := cr.UserToDeviceDistance(float64(vfo.frequency-vfo.roi.From), 0)
	bandwidthFromX, _ := cr.UserToDeviceDistance(float64(vfo.frequency-(vfo.bandwidth/2)-vfo.roi.From), 0)
	bandwidth, _ := cr.UserToDeviceDistance(float64(vfo.bandwidth), 0)
	cr.Restore()

	cr.SetLineWidth(1.5)
	cr.SetSourceRGB(0.6, 0.9, 1.0)
	cr.MoveTo(vfoX, 0)
	cr.LineTo(vfoX, height)
	cr.Stroke()

	cr.SetSourceRGBA(0.6, 0.9, 1.0, 0.2)
	cr.Rectangle(bandwidthFromX, 0, bandwidth, height)
	cr.Fill()

	cr.SetFontSize(20.0)
	freqText := fmt.Sprintf("%.2fkHz", vfo.frequency/1000)
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

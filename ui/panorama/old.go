package panorama

import (
	"fmt"
	"math"
	"sync"

	"github.com/gotk3/gotk3/cairo"
	"github.com/gotk3/gotk3/gtk"

	"github.com/ftl/panacotta/core"
)

// func New(builder *gtk.Builder, controller Controller) *View {
// 	result := View{
// 		view:       ui.Get(builder, "panoramaView").(*gtk.DrawingArea),
// 		controller: controller,

// 		dataLock:       new(sync.RWMutex),
// 		redrawInterval: (1 * time.Second) / time.Duration(5),
// 	}
// 	result.view.Connect("draw", result.onDraw)
// 	result.connectMouse()
// 	result.connectKeyboard()

// 	return &result
// }

type vfo struct {
	frequency core.Frequency
	band      core.Band
	roi       core.FrequencyRange
	mode      string
	bandwidth core.Frequency
}

type oldView struct {
	view       *gtk.DrawingArea
	controller Controller

	fftData  []float64
	vfo      vfo
	dataLock *sync.RWMutex

	mouse    mouse
	keyboard keyboard
}

// SetFFTData sets the current FFT data.
func (v *oldView) SetFFTData(data []float64) {
	v.dataLock.Lock()
	defer v.dataLock.Unlock()
	v.fftData = data

	// v.TriggerRedraw()
}

// SetVFO sets the VFO configuration.
func (v *oldView) SetVFO(frequency core.Frequency, band core.Band, roi core.FrequencyRange, mode string, bandwidth core.Frequency) {
	v.dataLock.Lock()
	defer v.dataLock.Unlock()
	v.vfo.frequency = frequency
	v.vfo.band = band
	v.vfo.roi = roi
	v.vfo.mode = mode
	v.vfo.bandwidth = bandwidth

	// v.TriggerRedraw()
}

// TriggerRedraw triggers a redraw according to the redraw interval.
// func (v *oldView) TriggerRedraw() {
// 	now := time.Now()
// 	if now.Sub(v.lastRedraw) < v.redrawInterval {
// 		return
// 	}

// 	v.lastRedraw = now
// 	v.view.QueueDraw()
// }

func (v *oldView) deviceToFrequency(x float64) core.Frequency {
	vfoROI := func() core.FrequencyRange {
		v.dataLock.RLock()
		defer v.dataLock.RUnlock()
		return v.vfo.roi
	}()
	scaleX := float64(v.view.GetAllocatedWidth()) / float64(vfoROI.Width())

	return vfoROI.From + core.Frequency(x/scaleX)
}

func (v *oldView) onDraw(da *gtk.DrawingArea, cr *cairo.Context) {
	data, vfo := func() ([]float64, vfo) {
		v.dataLock.RLock()
		defer v.dataLock.RUnlock()
		return v.fftData, v.vfo
	}()

	blockSize := len(data)
	if blockSize == 0 {
		return
	}

	// hzPerBin := float64(vfo.roi.Width()) / float64(blockSize)

	height, width := float64(da.GetAllocatedHeight()), float64(da.GetAllocatedWidth())

	scaleX := float64(width) / float64(vfo.roi.Width())
	maxY := 100.0

	m := new(cairo.Matrix)
	m.InitTranslate(0, height)
	m.Scale(scaleX, -height/maxY)

	// fillBackground(cr, width, height)
	dbmY := calculateInfoCoordinates(cr, m, vfo.frequency, vfo.roi, maxY)

	drawDBMScale(cr, dbmY, width)
	// drawModeIndicator(cr, m, vfo.band, vfo.roi, height)
	drawFreqScale(cr, m, vfo.roi, height)
	// drawVFO(cr, m, vfo, height)
	// drawFFT(cr, m, data, hzPerBin)
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

func _drawModeIndicator(cr *cairo.Context, matrix *cairo.Matrix, vfoBand core.Band, vfoROI core.FrequencyRange, height float64) {
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
		case core.ModeCW:
			cr.SetSourceRGB(0.4, 0, 0.4)
			y = lineWidth * 0.5
		case core.ModePhone:
			cr.SetSourceRGB(0.2, 0.4, 0)
			y = lineWidth * 0.5
		case core.ModeDigital:
			cr.SetSourceRGB(0, 0, 0.6)
			y = lineWidth * 0.5
		case core.ModeBeacon:
			cr.SetSourceRGB(1, 0, 0)
			y = lineWidth * 0.5
		case core.ModeContest:
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

func _drawFFT(cr *cairo.Context, matrix *cairo.Matrix, data []float64, hzPerBin float64) {
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

func _drawVFO(cr *cairo.Context, matrix *cairo.Matrix, vfo vfo, height float64) {
	cr.Save()
	defer cr.Restore()

	cr.Save()
	cr.Transform(matrix)
	vfoX, _ := cr.UserToDeviceDistance(float64(vfo.frequency-vfo.roi.From), 0)
	bandwidthFromX, _ := cr.UserToDeviceDistance(float64(vfo.frequency-(vfo.bandwidth/2)-vfo.roi.From), 0)
	bandwidth, _ := cr.UserToDeviceDistance(float64(vfo.bandwidth), 0)
	cr.Restore()

	cr.SetFontSize(20.0)
	freqText := fmt.Sprintf("%.2fkHz", vfo.frequency/1000)
	vfoExtents := cr.TextExtents("VFO")
	freqExtents := cr.TextExtents(freqText)
	padding := 4.0
	boxWidth := math.Max(vfoExtents.Width, freqExtents.Width) + 2*padding
	boxHeight := vfoExtents.Height + freqExtents.Height + 3*padding

	cr.SetSourceRGBA(0.6, 0.9, 1.0, 0.2)
	cr.Rectangle(bandwidthFromX, 0, bandwidth, height)
	cr.Fill()

	cr.SetSourceRGBA(0, 0, 0, 0.75)
	cr.Rectangle(vfoX, 0, boxWidth, boxHeight)
	cr.Fill()

	cr.SetLineWidth(1.5)
	cr.SetSourceRGB(0.6, 0.9, 1.0)
	cr.MoveTo(vfoX, 0)
	cr.LineTo(vfoX, height)
	cr.Stroke()

	cr.SetSourceRGB(0.6, 0.9, 1.0)
	cr.MoveTo(vfoX+padding, vfoExtents.Height+padding)
	cr.ShowText("VFO")
	cr.MoveTo(vfoX+padding, vfoExtents.Height+freqExtents.Height+2*padding)
	cr.ShowText(freqText)
}
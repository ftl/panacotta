package panorama

import (
	"fmt"
	"math"

	"github.com/ftl/panacotta/core"
	"github.com/gotk3/gotk3/cairo"
	"github.com/gotk3/gotk3/gtk"
)

type rect struct {
	top, left, bottom, right float64
}

func (r rect) width() float64 {
	return math.Abs(r.left - r.right)
}

func (r rect) height() float64 {
	return math.Abs(r.top - r.bottom)
}

type geometry struct {
	widget         rect
	dbScale        rect
	bandIndicator  rect
	frequencyScale rect
	modeIndicator  rect
	fft            rect
	vfo            rect
}

func (v *View) onDraw(da *gtk.DrawingArea, cr *cairo.Context) {
	data := v.currentData()

	fillBackground(cr)

	var g geometry
	g.widget.bottom, g.widget.right = float64(da.GetAllocatedHeight()), float64(da.GetAllocatedWidth())

	g.dbScale = drawDBScale(cr, g, data)
	g.bandIndicator = drawBandIndicator(cr, g, data)
	g.frequencyScale = drawFrequencyScale(cr, g, data)
	g.modeIndicator = drawModeIndicator(cr, g, data)
	g.fft = drawFFT(cr, g, data)
	g.vfo = drawVFO(cr, g, data)
}

func (v *View) currentData() core.Panorama {
	v.dataLock.RLock()
	defer v.dataLock.RUnlock()
	return v.data
}

func fillBackground(cr *cairo.Context) {
	cr.Save()
	defer cr.Restore()

	cr.SetSourceRGB(0, 0, 0)
	cr.Paint()
}

func drawDBScale(cr *cairo.Context, g geometry, data core.Panorama) rect {
	cr.Save()
	defer cr.Restore()

	r := rect{
		right:  0,
		bottom: g.widget.bottom,
	}

	return r
}

func drawBandIndicator(cr *cairo.Context, g geometry, data core.Panorama) rect {
	cr.Save()
	defer cr.Restore()

	r := rect{
		right:  g.dbScale.right,
		bottom: g.widget.bottom,
	}

	return r
}

func drawFrequencyScale(cr *cairo.Context, g geometry, data core.Panorama) rect {
	cr.Save()
	defer cr.Restore()

	const spacing = float64(2.0)
	r := rect{
		left:  g.dbScale.right,
		right: g.widget.right,
	}

	cr.SetFontSize(10.0)
	extents := cr.TextExtents("Hg")
	r.bottom = extents.Height + 2*spacing

	cr.SetSourceRGB(0.8, 0.8, 0.8)
	cr.SetLineWidth(0.5)

	cr.SetDash([]float64{2, 2}, 0)
	for _, mark := range data.FrequencyScale {
		x := r.left + float64(mark.X)
		if x < r.left || x > r.right {
			continue
		}
		cr.MoveTo(x, r.top)
		cr.LineTo(x, g.widget.bottom)
		cr.Stroke()

		freqText := fmt.Sprintf("%.0fk", float64(mark.Frequency)/1000.0)
		cr.MoveTo(x+spacing, r.bottom-spacing)
		cr.ShowText(freqText)
	}

	return r
}

func drawModeIndicator(cr *cairo.Context, g geometry, data core.Panorama) rect {
	cr.Save()
	defer cr.Restore()

	const height = float64(5.0)
	r := rect{
		left:  g.dbScale.right,
		top:   g.frequencyScale.bottom,
		right: g.widget.right,
	}
	r.bottom = r.top + 2*height

	cr.SetLineWidth(1.0)

	for _, portion := range data.Band.Portions {
		startX := r.left + float64(data.ToPx(portion.From))
		endX := r.left + float64(data.ToPx(portion.To))
		if endX < r.left || startX > r.right {
			continue
		}
		startX = math.Max(r.left, startX)
		endX = math.Min(r.right, endX)

		var yOffset float64
		switch portion.Mode {
		case core.ModeCW:
			cr.SetSourceRGB(0.4, 0, 0.4)
		case core.ModePhone:
			cr.SetSourceRGB(0.2, 0.4, 0)
		case core.ModeDigital:
			cr.SetSourceRGB(0, 0, 0.6)
		case core.ModeBeacon:
			cr.SetSourceRGB(1, 0, 0)
		case core.ModeContest:
			cr.SetSourceRGB(0.6, 0.3, 0)
			yOffset = height
		}

		cr.Rectangle(startX, r.top+yOffset, endX-startX, height)
		cr.Fill()
	}

	return r
}

func drawFFT(cr *cairo.Context, g geometry, data core.Panorama) rect {
	cr.Save()
	defer cr.Restore()

	r := rect{
		left:   g.dbScale.right,
		right:  g.widget.right,
		top:    g.modeIndicator.bottom,
		bottom: g.widget.bottom,
	}

	if len(data.Spectrum) == 0 {
		return r
	}
	startX := r.left + float64(data.Spectrum[0].X)
	endX := r.left + float64(data.Spectrum[len(data.Spectrum)-1].X)
	centerX := startX + (endX-startX)/2

	cr.SetSourceRGB(1, 0, 0)
	cr.SetLineWidth(3)
	cr.MoveTo(startX, r.top)
	cr.LineTo(startX, r.bottom)
	cr.Stroke()
	cr.MoveTo(centerX, r.top)
	cr.LineTo(centerX, r.bottom)
	cr.Stroke()
	cr.MoveTo(endX, r.top)
	cr.LineTo(endX, r.bottom)
	cr.Stroke()

	// cr.SetSourceRGBA(1, 1, 1, 0.3)
	// cr.MoveTo(startX, r.bottom)
	// for _, p := range data.Spectrum {
	// 	cr.LineTo(r.left+float64(p.X), r.bottom-(float64(p.Y))
	// }
	// cr.MoveTo(r.left+float64(data.Spectrum[len(data.Spectrum)-1].X), r.bottom)
	// cr.ClosePath()
	// cr.Fill()

	cr.SetSourceRGB(1, 1, 1)
	cr.SetLineWidth(0.5)
	cr.MoveTo(startX, r.bottom-float64(data.Spectrum[0].Y))
	for _, p := range data.Spectrum {
		cr.LineTo(r.left+float64(p.X), r.bottom-float64(p.Y))
	}
	cr.Stroke()

	return r
}

func drawVFO(cr *cairo.Context, g geometry, data core.Panorama) rect {
	cr.Save()
	defer cr.Restore()

	r := rect{
		top:    g.fft.top,
		bottom: g.fft.bottom,
	}

	cr.SetFontSize(20.0)

	freqX := g.fft.left + float64(data.ToPx(data.VFO.Frequency))
	vfoExtents := cr.TextExtents("VFO")
	freqText := fmt.Sprintf("%.2fkHz", data.VFO.Frequency/1000)
	freqExtents := cr.TextExtents(freqText)
	padding := 4.0
	boxWidth := math.Max(vfoExtents.Width, freqExtents.Width) + 2*padding
	boxHeight := vfoExtents.Height + freqExtents.Height + 3*padding
	filterX := g.fft.left + float64(data.VFOFilterFrom)
	filterWidth := float64(data.VFOFilterTo - data.VFOFilterFrom)

	cr.SetSourceRGBA(0.6, 0.9, 1.0, 0.2)
	cr.Rectangle(filterX, r.top, filterWidth, r.height())
	cr.Fill()

	cr.SetSourceRGBA(0, 0, 0, 0.75)
	cr.Rectangle(freqX, r.top, boxWidth, boxHeight)
	cr.Fill()

	cr.SetLineWidth(1.5)
	cr.SetSourceRGB(0.6, 0.9, 1.0)
	cr.MoveTo(freqX, r.top)
	cr.LineTo(freqX, r.bottom)
	cr.Stroke()

	cr.SetSourceRGB(0.6, 0.9, 1.0)
	cr.MoveTo(freqX+padding, r.top+vfoExtents.Height+padding)
	cr.ShowText("VFO")
	cr.MoveTo(freqX+padding, r.top+vfoExtents.Height+freqExtents.Height+2*padding)
	cr.ShowText(freqText)

	return r
}

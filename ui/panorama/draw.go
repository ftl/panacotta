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

func (r rect) contains(p point) bool {
	return r.left <= p.x && r.right >= p.x && r.top <= p.y && r.bottom >= p.y
}

func (r rect) toX(f core.Frct) float64 {
	return r.left + r.width()*float64(f)
}

func (r rect) toY(f core.Frct) float64 {
	return r.bottom - r.height()*float64(f)
}

type point struct {
	x, y float64
}

type geometry struct {
	mouse          point
	widget         rect
	dbScale        rect
	bandIndicator  rect
	frequencyScale rect
	modeIndicator  rect
	fft            rect
	vfo            rect
	peaks          []rect
}

var dim = struct {
	spacing                float64
	modeIndicatorHeight    float64
	frequencyScaleFontSize float64
	dbScaleFontSize        float64
}{
	spacing:                2.0,
	modeIndicatorHeight:    5.0,
	frequencyScaleFontSize: 10.0,
	dbScaleFontSize:        10.0,
}

func (v *View) onDraw(da *gtk.DrawingArea, cr *cairo.Context) {
	data := v.data
	fillBackground(cr)

	g := v.prepareGeometry(da, cr)
	g.dbScale = drawDBScale(cr, g, data)
	g.bandIndicator = drawBandIndicator(cr, g, data)
	g.frequencyScale = drawFrequencyScale(cr, g, data)
	g.modeIndicator = drawModeIndicator(cr, g, data)
	g.fft = drawFFT(cr, g, data)
	g.peaks = drawPeaks(cr, g, data)
	g.vfo = drawVFO(cr, g, data)

	v.geometry = g
}

func fillBackground(cr *cairo.Context) {
	cr.Save()
	defer cr.Restore()

	cr.SetSourceRGB(0, 0, 0)
	cr.Paint()
}

func (v *View) prepareGeometry(da *gtk.DrawingArea, cr *cairo.Context) geometry {
	cr.Save()
	defer cr.Restore()

	result := geometry{
		mouse:  point{x: v.mouse.x, y: v.mouse.y},
		widget: rect{bottom: float64(da.GetAllocatedHeight()), right: float64(da.GetAllocatedWidth())},
	}

	cr.SetFontSize(dim.frequencyScaleFontSize)
	frequencyScaleExtents := cr.TextExtents("Hg")
	cr.SetFontSize(dim.dbScaleFontSize)
	dbScaleExtents := cr.TextExtents("-000.0dB")

	result.frequencyScale.bottom = frequencyScaleExtents.Height + 2*dim.spacing
	result.modeIndicator.bottom = result.frequencyScale.bottom + 2*dim.modeIndicatorHeight
	result.dbScale.right = dbScaleExtents.Width + 2*dim.spacing
	result.fft = rect{
		top:    result.modeIndicator.bottom,
		left:   result.dbScale.right,
		bottom: result.widget.bottom,
		right:  result.widget.right,
	}

	return result
}

func drawDBScale(cr *cairo.Context, g geometry, data core.Panorama) rect {
	cr.Save()
	defer cr.Restore()

	r := rect{
		right:  g.fft.left,
		top:    g.fft.top,
		bottom: g.fft.bottom,
	}

	cr.SetFontSize(dim.dbScaleFontSize)
	cr.SetSourceRGB(0.8, 0.8, 0.8)
	cr.SetLineWidth(0.5)
	cr.SetDash([]float64{2, 2}, 0)
	for _, mark := range data.DBScale {
		y := r.toY(mark.Y)
		cr.MoveTo(r.right, y)
		cr.LineTo(g.widget.right, y)
		// TODO maybe use a color indication for the signal level similar to the waterfall
		cr.Stroke()

		dbText := fmt.Sprintf("%.0fdB", mark.DB)
		extents := cr.TextExtents(dbText)
		cr.MoveTo(r.right-extents.Width-dim.spacing, y+extents.Height/2)
		cr.ShowText(dbText)
	}

	cr.SetSourceRGB(1.0, 0.3, 0.3)
	cr.SetLineWidth(1.0)
	cr.SetDash([]float64{2, 2}, 0)
	y := r.toY(data.PeakThresholdLine)
	cr.MoveTo(r.left, y)
	cr.LineTo(g.widget.right, y)
	cr.Stroke()

	return r
}

func drawBandIndicator(cr *cairo.Context, g geometry, data core.Panorama) rect {
	cr.Save()
	defer cr.Restore()

	r := rect{
		left:   g.dbScale.left,
		right:  g.dbScale.right,
		bottom: g.dbScale.top,
	}
	mouseOver := r.contains(g.mouse)

	if mouseOver {
		cr.SetSourceRGB(1, 1, 1)
	} else {
		cr.SetSourceRGB(0.8, 0.8, 0.8)
	}
	cr.SetFontSize(15.0)

	bandText := string(data.Band.Name)
	extents := cr.TextExtents(bandText)
	x := (r.right - extents.Width - dim.spacing)
	y := (r.bottom + extents.Height) / 2

	cr.MoveTo(x, y)
	cr.ShowText(bandText)

	cr.SetSourceRGB(0.8, 0.8, 0.8)
	cr.SetLineWidth(0.5)
	cr.MoveTo(r.left, r.bottom)
	cr.LineTo(r.right, r.bottom)
	cr.Stroke()

	return r
}

func drawFrequencyScale(cr *cairo.Context, g geometry, data core.Panorama) rect {
	cr.Save()
	defer cr.Restore()

	r := rect{
		left:   g.fft.left,
		right:  g.fft.right,
		bottom: g.frequencyScale.bottom,
	}

	cr.SetFontSize(dim.frequencyScaleFontSize)
	cr.SetSourceRGB(0.8, 0.8, 0.8)
	cr.SetLineWidth(0.5)
	cr.SetDash([]float64{2, 2}, 0)
	for _, mark := range data.FrequencyScale {
		x := r.toX(mark.X)
		if x < r.left || x > r.right {
			continue
		}
		cr.MoveTo(x, r.top)
		cr.LineTo(x, g.widget.bottom)
		cr.Stroke()

		freqText := fmt.Sprintf("%.0fk", float64(mark.Frequency)/1000.0)
		cr.MoveTo(x+dim.spacing, r.bottom-dim.spacing)
		cr.ShowText(freqText)
	}

	return r
}

func drawModeIndicator(cr *cairo.Context, g geometry, data core.Panorama) rect {
	cr.Save()
	defer cr.Restore()

	r := rect{
		left:   g.frequencyScale.left,
		top:    g.frequencyScale.bottom,
		right:  g.frequencyScale.right,
		bottom: g.modeIndicator.bottom,
	}

	cr.SetLineWidth(1.0)

	for _, portion := range data.Band.Portions {
		startX := r.toX(data.FrequencyRange.ToFrct(portion.From))
		endX := r.toX(data.FrequencyRange.ToFrct(portion.To))
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
			yOffset = dim.modeIndicatorHeight
		}

		cr.Rectangle(startX, r.top+yOffset, endX-startX, dim.modeIndicatorHeight)
		cr.Fill()
	}

	return r
}

func drawFFT(cr *cairo.Context, g geometry, data core.Panorama) rect {
	cr.Save()
	defer cr.Restore()

	r := g.fft

	if len(data.Spectrum) == 0 {
		return r
	}
	startX := r.toX(data.Spectrum[0].X)

	cr.SetSourceRGBA(1, 1, 1, 0.3)
	cr.MoveTo(startX, r.bottom)
	for _, p := range data.Spectrum {
		cr.LineTo(r.toX(p.X), r.toY(p.Y))
	}
	cr.LineTo(r.toX(data.Spectrum[len(data.Spectrum)-1].X), r.bottom)
	cr.ClosePath()
	cr.Fill()

	cr.SetSourceRGB(1, 1, 1)
	cr.SetLineWidth(1.0)
	cr.MoveTo(startX, r.toY(data.Spectrum[0].Y))
	for _, p := range data.Spectrum {
		cr.LineTo(r.toX(p.X), r.toY(p.Y))
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

	freqX := g.fft.toX(data.VFOLine)
	padding := 4.0
	filterX := g.fft.toX(data.VFOFilterFrom)
	filterWidth := g.fft.toX(data.VFOFilterTo) - g.fft.toX(data.VFOFilterFrom)
	r.left = filterX
	r.right = filterX + filterWidth
	mouseOver := r.contains(g.mouse)

	if mouseOver {
		cr.SetSourceRGBA(0.6, 0.9, 1.0, 0.5)
	} else {
		cr.SetSourceRGBA(0.6, 0.9, 1.0, 0.2)
	}
	cr.Rectangle(filterX, r.top, filterWidth, r.height())
	cr.Fill()

	cr.SetLineWidth(1.5)
	cr.SetSourceRGB(0.6, 0.9, 1.0)
	cr.MoveTo(freqX, r.top)
	cr.LineTo(freqX, r.bottom)
	cr.Stroke()

	cr.SetFontSize(15.0)
	freqText := fmt.Sprintf("%s:%.2fkHz", data.VFO.Name, data.VFO.Frequency/1000)
	freqExtents := cr.TextExtents(freqText)
	leftSide := freqX+padding+freqExtents.Width < g.fft.right
	if leftSide {
		cr.MoveTo(freqX+padding, r.top+freqExtents.Height+padding)
	} else {
		cr.MoveTo(freqX-padding-freqExtents.Width, r.top+freqExtents.Height+padding)
	}
	cr.ShowText(freqText)

	cr.SetFontSize(10.0)
	sMeterText := core.SUnit(data.VFOSignalLevel).String()
	sMeterExtents := cr.TextExtents(sMeterText)
	if leftSide {
		cr.MoveTo(freqX+padding, r.top+freqExtents.Height+sMeterExtents.Height+2*padding)
	} else {
		cr.MoveTo(freqX-padding-sMeterExtents.Width, r.top+freqExtents.Height+sMeterExtents.Height+2*padding)
	}
	cr.ShowText(sMeterText)

	return r
}

func drawPeaks(cr *cairo.Context, g geometry, data core.Panorama) []rect {
	cr.Save()
	defer cr.Restore()

	padding := 4.0

	result := make([]rect, len(data.Peaks))
	for i, peak := range data.Peaks {
		fromX := g.fft.toX(peak.FromX)
		toX := g.fft.toX(peak.ToX)
		maxX := g.fft.toX(peak.MaxX)
		y := g.fft.toY(peak.ValueY)
		r := rect{
			left:   fromX,
			top:    g.fft.top,
			right:  toX,
			bottom: g.fft.bottom,
		}
		mouseOver := r.contains(g.mouse)

		cr.SetFontSize(10.0)
		freqText := fmt.Sprintf("%.2fkHz", peak.MaxFrequency/1000)
		freqExtents := cr.TextExtents(freqText)
		leftSide := maxX+padding+freqExtents.Width < g.fft.right

		sMeterText := core.SUnit(peak.ValueDB).String()
		sMeterExtents := cr.TextExtents(sMeterText)

		if mouseOver {
			cr.SetSourceRGBA(0.3, 1, 0.8, 0.4)
			cr.Rectangle(r.left, r.top, r.width(), r.height())
			cr.Fill()

			cr.SetSourceRGB(0.3, 1, 0.8)
			if leftSide {
				cr.MoveTo(maxX+padding, y+padding)
			} else {
				cr.MoveTo(maxX-padding-freqExtents.Width, y+padding)
			}
			cr.ShowText(freqText)
			if leftSide {
				cr.MoveTo(maxX+padding, y+freqExtents.Height+2*padding)
			} else {
				cr.MoveTo(maxX-padding-sMeterExtents.Width, y+freqExtents.Height+2*padding)
			}
			cr.ShowText(sMeterText)
		} else {
			cr.SetSourceRGBA(0.3, 1, 0.8, 0.2)
		}

		cr.SetSourceRGBA(0.3, 1, 0.8, 0.4)
		cr.SetLineWidth(1.5)
		cr.MoveTo(maxX, y) // r.top
		cr.LineTo(maxX, r.bottom)
		cr.Stroke()

		cr.SetSourceRGB(0.3, 1, 0.8)
		cr.SetFontSize(12.0)
		markText := "\u25BC"
		markExtents := cr.TextExtents(markText)
		cr.MoveTo(maxX-markExtents.Width/2, y)
		cr.ShowText(markText)

		result[i] = r
	}

	return result
}

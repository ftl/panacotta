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

func (v *View) onDraw(da *gtk.DrawingArea, cr *cairo.Context) {
	data := v.data

	fillBackground(cr)

	var g geometry
	g.mouse = point{v.mouse.x, v.mouse.y}
	g.widget.bottom, g.widget.right = float64(da.GetAllocatedHeight()), float64(da.GetAllocatedWidth())
	g.fft.left = float64(v.fftTopLeft.X)
	g.fft.top = float64(v.fftTopLeft.Y)

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

func drawDBScale(cr *cairo.Context, g geometry, data core.Panorama) rect {
	cr.Save()
	defer cr.Restore()

	const spacing = float64(2.0)
	r := rect{
		right:  g.fft.left,
		top:    g.fft.top,
		bottom: g.widget.bottom,
	}

	cr.SetFontSize(10.0)
	cr.SetSourceRGB(0.8, 0.8, 0.8)
	cr.SetLineWidth(0.5)
	cr.SetDash([]float64{2, 2}, 0)
	for _, mark := range data.DBScale {
		y := r.bottom - float64(mark.Y)
		cr.MoveTo(r.right, y)
		cr.LineTo(g.widget.right, y)
		// TODO maybe use a color indication for the signal level similar to the waterfall
		cr.Stroke()

		dbText := fmt.Sprintf("%.0fdB", mark.DB)
		extents := cr.TextExtents(dbText)
		cr.MoveTo(r.right-extents.Width-spacing, y+extents.Height/2)
		cr.ShowText(dbText)
	}

	cr.SetSourceRGB(1.0, 0.3, 0.3)
	cr.SetLineWidth(1.0)
	cr.SetDash([]float64{2, 2}, 0)
	y := r.bottom - float64(data.MeanLine)
	cr.MoveTo(r.left, y)
	cr.LineTo(g.widget.right, y)
	cr.Stroke()

	return r
}

func drawBandIndicator(cr *cairo.Context, g geometry, data core.Panorama) rect {
	cr.Save()
	defer cr.Restore()

	const spacing = float64(2.0)
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
	x := (r.right - extents.Width - spacing)
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

	const spacing = float64(2.0)
	r := rect{
		left:  g.fft.left,
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
		left:  g.frequencyScale.left,
		top:   g.frequencyScale.bottom,
		right: g.frequencyScale.right,
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
		left:   g.fft.left,
		right:  g.widget.right,
		top:    g.fft.top,
		bottom: g.widget.bottom,
	}

	if len(data.Spectrum) == 0 {
		return r
	}
	startX := r.left + float64(data.Spectrum[0].X)

	cr.SetSourceRGBA(1, 1, 1, 0.3)
	cr.MoveTo(startX, r.bottom)
	for _, p := range data.Spectrum {
		cr.LineTo(r.left+float64(p.X), r.bottom-float64(p.Y))
	}
	cr.LineTo(r.left+float64(data.Spectrum[len(data.Spectrum)-1].X), r.bottom)
	cr.ClosePath()
	cr.Fill()

	cr.SetSourceRGB(1, 1, 1)
	cr.SetLineWidth(1.0)
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

	cr.SetFontSize(15.0)
	freqText := fmt.Sprintf("%s:%.2fkHz", data.VFO.Name, data.VFO.Frequency/1000)
	freqExtents := cr.TextExtents(freqText)

	freqX := g.fft.left + float64(data.ToPx(data.VFO.Frequency))

	padding := 4.0
	filterX := g.fft.left + float64(data.VFOFilterFrom)
	filterWidth := float64(data.VFOFilterTo - data.VFOFilterFrom)
	r.left = filterX
	r.right = filterX + filterWidth
	leftSide := freqX+padding+freqExtents.Width < g.fft.right
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

	if leftSide {
		cr.MoveTo(freqX+padding, r.top+freqExtents.Height+padding)
	} else {
		cr.MoveTo(freqX-padding-freqExtents.Width, r.top+freqExtents.Height+padding)
	}
	cr.ShowText(freqText)

	return r
}

func drawPeaks(cr *cairo.Context, g geometry, data core.Panorama) []rect {
	cr.Save()
	defer cr.Restore()

	padding := 4.0

	result := make([]rect, len(data.Peaks))
	for i, peak := range data.Peaks {
		fromX := g.fft.left + float64(peak.FromX)
		toX := g.fft.left + float64(peak.ToX)
		maxX := g.fft.left + float64(peak.MaxX)
		y := g.fft.bottom - float64(peak.ValueY)
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
		freqX := g.fft.left + float64(data.ToPx(peak.MaxFrequency))
		leftSide := freqX+padding+freqExtents.Width < g.fft.right

		if mouseOver {
			cr.SetSourceRGBA(0.3, 1, 0.8, 0.4)
			cr.Rectangle(r.left, r.top, r.width(), r.height())
			cr.Fill()

			cr.SetSourceRGB(0.3, 1, 0.8)
			if leftSide {
				cr.MoveTo(freqX+padding, y+padding)
			} else {
				cr.MoveTo(freqX-padding-freqExtents.Width, y+padding)
			}
			cr.ShowText(freqText)
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

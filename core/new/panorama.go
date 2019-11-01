package new

import (
	"math"

	"github.com/ftl/panacotta/core"
)

// Panorama controller
type Panorama struct {
	width          core.Px
	frequencyRange core.FrequencyRange
	vfo            core.Frequency
	resolution     map[ViewMode]core.HzPerPx
	viewMode       ViewMode
}

// ViewMode of the panorama.
type ViewMode int

// All view modes.
const (
	ViewFixed ViewMode = iota
	ViewCentered
)

const (
	defaultFixedResolution    = core.HzPerPx(100)
	defaultCenteredResolution = core.HzPerPx(10)
)

// NewPanorama returns a new instance of panorama.
func NewPanorama(width core.Px, frequencyRange core.FrequencyRange, vfo core.Frequency) *Panorama {
	result := Panorama{
		width:          width,
		frequencyRange: frequencyRange,
		vfo:            vfo,
		resolution: map[ViewMode]core.HzPerPx{
			ViewFixed:    calcResolution(frequencyRange, width),
			ViewCentered: defaultCenteredResolution,
		},
		viewMode: ViewFixed,
	}
	return &result
}

func calcResolution(frequencyRange core.FrequencyRange, width core.Px) core.HzPerPx {
	return core.HzPerPx(float64(frequencyRange.Width()) / float64(width))
}

func (p *Panorama) updateFrequencyRange() {
	var lowerRatio, upperRatio core.Frequency
	if p.viewMode == ViewFixed && p.frequencyRange.Contains(p.vfo) {
		lowerRatio = (p.vfo - p.frequencyRange.From) / p.frequencyRange.Width()
		lowerRatio = core.Frequency(math.Max(0.1, math.Min(float64(lowerRatio), 0.9)))
		upperRatio = 1.0 - lowerRatio
	} else {
		lowerRatio = 0.5
		upperRatio = 0.5
	}

	frequencyWidth := core.Frequency(float64(p.width) * float64(p.resolution[p.viewMode]))
	p.frequencyRange.From = p.vfo - (lowerRatio * frequencyWidth)
	p.frequencyRange.To = p.vfo + (upperRatio * frequencyWidth)
}

// SetWidth in pixels
func (p *Panorama) SetWidth(width core.Px) {
	p.width = width
	p.updateFrequencyRange()
}

// From in Hz
func (p Panorama) From() core.Frequency {
	return p.frequencyRange.From
}

// To in Hz
func (p Panorama) To() core.Frequency {
	return p.frequencyRange.To
}

// Bandwidth in Hz
func (p Panorama) Bandwidth() core.Frequency {
	return p.frequencyRange.Width()
}

// SetVFO in Hz
func (p *Panorama) SetVFO(vfo core.Frequency) {
	p.vfo = vfo

	if p.frequencyRange.Contains(vfo) {
		p.updateFrequencyRange()
	}
}

// VFO frequency in Hz
func (p Panorama) VFO() core.Frequency {
	return p.vfo
}

// ToggleViewMode switches to the other view mode.
func (p *Panorama) ToggleViewMode() {
	if p.viewMode == ViewFixed {
		p.viewMode = ViewCentered
	} else {
		p.viewMode = ViewFixed
	}
	p.updateFrequencyRange()
}

// ZoomIn one step
func (p *Panorama) ZoomIn() {
	p.resolution[p.viewMode] /= 1.25
	p.updateFrequencyRange()
}

// ZoomOut one step
func (p *Panorama) ZoomOut() {
	p.resolution[p.viewMode] *= 1.25
	p.updateFrequencyRange()
}

// ZoomTo the given frequency range and switch to fixed view mode.
func (p *Panorama) ZoomTo(frequencyRange core.FrequencyRange) {
	p.viewMode = ViewFixed
	p.frequencyRange = frequencyRange
	p.resolution[p.viewMode] = calcResolution(p.frequencyRange, p.width)
}

// ResetZoom to the default of the current view mode
func (p *Panorama) ResetZoom() {
	switch p.viewMode {
	case ViewFixed:
		p.resolution[p.viewMode] = defaultFixedResolution
	case ViewCentered:
		p.resolution[p.viewMode] = defaultCenteredResolution
	}
	p.updateFrequencyRange()
}

// Drag the panorama horizontally by a certain amount of Px.
func (p *Panorama) Drag(Δ core.Px) {
	Δf := core.Frequency(float64(Δ) * float64(p.resolution[p.viewMode]))
	p.frequencyRange.Shift(Δf)
}

// FrequencyAt the given y coordinate within the panorama.
func (p Panorama) FrequencyAt(y core.Px) core.Frequency {
	return p.frequencyRange.From + core.Frequency(float64(y)*float64(p.resolution[p.viewMode]))
}

// Spectrum as a block of magnitude values. blocksize >= width
func (p Panorama) Spectrum() []float64 {
	return []float64{}
}

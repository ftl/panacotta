package new

import (
	"math"

	"github.com/ftl/panacotta/core"
	"github.com/ftl/panacotta/core/bandplan"
)

// Panorama controller
type Panorama struct {
	width          core.Px
	frequencyRange core.FrequencyRange
	vfoFrequency   core.Frequency
	vfoBand        bandplan.Band
	vfoFilterWidth core.Frequency
	vfoMode        string

	resolution map[ViewMode]core.HzPerPx
	viewMode   ViewMode

	fftData  []float64
	fftRange core.FrequencyRange
}

// PanoramaData contains everything to draw the panorama at a specific point in time.
type PanoramaData struct {
	FrequencyRange core.FrequencyRange
	VFOFrequency   core.Frequency
	VFOFilterWidth core.Frequency
	VFOMode        string
	VFOBand        bandplan.Band

	VFOLine        core.Px
	VFOFilterFrom  core.Px
	VFOFilterTo    core.Px
	FrequencyScale []core.FrequencyMark
	Spectrum       []core.PxPoint
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
func NewPanorama(width core.Px, frequencyRange core.FrequencyRange, vfoFrequency core.Frequency) *Panorama {
	result := Panorama{
		width:          width,
		frequencyRange: frequencyRange,
		vfoFrequency:   vfoFrequency,
		vfoBand:        bandplan.UnknownBand,
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
	if p.viewMode == ViewFixed && p.frequencyRange.Contains(p.vfoFrequency) {
		lowerRatio = (p.vfoFrequency - p.frequencyRange.From) / p.frequencyRange.Width()
		lowerRatio = core.Frequency(math.Max(0.1, math.Min(float64(lowerRatio), 0.9)))
		upperRatio = 1.0 - lowerRatio
	} else {
		lowerRatio = 0.5
		upperRatio = 0.5
	}

	frequencyWidth := core.Frequency(float64(p.width) * float64(p.resolution[p.viewMode]))
	p.frequencyRange.From = p.vfoFrequency - (lowerRatio * frequencyWidth)
	p.frequencyRange.To = p.vfoFrequency + (upperRatio * frequencyWidth)
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
func (p *Panorama) SetVFO(frequency core.Frequency, filterWidth core.Frequency, mode string) {
	p.vfoFrequency = frequency
	p.vfoFilterWidth = filterWidth
	p.vfoMode = mode

	if !p.vfoBand.Contains(frequency) {
		band := bandplan.IARURegion1.ByFrequency(frequency)
		if band.Width() > 0 {
			p.vfoBand = band
		}
	}

	if p.frequencyRange.Contains(frequency) {
		p.updateFrequencyRange()
	}
}

// VFO frequency in Hz
func (p Panorama) VFO() (frequency core.Frequency, filterWidth core.Frequency, mode string, band bandplan.Band) {
	return p.vfoFrequency, p.vfoFilterWidth, p.vfoMode, p.vfoBand
}

// SetFFT data
func (p *Panorama) SetFFT(data []float64, frequencyRange core.FrequencyRange) {
	p.fftData = data
	p.fftRange = frequencyRange
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

// Data to draw the current panorama.
func (p Panorama) Data() *PanoramaData {
	resolution := p.resolution[p.viewMode]
	result := PanoramaData{
		FrequencyRange: p.frequencyRange,
		VFOFrequency:   p.vfoFrequency,
		VFOFilterWidth: p.vfoFilterWidth,
		VFOMode:        p.vfoMode,
		VFOBand:        p.vfoBand,

		VFOLine: resolution.ToPx(p.vfoFrequency - p.frequencyRange.From),

		FrequencyScale: p.frequencyScale(),
		Spectrum:       p.spectrum(),
	}

	result.VFOFilterFrom = result.VFOLine - resolution.ToPx(p.vfoFilterWidth/2)
	result.VFOFilterTo = result.VFOLine + resolution.ToPx(p.vfoFilterWidth/2)

	return &result
}

func (p Panorama) frequencyScale() []core.FrequencyMark {
	resolution := p.resolution[p.viewMode]
	fZeros := float64(int(math.Log10(float64(p.frequencyRange.Width()))) - 1)
	fMagnitude := int(math.Pow(10, fZeros))
	fFactor := fMagnitude
	for resolution.ToPx(core.Frequency(fFactor)) < 200.0 {
		if fFactor%10 == 0 {
			fFactor *= 5
		} else {
			fFactor *= 10
		}
	}
	for resolution.ToPx(core.Frequency(fFactor)) > 300.0 {
		if fFactor%10 == 0 {
			fFactor /= 5
		} else {
			fFactor /= 10
		}
	}

	freqScale := make([]core.FrequencyMark, 0, int(p.frequencyRange.Width())/fFactor)
	for f := core.Frequency((int(p.frequencyRange.From) / fFactor) * fFactor); f < p.frequencyRange.To; f += core.Frequency(fFactor) {
		x := resolution.ToPx(f - p.frequencyRange.From)
		mark := core.FrequencyMark{
			X:         x,
			Frequency: f,
		}
		freqScale = append(freqScale, mark)
	}

	return freqScale
}

func (p Panorama) spectrum() []core.PxPoint {
	resolution := p.resolution[p.viewMode]
	fftResolution := float64(p.fftRange.Width()) / float64(len(p.fftData))
	result := make([]core.PxPoint, len(p.fftData))
	for i, d := range p.fftData {
		freq := p.fftRange.From * core.Frequency(float64(i)*fftResolution)
		result[i] = core.PxPoint{
			X: resolution.ToPx(freq - p.frequencyRange.From),
			Y: core.Px(d),
		}
	}
	return result
}

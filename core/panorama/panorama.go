package panorama

import (
	"log"
	"math"

	"github.com/ftl/panacotta/core"
)

// Panorama controller
type Panorama struct {
	width          core.Px
	height         core.Px
	frequencyRange core.FrequencyRange
	dbRange        core.DBRange
	vfo            core.VFO
	band           core.Band

	resolution    map[ViewMode]core.HzPerPx
	viewMode      ViewMode
	fullRangeMode bool

	fft core.FFT
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

// New returns a new instance of panorama.
func New(width core.Px, frequencyRange core.FrequencyRange, vfoFrequency core.Frequency) *Panorama {
	result := Panorama{
		width:          width,
		frequencyRange: frequencyRange,
		dbRange:        core.DBRange{From: -125, To: 15},
		resolution: map[ViewMode]core.HzPerPx{
			ViewFixed:    calcResolution(frequencyRange, width),
			ViewCentered: defaultCenteredResolution,
		},
		viewMode: ViewFixed,
	}

	result.vfo.Frequency = vfoFrequency

	return &result
}

// NewFullSpectrum returns a new instance of panorama in full-range mode.
func NewFullSpectrum(width core.Px, frequencyRange core.FrequencyRange, vfoFrequency core.Frequency) *Panorama {
	result := New(width, frequencyRange, vfoFrequency)
	result.fullRangeMode = true
	return result
}

func calcResolution(frequencyRange core.FrequencyRange, width core.Px) core.HzPerPx {
	return core.HzPerPx(float64(frequencyRange.Width()) / float64(width))
}

func (p *Panorama) updateFrequencyRange() {
	if math.IsNaN(float64(p.resolution[p.viewMode])) {
		p.setupFrequencyRange()
		return
	}

	var lowerRatio, upperRatio core.Frequency
	if p.viewMode == ViewFixed && p.frequencyRange.Contains(p.vfo.Frequency) {
		lowerRatio = (p.vfo.Frequency - p.frequencyRange.From) / p.frequencyRange.Width()
		lowerRatio = core.Frequency(math.Max(0.1, math.Min(float64(lowerRatio), 0.9)))
		upperRatio = 1.0 - lowerRatio
	} else {
		lowerRatio = 0.5
		upperRatio = 0.5
	}

	frequencyWidth := core.Frequency(float64(p.width) * float64(p.resolution[p.viewMode]))
	p.frequencyRange.From = p.vfo.Frequency - (lowerRatio * frequencyWidth)
	p.frequencyRange.To = p.vfo.Frequency + (upperRatio * frequencyWidth)

	log.Printf("frequency range %v %v %v", p.frequencyRange, p.frequencyRange.Width(), p.resolution[p.viewMode])
}

func (p *Panorama) setupFrequencyRange() {
	if p.vfo.Frequency == 0 || !p.band.Contains(p.vfo.Frequency) {
		return
	}

	if p.viewMode == ViewFixed {
		p.frequencyRange.From = p.band.From - 1000
		p.frequencyRange.To = p.band.To + 1000
	} else {
		p.frequencyRange.From = p.vfo.Frequency - 20000
		p.frequencyRange.From = p.vfo.Frequency + 20000
	}
	p.resolution[p.viewMode] = calcResolution(p.frequencyRange, p.width)

	log.Printf("frequency range %v %v %v", p.frequencyRange, p.frequencyRange.Width(), p.resolution[p.viewMode])
}

// SetSize in pixels
func (p *Panorama) SetSize(width, height core.Px) {
	if (width == p.width) && (height == p.height) {
		return
	}

	log.Printf("width %v height %v", width, height)

	p.width = width
	p.height = height
	p.updateFrequencyRange()
}

// FrequencyRange of the panorama
func (p Panorama) FrequencyRange() core.FrequencyRange {
	return p.frequencyRange
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
func (p *Panorama) SetVFO(vfo core.VFO) {
	p.vfo = vfo

	if !p.band.Contains(vfo.Frequency) {
		band := core.IARURegion1.ByFrequency(vfo.Frequency)
		if band.Width() > 0 {
			p.band = band
		}
	}

	log.Printf("vfo %v band %v", p.vfo, p.band)

	p.updateFrequencyRange()
}

// VFO frequency in Hz
func (p Panorama) VFO() (vfo core.VFO, band core.Band) {
	return p.vfo, p.band
}

// SetFFT data
func (p *Panorama) SetFFT(fft core.FFT) {
	p.fft = fft
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

// Drag the panorama horizontally by a certain amount of Hz.
func (p *Panorama) Drag(Δf core.Frequency) {
	p.frequencyRange.Shift(Δf)
}

// Data to draw the current panorama.
func (p Panorama) Data() core.Panorama {
	if p.fullRangeMode {
		return p.fullRangeData()
	}
	return p.data()
}

func (p Panorama) data() core.Panorama {
	resolution := p.resolution[p.viewMode]
	result := core.Panorama{
		FrequencyRange: p.frequencyRange,
		VFO:            p.vfo,
		Band:           p.band,
		Resolution:     p.resolution[p.viewMode],

		VFOLine: resolution.ToPx(p.vfo.Frequency - p.frequencyRange.From),

		FrequencyScale: p.frequencyScale(),
		DBScale:        p.dbScale(),
		Spectrum:       p.spectrum(),
	}

	result.VFOFilterFrom = result.VFOLine - resolution.ToPx(p.vfo.FilterWidth/2)
	result.VFOFilterTo = result.VFOLine + resolution.ToPx(p.vfo.FilterWidth/2)

	return result
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

func (p Panorama) dbScale() []core.DBMark {
	pxPerDB := float64(p.height) / float64(p.dbRange.Width())
	startDB := int(p.dbRange.From) - int(p.dbRange.From)%10
	markCount := (int(p.dbRange.To) - startDB) / 10
	if (int(p.dbRange.To)-startDB)%10 != 0 {
		markCount++
	}

	dbScale := make([]core.DBMark, markCount)
	for i := range dbScale {
		db := core.DB(startDB + i*10)
		dbScale[i] = core.DBMark{
			DB: db,
			Y:  core.Px(float64(db-p.dbRange.From) * pxPerDB),
		}
	}

	return dbScale
}

func (p Panorama) spectrum() []core.PxPoint {
	if len(p.fft.Data) == 0 || p.fft.Range.To < p.frequencyRange.From || p.fft.Range.From > p.frequencyRange.To {
		return []core.PxPoint{}
	}
	resolution := p.resolution[p.viewMode]
	fftResolution := float64(p.fft.Range.Width()) / float64(len(p.fft.Data))
	step := int(math.Max(1, math.Floor(float64(len(p.fft.Data))/float64(p.width))))
	start := int(math.Max(0, math.Floor(float64(p.frequencyRange.From-p.fft.Range.From)/fftResolution)))
	end := int(math.Min(float64(len(p.fft.Data)-1), math.Ceil(float64(p.frequencyRange.To-p.fft.Range.From)/fftResolution)))
	resultLength := (end - start + 1) / step
	if (end-start+1)%step != 0 {
		resultLength++
	}
	result := make([]core.PxPoint, resultLength)
	resultIndex := 0
	for i := start; i <= end; i += step {
		d := 0.0
		for j := i; j < i+step && j < len(p.fft.Data); j++ {
			d += p.fft.Data[j]
		}
		d /= float64(step)

		freq := p.fft.Range.From + core.Frequency(float64(i)*fftResolution)
		result[resultIndex] = core.PxPoint{
			X: resolution.ToPx(freq - p.frequencyRange.From),
			Y: core.Px((core.DB(d)-p.dbRange.From)/p.dbRange.Width()) * p.height,
		}
		resultIndex++
	}
	return result
}

func (p Panorama) fullRangeData() core.Panorama {
	if p.fft.Range.Width() == 0 || p.width == 0 {
		return core.Panorama{}
	}

	resolution := core.HzPerPx(float64(p.fft.Range.Width()) / float64(p.width))
	frequencyRange := p.fft.Range
	result := core.Panorama{
		FrequencyRange: frequencyRange,
		VFO:            p.vfo,
		Band:           p.band,
		Resolution:     resolution,

		VFOLine: resolution.ToPx(p.vfo.Frequency - frequencyRange.From),

		FrequencyScale: p.fullRangeFrequencyScale(),
		DBScale:        p.dbScale(),
		Spectrum:       p.fullRangeSpectrum(),
	}

	result.VFOFilterFrom = result.VFOLine - resolution.ToPx(p.vfo.FilterWidth/2)
	result.VFOFilterTo = result.VFOLine + resolution.ToPx(p.vfo.FilterWidth/2)

	return result
}

func (p Panorama) fullRangeFrequencyScale() []core.FrequencyMark {
	resolution := core.HzPerPx(float64(p.fft.Range.Width()) / float64(p.width))
	frequencyRange := p.fft.Range

	fZeros := float64(int(math.Log10(float64(frequencyRange.Width()))) - 1)
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

	freqScale := make([]core.FrequencyMark, 0, int(frequencyRange.Width())/fFactor)
	for f := core.Frequency((int(frequencyRange.From) / fFactor) * fFactor); f < frequencyRange.To; f += core.Frequency(fFactor) {
		x := resolution.ToPx(f - frequencyRange.From)
		mark := core.FrequencyMark{
			X:         x,
			Frequency: f,
		}
		freqScale = append(freqScale, mark)
	}

	return freqScale
}

func (p Panorama) fullRangeSpectrum() []core.PxPoint {
	resolution := core.HzPerPx(float64(p.fft.Range.Width()) / float64(p.width))
	frequencyRange := p.fft.Range

	fftResolution := float64(p.fft.Range.Width()) / float64(len(p.fft.Data))
	result := make([]core.PxPoint, len(p.fft.Data))
	for i, d := range p.fft.Data {
		freq := p.fft.Range.From + core.Frequency(float64(i)*fftResolution)
		result[i] = core.PxPoint{
			X: resolution.ToPx(freq - frequencyRange.From),
			Y: core.Px((core.DB(d)-p.dbRange.From)/p.dbRange.Width()) * p.height,
		}
	}
	return result
}

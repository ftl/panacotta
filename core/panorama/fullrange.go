package panorama

import (
	"math"

	"github.com/ftl/panacotta/core"
)

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
		MeanLine:       0.0,
		Peaks:          []core.PeakMark{},
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

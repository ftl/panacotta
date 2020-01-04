package panorama

import (
	"math"

	"github.com/ftl/panacotta/core"
)

func (p Panorama) fullRangeData() core.Panorama {
	if p.fft.Range.Width() == 0 || p.width == 0 {
		return core.Panorama{}
	}

	frequencyRange := p.fft.Range
	result := core.Panorama{
		FrequencyRange: frequencyRange,
		VFO:            p.vfo,
		Band:           p.band,
		Resolution:     p.resolution[p.viewMode],

		VFOLine:        frequencyRange.ToFrct(p.vfo.Frequency),
		VFOFilterFrom:  frequencyRange.ToFrct(p.vfo.Frequency - p.vfo.FilterWidth/2),
		VFOFilterTo:    frequencyRange.ToFrct(p.vfo.Frequency + p.vfo.FilterWidth/2),
		VFOSignalLevel: p.signalLevel(),

		FrequencyScale:     p.fullRangeFrequencyScale(),
		DBScale:            p.dbScale(),
		Spectrum:           p.fullRangeSpectrum(),
		PeakThresholdLevel: 0.0,
		Peaks:              []core.PeakMark{},
	}

	return result
}

func (p Panorama) fullRangeFrequencyScale() []core.FrequencyMark {
	frequencyRange := p.fft.Range
	fZeros := float64(int(math.Log10(float64(frequencyRange.Width()))) - 1)
	fMagnitude := int(math.Pow(10, fZeros))
	fFactor := fMagnitude
	if fFactor < 0 {
		return []core.FrequencyMark{}
	}

	for core.Frequency(fFactor)/frequencyRange.Width() < 0.1 {
		if fFactor%10 == 0 {
			fFactor *= 5
		} else {
			fFactor *= 10
		}
	}
	for core.Frequency(fFactor)/frequencyRange.Width() > 0.15 {
		if fFactor%10 == 0 {
			fFactor /= 5
		} else {
			fFactor /= 10
		}
	}

	freqScale := make([]core.FrequencyMark, 0, int(frequencyRange.Width())/fFactor)
	for f := core.Frequency((int(frequencyRange.From) / fFactor) * fFactor); f < frequencyRange.To; f += core.Frequency(fFactor) {
		mark := core.FrequencyMark{
			X:         frequencyRange.ToFrct(f),
			Frequency: f,
		}
		freqScale = append(freqScale, mark)
	}

	return freqScale
}

func (p Panorama) fullRangeSpectrum() []core.FPoint {
	frequencyRange := p.fft.Range
	result := make([]core.FPoint, len(p.fft.Data))
	for i, d := range p.fft.Data {
		result[i] = core.FPoint{
			X: frequencyRange.ToFrct(p.fft.Frequency(i)),
			Y: p.dbRange.ToFrct(core.DB(d)),
		}
	}
	return result
}

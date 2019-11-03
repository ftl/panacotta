package core

import (
	"fmt"
)

// Frequency represents a frequency in Hz.
type Frequency float64

func (f Frequency) String() string {
	return fmt.Sprintf("%.2fHz", f)
}

// FrequencyRange represents a range of frequencies.
type FrequencyRange struct {
	From, To Frequency
}

func (r FrequencyRange) String() string {
	return fmt.Sprintf("[%v,%v]", r.From, r.To)
}

// Width of the frequency range.
func (r FrequencyRange) Width() Frequency {
	return r.To - r.From
}

// Contains the given frequency.
func (r FrequencyRange) Contains(f Frequency) bool {
	return f >= r.From && f <= r.To
}

// Shift the frequency by the given Δ.
func (r *FrequencyRange) Shift(Δ Frequency) {
	r.From += Δ
	r.To += Δ
}

// Configuration parameters of the application.
type Configuration struct {
	FrequencyCorrection int
	Testmode            bool
	VFOHost             string
	FFTPerSecond        int
}

// SamplesInput interface.
type SamplesInput interface {
	Samples() <-chan []byte
	Close() error
}

// Px unit for pixels
type Px float64

// PxPoint unit for pixel coordinates
type PxPoint struct {
	X, Y Px
}

// FrequencyMark on the frequency scale
type FrequencyMark struct {
	Frequency Frequency
	X         Px
}

// HzPerPx unit for resolution
type HzPerPx float64

// ToPx converts the given Frequency in Hz to Px
func (r HzPerPx) ToPx(f Frequency) Px {
	return Px(float64(f) / float64(r))
}

// ToHz converts the given Px to Hz
func (r HzPerPx) ToHz(p Px) Frequency {
	return Frequency(float64(p) * float64(r))
}

// Panorama current state
type Panorama struct {
	FrequencyRange FrequencyRange
	VFO            VFO
	Band           Band

	VFOLine        Px
	VFOFilterFrom  Px
	VFOFilterTo    Px
	FrequencyScale []FrequencyMark
	Spectrum       []PxPoint
}

// VFO current state
type VFO struct {
	Frequency   Frequency
	FilterWidth Frequency
	Mode        string
}

// FFT data and the corresponding frequency range
type FFT struct {
	Data  []float64
	Range FrequencyRange
}

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

// Center frequency of this range.
func (r FrequencyRange) Center() Frequency {
	return r.From + (r.To-r.From)/2
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

// Expanded returns a new expanded range.
func (r FrequencyRange) Expanded(Δ Frequency) FrequencyRange {
	return FrequencyRange{From: r.From - Δ, To: r.To + Δ}
}

// DB represents decibel (dB).
type DB float64

func (f DB) String() string {
	return fmt.Sprintf("%.2fdB", f)
}

// DBRange represents a range of dB.
type DBRange struct {
	From, To DB
}

func (r DBRange) String() string {
	return fmt.Sprintf("[%v,%v]", r.From, r.To)
}

// Width of the dB range.
func (r DBRange) Width() DB {
	return r.To - r.From
}

// Contains the given value in dB.
func (r DBRange) Contains(value DB) bool {
	return value >= r.From && value <= r.To
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
	Samples() <-chan []complex128
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

// DBMark on the dB scale
type DBMark struct {
	DB DB
	Y  Px
}

// HzPerPx unit for resolution
type HzPerPx float64

// ToPx converts the given Frequency in Hz to Px
func (r HzPerPx) ToPx(f Frequency) Px {
	return Px(float64(f) / float64(r))
}

// ToHz converts the given Px to Hz
func (r HzPerPx) ToHz(x Px) Frequency {
	return Frequency(float64(x) * float64(r))
}

// Panorama current state
type Panorama struct {
	FrequencyRange FrequencyRange
	VFO            VFO
	Band           Band
	Resolution     HzPerPx

	VFOLine        Px
	VFOFilterFrom  Px
	VFOFilterTo    Px
	FrequencyScale []FrequencyMark
	DBScale        []DBMark
	Spectrum       []PxPoint
	MeanLine       Px
	Peaks          []Px
}

// ToPx converts the given frequency in Hz to Px within the panorama.
func (p Panorama) ToPx(f Frequency) Px {
	return p.Resolution.ToPx(f - p.FrequencyRange.From)
}

// ToHz converts the given Px within the panorama to Hz.
func (p Panorama) ToHz(x Px) Frequency {
	return p.Resolution.ToHz(x) + p.FrequencyRange.From
}

// VFO current state
type VFO struct {
	Frequency   Frequency
	FilterWidth Frequency
	Mode        string
}

// FFT data and the corresponding frequency range
type FFT struct {
	Data          []float64
	Range         FrequencyRange
	Mean          float64
	PeakThreshold float64
	Peaks         []int
}

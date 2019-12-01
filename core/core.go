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

func (l DB) String() string {
	return fmt.Sprintf("%.2fdB", l)
}

func (l DB) ToSUnit() (s int, unit SUnit, add DB) {
	for i := len(SUnits) - 1; i >= 0; i-- {
		if l >= DB(SUnits[i]) {
			s = i
			unit = SUnits[i]
			add = l - DB(unit)
			return s, unit, add
		}
	}
	return 0, S0, l - DB(S0)
}

type SUnit DB

const (
	S0 SUnit = -127
	S1 SUnit = -121
	S2 SUnit = -115
	S3 SUnit = -109
	S4 SUnit = -103
	S5 SUnit = -97
	S6 SUnit = -91
	S7 SUnit = -85
	S8 SUnit = -79
	S9 SUnit = -73
)

var SUnits = []SUnit{S0, S1, S2, S3, S4, S5, S6, S7, S8, S9}

func (u SUnit) String() string {
	s, _, add := DB(u).ToSUnit()
	if s == 9 {
		return fmt.Sprintf("S%d+%.0fdB", s, add)
	} else if s > 0 {
		return fmt.Sprintf("S%d", s)
	} else {
		return fmt.Sprintf("S%d%.0fdB", s, add)
	}
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

// PeakMark contains all information to visualize a peak
type PeakMark struct {
	FromX        Px
	ToX          Px
	MaxX         Px
	MaxFrequency Frequency
	ValueY       Px
	ValueDB      DB
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
	Peaks          []PeakMark
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
	Name        string
	Frequency   Frequency
	FilterWidth Frequency
	Mode        string
	SignalLevel DB
}

// FFT data and the corresponding frequency range
type FFT struct {
	Data          []float64
	Range         FrequencyRange
	Mean          float64
	PeakThreshold float64
	Peaks         []PeakIndexRange
}

type PeakIndexRange struct {
	From  int
	To    int
	Max   int
	Value float64
}

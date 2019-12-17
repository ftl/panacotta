package core

import (
	"fmt"
	"math"
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

// ToFrct returns the fraction of the given frequency in this range.
func (r FrequencyRange) ToFrct(f Frequency) Frct {
	return Frct((f - r.From) / (r.To - r.From))
}

// ToFrequency converts the given fraction into a frequency from this range.
func (r FrequencyRange) ToFrequency(f Frct) Frequency {
	return r.From + (r.To-r.From)*Frequency(f)
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

func (r DBRange) Normalized() DBRange {
	if r.From > r.To {
		return DBRange{
			From: r.To,
			To:   r.From,
		}
	}
	return r
}

// Width of the dB range.
func (r DBRange) Width() DB {
	return DB(math.Abs(float64(r.To - r.From)))
}

// Contains the given value in dB.
func (r DBRange) Contains(value DB) bool {
	return value >= r.From && value <= r.To
}

// ToFrct returns the fraction of the given value in this range.
func (r DBRange) ToFrct(value DB) Frct {
	return Frct(float64((value - r.From) / r.Width()))
}

// ToDB converts the given fraction into a DB value in this range.
func (r DBRange) ToDB(f Frct) DB {
	return r.From + (r.To-r.From)*DB(f)
}

// Configuration parameters of the application.
type Configuration struct {
	FrequencyCorrection int
	Testmode            bool
	VFOHost             string
	FFTPerSecond        int
	DynamicRange        DBRange
}

// ViewMode of the panorama.
type ViewMode int

// All view modes.
const (
	ViewFixed ViewMode = iota
	ViewCentered
)

// SamplesInput interface.
type SamplesInput interface {
	Samples() <-chan []complex128
	Close() error
}

// Frct is a fraction of height or width; this is a abstraction of the coordinates on the screen.
type Frct float64

// FPoint represents a point on the screen using the Frct unit for its coordinates.
type FPoint struct {
	X, Y Frct
}

// FrequencyMark on the frequency scale
type FrequencyMark struct {
	Frequency Frequency
	X         Frct
}

// DBMark on the dB scale
type DBMark struct {
	DB DB
	Y  Frct
}

// PeakMark contains all information to visualize a peak
type PeakMark struct {
	FromX        Frct
	ToX          Frct
	MaxX         Frct
	MaxFrequency Frequency
	ValueY       Frct
	ValueDB      DB
}

// Px unit for pixels
type Px float64

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

	VFOLine        Frct
	VFOFilterFrom  Frct
	VFOFilterTo    Frct
	VFOSignalLevel DB

	FrequencyScale    []FrequencyMark
	DBScale           []DBMark
	Spectrum          []FPoint
	PeakThresholdLine Frct
	Peaks             []PeakMark
	Waterline         []Frct
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
}

// FFT data and the corresponding frequency range
type FFT struct {
	Data          []float64
	Range         FrequencyRange
	Mean          float64
	PeakThreshold float64
	Peaks         []PeakIndexRange
}

// Resolution of this FFT in Hz per Bin
func (fft FFT) Resolution() float64 {
	return float64(fft.Range.Width()) / float64(len(fft.Data))
}

// Frequency returns the center frequency of the ith bin of this FFT.
func (fft FFT) Frequency(i int) Frequency {
	return fft.Range.From + Frequency(float64(i)*fft.Resolution()+fft.Resolution()/2)
}

// ToIndex returns the index of the bin that the given frequency belongs to.
func (fft FFT) ToIndex(f Frequency) int {
	return int(float64(f-fft.Range.From) / fft.Resolution())
}

// PeakIndexRange contains the index values within FFT data that describe a peak.
type PeakIndexRange struct {
	From  int
	To    int
	Max   int
	Value float64
}

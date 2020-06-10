package core

import (
	"github.com/ftl/hamradio"
	"github.com/ftl/hamradio/bandplan"
)

// Frequeny alias for hamradio.Frequency
type Frequency = hamradio.Frequency

// FrequencyRange alias for hamradio.FrequencyRange
type FrequencyRange = hamradio.FrequencyRange

// DB alias for hamradio.DB
type DB = hamradio.DB

// SUnit alias for hamradio.SUnit
type SUnit = hamradio.SUnit

const (
	S0 SUnit = hamradio.S0
	S1 SUnit = hamradio.S1
	S2 SUnit = hamradio.S2
	S3 SUnit = hamradio.S3
	S4 SUnit = hamradio.S4
	S5 SUnit = hamradio.S5
	S6 SUnit = hamradio.S6
	S7 SUnit = hamradio.S7
	S8 SUnit = hamradio.S8
	S9 SUnit = hamradio.S9
)

// SUnits contains all S-units (S0-S9)
var SUnits = hamradio.SUnits

// DBRange alias for hamradio.DBRange
type DBRange = hamradio.DBRange

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

// ToFrequencyFrct returns the fraction of the given frequency in the given range.
func ToFrequencyFrct(f Frequency, r FrequencyRange) Frct {
	return Frct((f - r.From) / r.Width())
}

// FrctToFrequency converts the given fraction into a frequency from the given range.
func FrctToFrequency(f Frct, r FrequencyRange) Frequency {
	return r.From + (r.To-r.From)*Frequency(f)
}

// ToDBFrct returns the fraction of the given value in this range.
func ToDBFrct(value DB, r DBRange) Frct {
	return Frct((value - r.From) / r.Width())
}

// FrctToDB converts the given fraction into a DB value in the given range.
func FrctToDB(f Frct, r DBRange) DB {
	return r.From + (r.To-r.From)*DB(f)
}

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
	Band           bandplan.Band
	Resolution     HzPerPx

	VFOLine        Frct
	VFOFilterFrom  Frct
	VFOFilterTo    Frct
	VFOSignalLevel DB

	FrequencyScale     []FrequencyMark
	DBScale            []DBMark
	Spectrum           []FPoint
	PeakThresholdLevel Frct
	SigmaEnvelope      []FPoint
	Peaks              []PeakMark
	Waterline          []Frct
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
	SigmaEnvelope []float64
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

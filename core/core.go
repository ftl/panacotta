package core

import "fmt"

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

// HzPerPx unit for resolution
type HzPerPx float64

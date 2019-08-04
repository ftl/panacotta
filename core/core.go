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
func (r *FrequencyRange) Width() Frequency {
	return r.To - r.From
}

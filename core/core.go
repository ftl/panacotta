package core

// Frequency represents a frequency in Hz.
type Frequency float64

// FrequencyRange represents a range of frequencies.
type FrequencyRange struct {
	From, To Frequency
}

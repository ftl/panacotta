package dsp

import (
	"log"
	"math"
	"math/rand"
	"time"
)

// NewRandomInput returns a new SamplesInput that produces random samples.
func NewRandomInput(blockSize int, sampleRate int) *RandomInput {
	result := RandomInput{
		samples: make(chan []complex128, 1),
		done:    make(chan struct{}),
	}

	go func() {
		defer log.Print("RandomInput shutdown")
		for {
			nextBlock := make([]complex128, blockSize)
			for i := range nextBlock {
				nextBlock[i] = complex(rand.Float64(), rand.Float64())
			}
			select {
			case result.samples <- nextBlock:
				time.Sleep(time.Duration(float64(blockSize)/float64(sampleRate)*1000.0) * time.Millisecond)
			case <-result.done:
				close(result.samples)
				return
			}
		}
	}()

	return &result
}

type RandomInput struct {
	samples chan []complex128
	done    chan struct{}
}

func (i *RandomInput) Samples() <-chan []complex128 {
	return i.samples
}

func (i *RandomInput) Close() error {
	close(i.done)
	return nil
}

// NewToneInput returns a new SamplesInput that produces samples of a sine wave with the given frequency.
func NewToneInput(blockSize int, sampleRate int, f float64) *ToneInput {
	result := ToneInput{
		samples: make(chan []complex128, 1),
		done:    make(chan struct{}),
	}
	ratio := f / float64(sampleRate)
	ω := 2.0 * math.Pi * ratio

	go func() {
		defer log.Print("ToneInput shutdown")
		for {
			nextBlock := make([]complex128, blockSize)
			for i := range nextBlock {
				t := float64(i)
				nextBlock[i] = complex(math.Cos(ω*t), math.Sin(ω*t))
			}
			select {
			case result.samples <- nextBlock:
				time.Sleep(time.Duration(float64(blockSize)/float64(sampleRate)*1000.0) * time.Millisecond)
			case <-result.done:
				close(result.samples)
				return
			}
		}
	}()

	return &result
}

type ToneInput struct {
	samples chan []complex128
	done    chan struct{}
}

func (i *ToneInput) Samples() <-chan []complex128 {
	return i.samples
}

func (i *ToneInput) Close() error {
	close(i.done)
	return nil
}

// NewSweepInput returns a new SamplesInput that produces samples of a sine wave with the given frequency.
func NewSweepInput(blockSize int, sampleRate int, from, to, step float64) *SweepInput {
	result := SweepInput{
		samples: make(chan []complex128, 1),
		done:    make(chan struct{}),
	}
	go func() {
		defer log.Print("SweepInput shutdown")
		f := from
		for {
			ratio := f / float64(sampleRate)
			ω := 2.0 * math.Pi * ratio
			nextBlock := make([]complex128, blockSize)
			for i := range nextBlock {
				t := float64(i)
				nextBlock[i] = complex(math.Cos(ω*t), math.Sin(ω*t))
			}
			select {
			case result.samples <- nextBlock:
				time.Sleep(time.Duration(float64(blockSize)/float64(sampleRate)*1000.0) * time.Millisecond)
			case <-result.done:
				close(result.samples)
				return
			}
			f += step
			if f > to {
				f = from
			}
		}
	}()

	return &result
}

type SweepInput struct {
	samples chan []complex128
	done    chan struct{}
}

func (i *SweepInput) Samples() <-chan []complex128 {
	return i.samples
}

func (i *SweepInput) Close() error {
	close(i.done)
	return nil
}

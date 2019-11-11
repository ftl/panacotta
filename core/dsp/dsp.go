package dsp

import "github.com/ftl/panacotta/core"

func New(sampleRate int, ifFrequency, rxOffset core.Frequency) *DSP {
	result := DSP{
		cancel:    make(chan struct{}),
		workInput: make(chan work, 1),
		fft:       make(chan core.FFT, 1),

		sampleRate: sampleRate,
		ifCenter:   ifFrequency,
		rxCenter:   ifFrequency + rxOffset,
	}

	go result.run()

	return &result
}

type DSP struct {
	cancel    chan struct{}
	workInput chan work
	fft       chan core.FFT

	sampleRate int
	ifCenter   core.Frequency
	rxCenter   core.Frequency // actual receiving frequency

	vfo      core.VFO
	fftRange core.FrequencyRange
}

type work struct {
	samples  []byte
	fftRange core.FrequencyRange
	vfo      core.VFO
}

func (d *DSP) run() {
	for {
		select {
		case work := <-d.workInput:
			// TODO calc blocksize and effective range
			d.fftRange = work.fftRange
			d.vfo = work.vfo

			// TODO produce FFT
			d.fft <- core.FFT{}
		case <-d.cancel:
			return
		}
	}
}

func (d *DSP) Stop() {
	select {
	case <-d.cancel:
		return
	default:
		close(d.cancel)
	}
}

func (d *DSP) ProcessSamples(samples []byte, fftRange core.FrequencyRange, vfo core.VFO) {
	d.workInput <- work{samples, fftRange, vfo}
}

func (d *DSP) FFT() chan core.FFT {
	return d.fft
}

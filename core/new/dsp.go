package new

import "github.com/ftl/panacotta/core"

func NewDSP(sampleRate int, ifFrequency, rxOffset core.Frequency) *DSP {
	result := DSP{
		cancel:    make(chan struct{}),
		workInput: make(chan work, 1),
		vfoInput:  make(chan core.VFO, 1),
		FFT:       make(chan core.FFT, 1),

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
	vfoInput  chan core.VFO
	FFT       chan core.FFT

	sampleRate int
	ifCenter   core.Frequency
	rxCenter   core.Frequency // actual receiving frequency

	vfo      core.VFO
	fftRange core.FrequencyRange
}

type work struct {
	samples  []byte
	fftRange core.FrequencyRange
}

func (d *DSP) run() {
	for {
		select {
		case vfo := <-d.vfoInput:
			d.vfo = vfo
		case work := <-d.workInput:
			// TODO calc blocksize and effective range
			d.fftRange = work.fftRange

			// TODO produce FFT
			d.FFT <- core.FFT{}
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

func (d *DSP) SetVFO(vfo core.VFO) {
	d.vfoInput <- vfo
}

func (d *DSP) ProcessSamples(samples []byte, fftRange core.FrequencyRange) {
	d.workInput <- work{samples, fftRange}
}

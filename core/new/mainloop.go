package new

import "github.com/ftl/panacotta/core"

func NewMainLoop(samplesInput core.SamplesInput, dsp *DSP, panorama *Panorama) *MainLoop {
	return &MainLoop{
		cancel: make(chan struct{}),
		Done:   make(chan struct{}),

		samplesInput: samplesInput,
		dsp:          dsp,
		panorama:     panorama,
	}
}

type MainLoop struct {
	cancel chan struct{}
	Done   chan struct{}

	samplesInput core.SamplesInput
	dsp          *DSP
	panorama     *Panorama
}

func (m *MainLoop) Start() {
	go func() {
		for {
			select {
			case samples := <-m.samplesInput.Samples():
				m.dsp.ProcessSamples(samples, m.panorama.frequencyRange)
			case fft := <-m.dsp.FFT:
				m.panorama.SetFFT(fft)
			case <-m.cancel:
				close(m.Done)
				return
			}
		}
	}()
}

func (m *MainLoop) Stop() {
	select {
	case <-m.cancel:
		return
	default:
		close(m.cancel)
	}
}

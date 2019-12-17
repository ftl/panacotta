package app

import (
	"log"
	"math"
	"time"

	"github.com/ftl/panacotta/core"
)

func newMainLoop(samplesInput core.SamplesInput, dsp dspType, vfo vfoType, panorama panoramaType, fftPerSecond int) *mainLoop {
	redrawInterval := (1 * time.Second) / time.Duration(fftPerSecond)
	result := &mainLoop{
		samplesInput: samplesInput,
		dsp:          dsp,
		vfo:          vfo,
		panorama:     panorama,

		redrawInterval: redrawInterval,
		redrawTick:     time.NewTicker(redrawInterval),
		needFFTData:    true,
		command:        make(chan command, 1),

		panoramaData: make(chan core.Panorama, 1),
	}

	return result
}

type command func()

type mainLoop struct {
	samplesInput core.SamplesInput
	dsp          dspType
	vfo          vfoType
	tuner        tuner
	panorama     panoramaType

	redrawInterval time.Duration
	redrawTick     *time.Ticker
	needFFTData    bool
	command        chan command

	panoramaData chan core.Panorama
}

type dspType interface {
	ProcessSamples(samples []complex128, fftRange core.FrequencyRange, vfo core.VFO)
	FFT() chan core.FFT
}

type vfoType interface {
	Data() <-chan core.VFO
	TuneBy(Δf core.Frequency)
	TuneTo(f core.Frequency)
}

type panoramaType interface {
	VFO() (core.VFO, core.Band)
	FrequencyRange() core.FrequencyRange
	SetSize(core.Px, core.Px)
	SetFFT(core.FFT)
	SetVFO(core.VFO)
	Data() core.Panorama
	ToggleViewMode()
	ZoomIn()
	ZoomOut()
	ZoomToBand()
	ResetZoom()
	FinerDynamicRange()
	CoarserDynamicRange()
}

func (m *mainLoop) Run(stop chan struct{}) {
	defer log.Print("main loop shutdown")
	for {
		select {
		case samples := <-m.samplesInput.Samples():
			if !m.needFFTData {
				continue
			}

			vfo, _ := m.panorama.VFO()
			frequencyRange := m.panorama.FrequencyRange()
			m.dsp.ProcessSamples(samples, frequencyRange, vfo)
			m.needFFTData = false
		case fft := <-m.dsp.FFT():
			m.panorama.SetFFT(fft)
		case <-m.redrawTick.C:
			select {
			case m.panoramaData <- m.panorama.Data():
				m.needFFTData = true
			default:
				log.Print("trigger redraw hangs")
			}
		case vfo := <-m.vfo.Data():
			m.panorama.SetVFO(vfo)
		case command := <-m.command:
			command()
		case <-stop:
			m.redrawTick.Stop()
			return
		}
	}
}

// Panorama data for drawing
func (m *mainLoop) Panorama() <-chan core.Panorama {
	return m.panoramaData
}

func (m *mainLoop) q(cmd command) {
	select {
	case m.command <- cmd:
	default:
		log.Print("Mainloop.q hangs")
	}
}

// SetPanoramaWidth in Px
func (m *mainLoop) SetPanoramaSize(width, height core.Px) {
	m.q(func() {
		m.panorama.SetSize(width, height)
	})
}

// TuneTo the given frequency.
func (m *mainLoop) TuneTo(f core.Frequency) {
	m.q(func() {
		m.vfo.TuneTo(f)
	})
}

// TuneBy the given frequency.
func (m *mainLoop) TuneBy(Δf core.Frequency) {
	m.q(func() {
		m.vfo.TuneBy(Δf)
	})
}

// TuneUp the VFO.
func (m *mainLoop) TuneUp() {
	m.q(func() {
		m.vfo.TuneBy(m.tuner.dial())
	})
}

// TuneDown the VFO.
func (m *mainLoop) TuneDown() {
	m.q(func() {
		m.vfo.TuneBy(-m.tuner.dial())
	})
}

// ToggleViewMode of the panorama.
func (m *mainLoop) ToggleViewMode() {
	m.q(func() {
		m.panorama.ToggleViewMode()
	})
}

// ZoomIn on the panorama.
func (m *mainLoop) ZoomIn() {
	m.q(func() {
		m.panorama.ZoomIn()
	})
}

// ZoomOut of the panorama.
func (m *mainLoop) ZoomOut() {
	m.q(func() {
		m.panorama.ZoomOut()
	})
}

// ZoomToBand of the panorama.
func (m *mainLoop) ZoomToBand() {
	m.q(func() {
		m.panorama.ZoomToBand()
	})
}

// ResetZoom of the panorama.
func (m *mainLoop) ResetZoom() {
	m.q(func() {
		m.panorama.ResetZoom()
	})
}

func (m *mainLoop) FinerDynamicRange() {
	m.q(func() {
		m.panorama.FinerDynamicRange()
	})
}

func (m *mainLoop) CoarserDynamicRange() {
	m.q(func() {
		m.panorama.CoarserDynamicRange()
	})
}

type tuner struct {
	lastDial time.Time
}

func (t *tuner) dial() core.Frequency {
	now := time.Now()
	defer func() {
		t.lastDial = now
	}()
	rate := int(time.Second / now.Sub(t.lastDial))
	a := 0.1
	max := 500.0
	return core.Frequency((int(math.Min(math.Pow(a*float64(rate), 2), max))/10 + 1) * 10)
}

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

		panoramaData:   make(chan core.Panorama, 1),
		panoramaSize:   make(chan core.PxPoint, 1),
		tuneTo:         make(chan core.Frequency, 1),
		tuneBy:         make(chan core.Frequency, 1),
		toggleViewMode: make(chan struct{}, 1),
		zoomIn:         make(chan struct{}, 1),
		zoomOut:        make(chan struct{}, 1),
		resetZoom:      make(chan struct{}, 1),
	}

	return result
}

// MainLoop coordinates the work of all the components.
type mainLoop struct {
	samplesInput core.SamplesInput
	dsp          dspType
	vfo          vfoType
	tuner        tuner
	panorama     panoramaType

	redrawInterval time.Duration
	redrawTick     *time.Ticker
	needFFTData    bool

	panoramaData   chan core.Panorama
	panoramaSize   chan core.PxPoint
	tuneTo         chan core.Frequency
	tuneBy         chan core.Frequency
	toggleViewMode chan struct{}
	zoomIn         chan struct{}
	zoomOut        chan struct{}
	resetZoom      chan struct{}
}

// DSP from the main loop's perspective.
type dspType interface {
	ProcessSamples(samples []complex128, fftRange core.FrequencyRange, vfo core.VFO)
	FFT() chan core.FFT
}

// VFO from the main loop's perspective.
type vfoType interface {
	Data() <-chan core.VFO
	TuneBy(Δf core.Frequency)
	TuneTo(f core.Frequency)
}

// Panorama from the main loop's perspective.
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
	ResetZoom()
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
		case size := <-m.panoramaSize:
			m.panorama.SetSize(size.X, size.Y)
		case f := <-m.tuneTo:
			m.vfo.TuneTo(f)
		case Δf := <-m.tuneBy:
			m.vfo.TuneBy(Δf)
		case <-m.toggleViewMode:
			m.panorama.ToggleViewMode()
		case <-m.zoomIn:
			m.panorama.ZoomIn()
		case <-m.zoomOut:
			m.panorama.ZoomOut()
		case <-m.resetZoom:
			m.panorama.ResetZoom()
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

// SetPanoramaWidth in Px
func (m *mainLoop) SetPanoramaSize(width, height core.Px) {
	select {
	case m.panoramaSize <- core.PxPoint{width, height}:
	default:
		log.Print("SetPanoramaSize hangs")
	}
}

// TuneTo the given frequency.
func (m *mainLoop) TuneTo(f core.Frequency) {
	select {
	case m.tuneTo <- f:
	default:
		log.Print("TuneTo hangs")
	}
}

// TuneBy the given frequency.
func (m *mainLoop) TuneBy(Δf core.Frequency) {
	select {
	case m.tuneBy <- Δf:
	default:
		log.Print("TuneBy hangs")
	}
}

// TuneUp the VFO.
func (m *mainLoop) TuneUp() {
	select {
	case m.tuneBy <- m.tuner.dial():
	default:
		log.Print("TuneUp hangs")
	}
}

// TuneDown the VFO.
func (m *mainLoop) TuneDown() {
	select {
	case m.tuneBy <- -m.tuner.dial():
	default:
		log.Print("TuneDown hangs")
	}
}

// ToggleViewMode of the panorama.
func (m *mainLoop) ToggleViewMode() {
	select {
	case m.toggleViewMode <- struct{}{}:
	default:
		log.Print("ToggleViewMode hangs")
	}
}

// ZoomIn on the panorama.
func (m *mainLoop) ZoomIn() {
	select {
	case m.zoomIn <- struct{}{}:
	default:
		log.Print("ZoomIn hangs")
	}
}

// ZoomOut of the panorama.
func (m *mainLoop) ZoomOut() {
	select {
	case m.zoomOut <- struct{}{}:
	default:
		log.Print("ZoomOut hangs")
	}
}

// ResetZoom of the panorama.
func (m *mainLoop) ResetZoom() {
	select {
	case m.resetZoom <- struct{}{}:
	default:
		log.Print("ResetZoom hangs")
	}
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
	a := 0.3
	max := 500.0
	return core.Frequency((int(math.Min(math.Pow(a*float64(rate), 2), max))/10 + 1) * 10)
}

package new

import (
	"math"
	"time"

	"github.com/ftl/panacotta/core"
)

// NewMainLoop instance
func NewMainLoop(samplesInput core.SamplesInput, dsp *DSP, vfo VFO, panorama *Panorama) *MainLoop {
	return &MainLoop{
		cancel: make(chan struct{}),
		Done:   make(chan struct{}),

		samplesInput: samplesInput,
		dsp:          dsp,
		vfo:          vfo,
		panorama:     panorama,

		panoramaData:   make(chan core.Panorama, 1),
		panoramaWidth:  make(chan core.Px, 1),
		tuneTo:         make(chan core.Frequency, 1),
		tuneBy:         make(chan core.Frequency, 1),
		toggleViewMode: make(chan struct{}, 1),
		zoomIn:         make(chan struct{}, 1),
		zoomOut:        make(chan struct{}, 1),
		resetZoom:      make(chan struct{}, 1),
	}
}

// MainLoop coordinates the work of all the components.
type MainLoop struct {
	cancel chan struct{}
	Done   chan struct{}

	samplesInput core.SamplesInput
	dsp          *DSP
	vfo          VFO
	tuner        tuner
	panorama     *Panorama

	panoramaData   chan core.Panorama
	panoramaWidth  chan core.Px
	tuneTo         chan core.Frequency
	tuneBy         chan core.Frequency
	toggleViewMode chan struct{}
	zoomIn         chan struct{}
	zoomOut        chan struct{}
	resetZoom      chan struct{}
}

// VFO from the main loop's perspective.
type VFO interface {
	Data() <-chan core.VFO
	TuneBy(Δf core.Frequency)
	TuneTo(f core.Frequency)
}

func (m *MainLoop) run() {
	for {
		select {
		case samples := <-m.samplesInput.Samples():
			m.dsp.ProcessSamples(samples, m.panorama.frequencyRange)
		case fft := <-m.dsp.FFT:
			m.panorama.SetFFT(fft)
			m.panoramaData <- m.panorama.Data() // TODO do this with a timer, based on the desired frame rate
		case vfo := <-m.vfo.Data():
			m.panorama.SetVFO(vfo)
		case width := <-m.panoramaWidth:
			m.panorama.SetWidth(width)
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
		case <-m.cancel:
			close(m.Done)
			return
		}
	}
}

// Start the main loop
func (m *MainLoop) Start() {
	go m.run()
}

// Stop the main loop
func (m *MainLoop) Stop() {
	select {
	case <-m.cancel:
		return
	default:
		close(m.cancel)
	}
}

// Panorama data for drawing
func (m *MainLoop) Panorama() <-chan core.Panorama {
	return m.panoramaData
}

// SetPanoramaWidth in Px
func (m *MainLoop) SetPanoramaWidth(width core.Px) {
	m.panoramaWidth <- width
}

// TuneTo the given frequency.
func (m *MainLoop) TuneTo(f core.Frequency) {
	m.tuneTo <- f
}

// TuneUp the VFO.
func (m *MainLoop) TuneUp() {
	m.tuneBy <- m.tuner.dial()
}

// TuneDown the VFO.
func (m *MainLoop) TuneDown() {
	m.tuneBy <- -m.tuner.dial()
}

// ToggleViewMode of the panorama.
func (m *MainLoop) ToggleViewMode() {
	m.toggleViewMode <- struct{}{}
}

// ZoomIn on the panorama.
func (m *MainLoop) ZoomIn() {
	m.zoomIn <- struct{}{}
}

// ZoomOut of the panorama.
func (m *MainLoop) ZoomOut() {
	m.zoomOut <- struct{}{}
}

// ResetZoom of the panorama.
func (m *MainLoop) ResetZoom() {
	m.resetZoom <- struct{}{}
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

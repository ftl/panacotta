package app

import (
	"log"
	"sync"

	"github.com/ftl/panacotta/core"
	"github.com/ftl/panacotta/core/bandplan"
	"github.com/ftl/panacotta/core/rtlsdr"
	"github.com/ftl/panacotta/core/rx"
	"github.com/ftl/panacotta/core/vfo"
)

// NewController returns a new instance of the AppController interface.
func NewController() *Controller {
	return &Controller{}
}

// PanoramaView shows FFT data, the VFO ROI, the VFO frequency.
type PanoramaView interface {
	SetFFTData([]float64)
	SetVFO(frequency core.Frequency, band bandplan.Band, roi core.FrequencyRange)
}

// Controller for the application.
type Controller struct {
	done         chan struct{}
	subProcesses *sync.WaitGroup

	rx  *rx.Receiver
	vfo *vfo.VFO

	panorama PanoramaView
}

// Startup the application.
func (c *Controller) Startup() {
	c.done = make(chan struct{})
	c.subProcesses = new(sync.WaitGroup)

	ifCenter := 67899000   // this is fix for the FT-450D
	rxBandwidth := 1800000 // this is the sample rate
	rxCenter := ifCenter + (rxBandwidth / 4)
	log.Printf("rx @ %v", rxCenter)

	dongle, err := rtlsdr.Open(rxCenter, rxBandwidth, -50)
	if err != nil {
		log.Fatal(err)
	}
	c.rx = rx.New(dongle, core.Frequency(ifCenter), core.Frequency(rxCenter), core.Frequency(rxBandwidth))
	c.rx.OnFFTAvailable(c.panorama.SetFFTData)
	c.rx.OnVFOChange(c.panorama.SetVFO)

	c.vfo, err = vfo.Open("afu.fritz.box:4532")
	if err != nil {
		log.Fatal(err)
	}
	c.vfo.OnFrequencyChange(func(f core.Frequency) {
		log.Print("Current frequency: ", f)
	})
	c.vfo.OnFrequencyChange(c.rx.SetVFOFrequency)

	c.rx.Run(c.done, c.subProcesses)
	c.vfo.Run(c.done, c.subProcesses)
}

// Shutdown the application.
func (c *Controller) Shutdown() {
	close(c.done)
	c.subProcesses.Wait()
}

// SetPanoramaView sets the panorama view.
func (c *Controller) SetPanoramaView(view PanoramaView) {
	c.panorama = view
}

// Tune the VFO to the given frequency.
func (c *Controller) Tune(f core.Frequency) {
	c.vfo.SetFrequency(core.Frequency(int(f/10) * 10))
}

// FineTuneUp moves the VFO frequency 10Hz upwards.
func (c *Controller) FineTuneUp() {
	log.Print("fine tune up")
	c.vfo.MoveFrequency(10)
}

// FineTuneDown moves the VFO frequency 10Hz downwards.
func (c *Controller) FineTuneDown() {
	log.Print("fine tune down")
	c.vfo.MoveFrequency(-10)
}

// ToggleViewMode of the panorama view.
func (c *Controller) ToggleViewMode() {
	switch c.rx.ViewMode() {
	case rx.ViewFullBand:
		c.rx.SetViewMode(rx.ViewCentered)
	case rx.ViewCentered:
		c.rx.SetViewMode(rx.ViewFullBand)
	}
}

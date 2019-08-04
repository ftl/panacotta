package app

import (
	"log"
	"sync"

	"github.com/ftl/panacotta/core"
	"github.com/ftl/panacotta/core/rtlsdr"
	"github.com/ftl/panacotta/core/rx"
	"github.com/ftl/panacotta/core/vfo"
)

// NewController returns a new instance of the AppController interface.
func NewController() *Controller {
	return &Controller{}
}

// View implementation used by this controller.
type View interface {
}

// PanoramaView shows FFT data, the VFO ROI, the VFO frequency.
type PanoramaView interface {
	SetFFTData([]float64)
	SetVFO(frequency core.Frequency, roi core.FrequencyRange)
}

// Controller for the application.
type Controller struct {
	done         chan struct{}
	subProcesses *sync.WaitGroup

	panorama PanoramaView
}

// Startup the application.
func (c *Controller) Startup() {
	c.done = make(chan struct{})
	c.subProcesses = new(sync.WaitGroup)

	ifCenter := 67899000   // this is fix for the FT-450D
	rxBandwidth := 1800000 // this is the sample rate
	rxCenter := ifCenter - (rxBandwidth / 4)

	dongle, err := rtlsdr.Open(rxCenter, rxBandwidth, -50)
	if err != nil {
		log.Fatal(err)
	}
	rx := rx.New(dongle, core.Frequency(ifCenter), core.Frequency(rxCenter), core.Frequency(rxBandwidth))
	rx.OnFFTAvailable(c.panorama.SetFFTData)
	rx.OnVFOChange(c.panorama.SetVFO)

	vfo, err := vfo.Open("afu.fritz.box:4532")
	if err != nil {
		log.Fatal(err)
	}
	vfo.OnFrequencyChange(func(f core.Frequency) {
		log.Print("Current frequency: ", f)
	})
	vfo.OnFrequencyChange(rx.SetVFOFrequency)

	rx.Run(c.done, c.subProcesses)
	vfo.Run(c.done, c.subProcesses)
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

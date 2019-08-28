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
func NewController(config core.Configuration) *Controller {
	return &Controller{
		config: config,
	}
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

	config core.Configuration
	rx     *rx.Receiver
	vfo    *vfo.VFO

	panorama PanoramaView
}

// Startup the application.
func (c *Controller) Startup() {
	c.done = make(chan struct{})
	c.subProcesses = new(sync.WaitGroup)

	// configuration
	ifCenter := 67899000   // this is fix for the FT-450D and specific to our method
	rxBandwidth := 1800000 // this is the sample rate and specific to our method
	blockSize := 32768     // 131072    // this is the number of *complex* samples in one block

	rxCenter := ifCenter + (rxBandwidth / 4)
	log.Printf("RX @ %v %d ppm", rxCenter, c.config.FrequencyCorrection)

	samplesInput, err := c.openSamplesInput(rxCenter, rxBandwidth, blockSize, c.config.FrequencyCorrection, c.config.Testmode)
	if err != nil {
		log.Fatal(err)
	}

	c.rx = rx.New(samplesInput, blockSize, core.Frequency(ifCenter), core.Frequency(rxCenter), core.Frequency(rxBandwidth))
	c.rx.OnFFTAvailable(c.panorama.SetFFTData)
	c.rx.OnVFOChange(c.panorama.SetVFO)

	c.vfo, err = vfo.Open(c.config.VFOHost)
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

func (c *Controller) openSamplesInput(centerFrequency int, sampleRate int, blockSize int, frequencyCorrection int, testmode bool) (core.SamplesInput, error) {
	// if testmode {
	// 	return new(rx.RandomReader), nil
	// }
	return rtlsdr.Open(centerFrequency, sampleRate, blockSize, frequencyCorrection)
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

package app

import (
	"log"

	"github.com/ftl/panacotta/core"
	"github.com/ftl/panacotta/core/dsp"
	"github.com/ftl/panacotta/core/panorama"
	"github.com/ftl/panacotta/core/rtlsdr"
	"github.com/ftl/panacotta/core/rx"
	"github.com/ftl/panacotta/core/vfo"
)

// New returns a new app Controller.
func New(config core.Configuration) *Controller {
	return &Controller{
		config: config,
		stop:   make(chan struct{}),
	}
}

// Controller for the application.
type Controller struct {
	*mainLoop
	stop chan struct{}

	config        core.Configuration
	fullRangeMode bool
}

// Startup the application.
func (c *Controller) Startup() {
	// configuration
	ifCenter := 67899000  // this is fix for the FT-450D and specific to our method
	sampleRate := 1800000 // 2097152 // 1800000 // this is specific to our method
	blockSize := 32768    // 131072    // this is the number of *complex* samples in one block
	c.fullRangeMode = false

	rxCenter := ifCenter - (sampleRate / 4)
	log.Printf("RX @ %v %d ppm", rxCenter, c.config.FrequencyCorrection)
	log.Printf("FFT per second: %d", c.config.FFTPerSecond)

	samplesInput, err := c.openSamplesInput(rxCenter, sampleRate, blockSize, c.config.FrequencyCorrection, c.config.Testmode)
	if err != nil {
		log.Fatal(err)
	}

	vfo, err := vfo.Open(c.config.VFOHost)
	if err != nil {
		log.Fatal(err)
	}
	go vfo.Run(c.stop)

	var (
		d *dsp.DSP
		p *panorama.Panorama
	)
	if c.fullRangeMode {
		d = dsp.NewFullRange(sampleRate, core.Frequency(ifCenter), core.Frequency(-sampleRate/4))
		p = panorama.NewFullSpectrum(0, core.FrequencyRange{}, 0)
	} else {
		d = dsp.New(sampleRate, core.Frequency(ifCenter), core.Frequency(-sampleRate/4))
		p = panorama.New(0, core.FrequencyRange{}, 0)

	}
	go d.Run(c.stop)

	c.mainLoop = newMainLoop(samplesInput, d, vfo, p)
	go c.mainLoop.Run(c.stop)
}

func (c *Controller) openSamplesInput(centerFrequency int, sampleRate int, blockSize int, frequencyCorrection int, testmode bool) (core.SamplesInput, error) {
	if testmode {
		log.Printf("Testmode, using random samples input")
		return rx.NewRandomInput(blockSize, sampleRate), nil
		// return rx.NewToneInput(blockSize, sampleRate, 460000.0), nil
		// return rx.NewSweepInput(blockSize, sampleRate, -float64(sampleRate/2), float64(sampleRate/2), float64(sampleRate)*0.001), nil
		// return rx.NewSweepInput(blockSize, sampleRate, 0, float64(sampleRate), float64(sampleRate)*0.001), nil
	}
	return rtlsdr.Open(centerFrequency, sampleRate, blockSize, frequencyCorrection)
}

// Shutdown the application.
func (c *Controller) Shutdown() {
	close(c.stop)
}

// Done indicate done.
func (c *Controller) Done() chan struct{} {
	return c.stop
}

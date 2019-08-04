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

// FFTView for FFT data
type FFTView interface {
	ShowData([]float64)
}

// Controller for the application.
type Controller struct {
	done         chan struct{}
	subProcesses *sync.WaitGroup

	fftView FFTView
}

// Startup the application.
func (c *Controller) Startup() {
	c.done = make(chan struct{})
	c.subProcesses = new(sync.WaitGroup)

	dongle, err := rtlsdr.Open(67899000, 1800000, -50)
	if err != nil {
		log.Fatal(err)
	}
	rx := rx.New(dongle, c.fftView.ShowData)

	vfo, err := vfo.Open("afu.fritz.box:4532")
	if err != nil {
		log.Fatal(err)
	}
	vfo.OnFrequencyChange(func(f core.Frequency) {
		log.Print("Current frequency: ", f)
	})

	rx.Run(c.done, c.subProcesses)
	vfo.Run(c.done, c.subProcesses)
}

// Shutdown the application.
func (c *Controller) Shutdown() {
	close(c.done)
	c.subProcesses.Wait()
}

// SetFFTView sets the FFT view.
func (c *Controller) SetFFTView(fftView FFTView) {
	c.fftView = fftView
}

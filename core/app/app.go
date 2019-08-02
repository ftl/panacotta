package app

import (
	"log"
	"sync"

	"github.com/ftl/panacotta/core/rtlsdr"
	"github.com/ftl/panacotta/core/rx"
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
	ShowData([]complex128)
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
	rx.Run(c.done, c.subProcesses)
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

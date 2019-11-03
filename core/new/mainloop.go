package new

import "github.com/ftl/panacotta/core"

func NewMainLoop(samplesInput core.SamplesInput) *MainLoop {
	return &MainLoop{
		cancel: make(chan struct{}),
		Done:   make(chan struct{}),

		samples: samplesInput,
	}
}

type MainLoop struct {
	cancel chan struct{}
	Done   chan struct{}

	samples core.SamplesInput
}

func (m *MainLoop) Start() {
	go func() {
		for {
			select {
			case block := <-m.samples.Samples():
				_ = block
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

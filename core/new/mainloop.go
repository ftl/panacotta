package new

func NewMainLoop() *MainLoop {
	return &MainLoop{
		cancel: make(chan struct{}),
		done:   make(chan struct{}),

		samples: make(chan []float64, 1),
	}
}

type MainLoop struct {
	cancel chan struct{}
	done   chan struct{}

	samples chan []float64
}

func (m *MainLoop) Start() {
	go func() {
		for {
			select {
			case <-m.cancel:
				close(m.done)
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

func (m MainLoop) Done() chan struct{} {
	return m.done
}

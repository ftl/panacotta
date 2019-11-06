package new

import (
	"testing"
	"time"

	"github.com/ftl/panacotta/core"
	"github.com/stretchr/testify/assert"
)

func TestStopAndDone(t *testing.T) {
	m := NewMainLoop(&mockInput{}, &DSP{}, &mockVFO{}, &Panorama{})

	m.Start()
	start := time.Now()
	go func() {
		time.Sleep(100 * time.Millisecond)
		m.Stop()
	}()
	<-m.Done
	duration := time.Since(start)

	assert.True(t, duration > 100*time.Millisecond)
}

type mockInput struct{}

func (m *mockInput) Samples() <-chan []byte {
	return make(chan []byte)
}

func (m *mockInput) Close() error {
	return nil
}

type mockVFO struct{}

func (m *mockVFO) Data() <-chan core.VFO {
	return make(chan core.VFO)
}

func (m *mockVFO) TuneBy(Δf core.Frequency) {}

func (m *mockVFO) TuneTo(f core.Frequency) {}

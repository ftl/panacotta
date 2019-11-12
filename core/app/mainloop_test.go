package app

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/ftl/panacotta/core"
)

func TestStopAndDone(t *testing.T) {
	m := newMainLoop(&mockInput{}, &mockDSP{}, &mockVFO{}, &mockPanorama{})

	stop := make(chan struct{})
	start := time.Now()
	go func() {
		time.Sleep(100 * time.Millisecond)
		close(stop)
	}()
	m.Run(stop)
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

type mockDSP struct{}

func (m *mockDSP) ProcessSamples(samples []byte, fftRange core.FrequencyRange, vfo core.VFO) {}

func (m *mockDSP) FFT() chan core.FFT {
	return make(chan core.FFT)
}

type mockPanorama struct{}

func (m *mockPanorama) VFO() (core.VFO, core.Band) {
	return core.VFO{}, core.UnknownBand
}

func (m *mockPanorama) FrequencyRange() core.FrequencyRange {
	return core.FrequencyRange{}
}

func (m *mockPanorama) SetWidth(core.Px) {}

func (m *mockPanorama) SetFFT(core.FFT) {}

func (m *mockPanorama) SetVFO(core.VFO) {}

func (m *mockPanorama) Data() core.Panorama {
	return core.Panorama{}
}

func (m *mockPanorama) TuneTo(core.Frequency) {}

func (m *mockPanorama) TuneBy(core.Frequency) {}

func (m *mockPanorama) ToggleViewMode() {}

func (m *mockPanorama) ZoomIn() {}

func (m *mockPanorama) ZoomOut() {}

func (m *mockPanorama) ResetZoom() {}

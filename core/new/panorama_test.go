package new

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ftl/panacotta/core"
)

func TestWidth(t *testing.T) {
	p := NewPanorama(100, core.FrequencyRange{1000.0, 1200.0}, 1100.0)

	p.SetWidth(200)

	assert.Equal(t, core.Px(200), p.width)
	assert.Equal(t, core.Frequency(900.0), p.From())
	assert.Equal(t, core.Frequency(1300.0), p.To())

	p.SetWidth(50)

	assert.Equal(t, core.Px(50), p.width)
	assert.Equal(t, core.Frequency(1050.0), p.From())
	assert.Equal(t, core.Frequency(1150.0), p.To())

	p.SetVFO(core.VFO{1130.0, 10.0, ""})
	p.SetWidth(100)

	assert.Equal(t, core.Px(100), p.width)
	assert.Equal(t, core.Frequency(970.0), p.From())
	assert.Equal(t, core.Frequency(1170.0), p.To())
}

func TestToggleViewMode(t *testing.T) {
	p := NewPanorama(100, core.FrequencyRange{1000.0, 1200.0}, 1100.0)
	p.resolution[ViewCentered] = 1

	p.SetVFO(core.VFO{1150.0, 10.0, ""})

	assert.Equal(t, core.Frequency(1000.0), p.From())
	assert.Equal(t, core.Frequency(1200.0), p.To())

	p.ToggleViewMode()

	assert.Equal(t, core.Frequency(1100.0), p.From())
	assert.Equal(t, core.Frequency(1200.0), p.To())

	p.ToggleViewMode()

	assert.Equal(t, core.Frequency(1050.0), p.From())
	assert.Equal(t, core.Frequency(1250.0), p.To())
}

func TestCenteredVFO(t *testing.T) {
	p := NewPanorama(100, core.FrequencyRange{1000.0, 1200.0}, 1100.0)
	p.resolution[ViewCentered] = 1
	p.viewMode = ViewCentered

	p.SetVFO(core.VFO{1150.0, 10.0, ""})

	assert.Equal(t, core.Frequency(1100.0), p.From())
	assert.Equal(t, core.Frequency(1200.0), p.To())
}

func TestFixedVFO(t *testing.T) {
	p := NewPanorama(100, core.FrequencyRange{1000.0, 1200.0}, 1100.0)
	p.viewMode = ViewFixed

	p.SetVFO(core.VFO{1150.0, 10.0, ""})

	assert.Equal(t, core.Frequency(1000.0), p.From())
	assert.Equal(t, core.Frequency(1200.0), p.To())

	p.SetVFO(core.VFO{1190.0, 10.0, ""})

	assert.Equal(t, core.Frequency(1010.0), p.From())
	assert.Equal(t, core.Frequency(1210.0), p.To())

	p.SetVFO(core.VFO{2000.0, 10.0, ""})

	assert.Equal(t, core.Frequency(1010.0), p.From())
	assert.Equal(t, core.Frequency(1210.0), p.To())
}

func TestZoom(t *testing.T) {
	p := NewPanorama(1000, core.FrequencyRange{100000.0, 120000.0}, 110000.0)

	p.ZoomIn()

	assert.Equal(t, core.Frequency(102000.0), p.From())
	assert.Equal(t, core.Frequency(118000.0), p.To())

	p.ZoomOut()

	assert.Equal(t, core.Frequency(100000.0), p.From())
	assert.Equal(t, core.Frequency(120000.0), p.To())

	p.viewMode = ViewCentered
	p.ZoomTo(core.FrequencyRange{110000.0, 115000.0})

	assert.Equal(t, core.Frequency(110000.0), p.From())
	assert.Equal(t, core.Frequency(115000.0), p.To())
	assert.Equal(t, ViewFixed, p.viewMode)
	assert.Equal(t, core.HzPerPx(5.0), p.resolution[p.viewMode])

	p.ResetZoom()
	assert.Equal(t, core.Frequency(100000.0), p.From())
	assert.Equal(t, core.Frequency(200000.0), p.To())
	assert.Equal(t, defaultFixedResolution, p.resolution[p.viewMode])
}

func TestDrag(t *testing.T) {
	p := NewPanorama(1000, core.FrequencyRange{100000.0, 120000.0}, 110000.0)

	p.Drag(-500)

	assert.Equal(t, core.Frequency(90000.0), p.From())
	assert.Equal(t, core.Frequency(110000.0), p.To())

	p.Drag(500)

	assert.Equal(t, core.Frequency(100000.0), p.From())
	assert.Equal(t, core.Frequency(120000.0), p.To())
}

func TestFrequencyAt(t *testing.T) {
	p := NewPanorama(1000, core.FrequencyRange{100000.0, 120000.0}, 110000.0)

	assert.Equal(t, core.Frequency(105000.0), p.FrequencyAt(250))
	assert.Equal(t, core.Frequency(110000.0), p.FrequencyAt(500))
	assert.Equal(t, core.Frequency(115000.0), p.FrequencyAt(750))
}

func TestFrequencyScale(t *testing.T) {
	p := NewPanorama(1000, core.FrequencyRange{100300.0, 120700.0}, 110000.0)

	scale1 := p.frequencyScale()
	offset1 := int(scale1[1].X - scale1[0].X)

	assert.Equal(t, 5, len(scale1))
	assert.True(t, offset1 > 200)
	assert.True(t, scale1[1].X-scale1[0].X < 300)

	p.SetWidth(2000)
	scale2 := p.frequencyScale()
	offset2 := int(scale2[1].X - scale2[0].X)

	assert.Equal(t, 9, len(scale2))
	assert.Equal(t, offset1, offset2)
}

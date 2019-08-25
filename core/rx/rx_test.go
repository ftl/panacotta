package rx

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ftl/panacotta/core"
)

func TestRXVFOFrequencyConversion(t *testing.T) {
	f := core.Frequency(7100000)
	rx := New(nil, 67899000, 68349000, 1800000)
	rx.vfoFrequency = 7070000

	actual := rx.vfoToRx(rx.rxToVFO(f))

	assert.Equal(t, f, actual)
}

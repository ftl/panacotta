package new

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/ftl/panacotta/core/rx"
)

func TestStopAndDone(t *testing.T) {
	m := NewMainLoop(rx.NewRandomInput(1)) // TODO use a mock input

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

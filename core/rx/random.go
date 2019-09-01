package rx

import (
	"math/rand"
	"time"
)

// NewRandomInput returns a new SamplesInput that produces random samples.
func NewRandomInput(blockSize int) *RandomInput {
	result := RandomInput{
		bytes: make(chan []byte, 1),
		done:  make(chan struct{}),
	}

	go func() {
		for {
			nextBlock := make([]byte, blockSize*2)
			rand.Read(nextBlock)
			select {
			case result.bytes <- nextBlock:
				time.Sleep(10 * time.Millisecond)
			case <-result.done:
				close(result.bytes)
				return
			}
		}
	}()

	return &result
}

type RandomInput struct {
	bytes chan []byte
	done  chan struct{}
}

func (i *RandomInput) Samples() <-chan []byte {
	return i.bytes
}

func (i *RandomInput) Close() error {
	close(i.done)
	return nil
}

package rx

import (
	"math/rand"
)

// RandomReader returns random values.
type RandomReader struct{}

func (r *RandomReader) Read(p []byte) (n int, err error) {
	return rand.Read(p)
}

// Close the reader.
func (r *RandomReader) Close() error {
	return nil
}

// NullReader returns 0.
type NullReader struct{}

func (r *NullReader) Read(p []byte) (n int, err error) {
	for i := range p {
		p[i] = 0
	}
	return len(p), nil
}

// Close the reader.
func (r *NullReader) Close() error {
	return nil
}

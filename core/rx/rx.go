package rx

import (
	"io"
	"log"
	"math"
	"sync"
	"time"

	"github.com/pkg/errors"
	"gonum.org/v1/gonum/fourier"
)

// New instance of the receiver.
func New(in io.ReadCloser, fftSink FFTSink) *Receiver {
	result := Receiver{
		in:      in,
		fftSink: fftSink,
	}
	result.readBlock = result.readIQBlock8
	return &result
}

// Receiver type
type Receiver struct {
	in        io.ReadCloser
	readBlock blockReader
	fftSink   FFTSink
}

// FFTSink receives FFT data from the receiver.
type FFTSink func([]complex128)
type blockReader func(blocksize int) ([]complex128, error)

// Run this receiver.
func (r *Receiver) Run(stop chan struct{}, wait *sync.WaitGroup) {
	const blockSize = 131072

	wait.Add(1)
	go func() {
		defer wait.Done()
		defer r.in.Close()

		fa := fourier.NewCmplxFFT(blockSize)
		for {
			select {
			case <-time.After(1 * time.Millisecond):
				start := time.Now()
				block, err := r.readBlock(blockSize)
				log.Printf("data received after %v", time.Now().Sub(start))
				if err == io.EOF {
					log.Print("Waiting for data")
					continue
				} else if err != nil {
					log.Print("ERROR", err)
					continue
				}

				fftdata := fa.Coefficients(nil, block)
				r.fftSink(fftdata)
			case <-stop:
				return
			}
		}
	}()
}

func (r *Receiver) readIQBlock8(blocksize int) ([]complex128, error) {
	if blocksize%2 != 0 {
		return []complex128{}, errors.New("blocksize must be even")
	}

	result := make([]complex128, blocksize)

	buf := make([]byte, blocksize*2)
	_, err := r.in.Read(buf)
	if err != nil {
		return []complex128{}, errors.Wrap(err, "cannot read block of 8-bit samples")
	}

	for i := 0; i < len(buf); i += 2 {
		qSample := normalizeSampleUint8(buf[i])
		iSample := normalizeSampleUint8(buf[i+1])
		result[i/2] = complex(iSample, qSample)
	}

	return result, nil
}

func normalizeSampleUint8(s byte) float64 {
	return (float64(s) - float64(math.MaxInt8)) / float64(math.MaxInt8)
}

func (r *Receiver) readIBlock8(blocksize int) ([]complex128, error) {
	if blocksize%2 != 0 {
		return []complex128{}, errors.New("blocksize must be even")
	}

	result := make([]complex128, blocksize)

	buf := make([]byte, blocksize)
	_, err := r.in.Read(buf)
	if err != nil {
		return []complex128{}, errors.Wrap(err, "cannot read block of 8-bit samples")
	}

	for i := 0; i < len(buf); i++ {
		sample := normalizeSampleUint8(buf[i])
		result[i] = complex(sample, 0)
	}

	return result, nil
}

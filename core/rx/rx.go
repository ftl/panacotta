package rx

import (
	"io"
	"log"
	"math"
	"sync"
	"time"

	"github.com/ftl/panacotta/core"
	"github.com/pkg/errors"
)

// New instance of the receiver.
func New(in io.ReadCloser, fftSink FFTSink) *Receiver {
	result := Receiver{
		in:        in,
		readBlock: readIQBlock8,
		fft:       newFFT(),
		fftSink:   fftSink,

		blockSize: 131072,
	}
	return &result
}

// Receiver type
type Receiver struct {
	in        io.ReadCloser
	readBlock blockReader
	fft       *fft
	fftSink   FFTSink

	blockSize int // fix

	vfoFrequency    core.Frequency      // updated from outside
	rangeOfInterest core.FrequencyRange // depends on vfoFrequency

	ifCenter    core.Frequency // fix, corresponds to the vfoFrequency in the IF range
	rxCenter    core.Frequency // fix, maybe variable depending on the rangeOfInterest
	rxBandwidth core.Frequency // == sample rate, fix
}

// FFTData type.
type FFTData struct {
	Samples []float64
	Range   core.FrequencyRange
}

// FFTSink receives FFT data from the receiver.
type FFTSink func([]float64)

type blockReader func(in io.Reader, blocksize int) ([]complex128, error)

// Run this receiver.
func (r *Receiver) Run(stop chan struct{}, wait *sync.WaitGroup) {
	wait.Add(1)
	go func() {
		defer wait.Done()
		defer r.in.Close()

		for {
			select {
			case <-time.After(1 * time.Millisecond):
				block, err := r.readBlock(r.in, r.blockSize)
				if err == io.EOF {
					log.Print("Waiting for data")
					continue
				} else if err != nil {
					log.Print("Reading incoming data failed:", err)
					continue
				}

				_, fftdata := r.fft.calculate(block)
				r.fftSink(fftdata)
			case <-stop:
				return
			}
		}
	}()
}

func readIQBlock8(in io.Reader, blocksize int) ([]complex128, error) {
	if blocksize%2 != 0 {
		return []complex128{}, errors.New("blocksize must be even")
	}

	result := make([]complex128, blocksize)

	buf := make([]byte, blocksize*2)
	_, err := in.Read(buf)
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

func readIBlock8(in io.Reader, blocksize int) ([]complex128, error) {
	if blocksize%2 != 0 {
		return []complex128{}, errors.New("blocksize must be even")
	}

	result := make([]complex128, blocksize)

	buf := make([]byte, blocksize)
	_, err := in.Read(buf)
	if err != nil {
		return []complex128{}, errors.Wrap(err, "cannot read block of 8-bit samples")
	}

	for i := 0; i < len(buf); i++ {
		sample := normalizeSampleUint8(buf[i])
		result[i] = complex(sample, 0)
	}

	return result, nil
}

func normalizeSampleUint8(s byte) float64 {
	return (float64(s) - float64(math.MaxInt8)) / float64(math.MaxInt8)
}

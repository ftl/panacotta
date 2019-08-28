package rx

import (
	"math"
	"math/cmplx"
)

type step func(in <-chan complex128) <-chan complex128

func buildPipeline(initialBlockSize int, steps ...step) (in chan<- complex128, out <-chan complex128) {
	inRaw := make(chan complex128, initialBlockSize)
	var d <-chan complex128
	d = inRaw
	for _, s := range steps {
		d = s(d)
	}
	in = inRaw
	out = d
	return
}

func serializeBlock(to chan<- complex128, block []complex128) {
	// startTime := time.Now()
	// defer log.Printf("serializeBlock %v", time.Now().Sub(startTime))

	for _, s := range block {
		to <- s
	}
}

func accumulateSamples(in <-chan complex128) <-chan []complex128 {
	blockSize := cap(in)
	result := make(chan []complex128)

	go func() {
		defer close(result)
		out := make([]complex128, blockSize)
		var i int
		for s := range in {
			out[i] = s
			i++
			if i == blockSize {
				result <- out
				out = make([]complex128, blockSize)
				i = 0
			}
		}
	}()

	return result
}

func fir(coeff []float64) step {
	return func(in <-chan complex128) <-chan complex128 {
		result := make(chan complex128, cap(in))

		go func() {
			defer close(result)
			order := len(coeff)
			buf := make([]complex128, order)
			bufIndex := 0
			for s := range in {
				var out complex128

				buf[bufIndex] = s
				for j, c := range coeff {
					bi := (order + bufIndex - j) % order
					out += buf[bi] * complex(c, 0)
				}
				bufIndex = (bufIndex + 1) % order
				result <- out
			}
		}()

		return result
	}
}

func cfir(coeff []complex128) step {
	return func(in <-chan complex128) <-chan complex128 {
		result := make(chan complex128, cap(in))

		go func() {
			defer close(result)
			order := len(coeff)
			buf := make([]complex128, order)
			bufIndex := 0
			for s := range in {
				var out complex128

				buf[bufIndex] = s
				for j, c := range coeff {
					bi := (order + bufIndex - j) % order
					out += buf[bi] * c
				}
				bufIndex = (bufIndex + 1) % order
				result <- out
			}
		}()

		return result
	}
}

func downsample(factor int) step {
	return func(in <-chan complex128) <-chan complex128 {
		oldBlockSize := cap(in)
		newBlockSize := oldBlockSize / factor
		if oldBlockSize%factor != 0 {
			newBlockSize++
		}
		result := make(chan complex128, newBlockSize)

		go func() {
			defer close(result)
			i := 0
			for s := range in {
				if i == 0 {
					result <- s
				}
				i = (i + 1) % factor
			}
		}()

		return result
	}
}

func shift(delta float64) step {
	return func(in <-chan complex128) <-chan complex128 {
		blockSize := cap(in)
		result := make(chan complex128, blockSize)

		go func() {
			defer close(result)
			i := 0
			for s := range in {
				t := float64(i) / float64(blockSize)
				result <- s * cmplx.Exp(complex(0, 2.0*math.Pi*delta*t))
				i = (i + 1) % blockSize
			}
		}()

		return result
	}
}

func shiftFIR(coeff []float64, delta float64, blockSize int) []complex128 {
	result := make([]complex128, len(coeff))
	scaledDelta := delta * float64(len(coeff)) / float64(blockSize)

	for i, c := range coeff {
		t := float64(i) / float64(len(coeff))
		result[i] = complex(c, 0) * cmplx.Exp(complex(0, 2.0*math.Pi*scaledDelta*t))
	}

	return result
}

func shiftCFIR(coeff []complex128, delta float64, blockSize int) []complex128 {
	result := make([]complex128, len(coeff))
	scaledDelta := delta * float64(len(coeff)) / float64(blockSize)

	for i, c := range coeff {
		t := float64(i) / float64(len(coeff))
		result[i] = c * cmplx.Exp(complex(0, 2.0*math.Pi*scaledDelta*t))
	}

	return result
}

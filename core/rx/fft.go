package rx

import (
	"math"

	dsp "github.com/mjibson/go-dsp/fft"
)

func newFFT() *fft {
	result := fft{
		smoothingBuffer: make([][]complex128, 5),
		maxResultSize:   10000,
	}
	return &result
}

type fft struct {
	smoothingBuffer [][]complex128
	smoothingIndex  int
	maxResultSize   int
}

func (f *fft) calculate(samplesBlock []complex128, fromBin, toBin int) (raw, smoothed []float64) {
	blockSize := len(samplesBlock)
	data := dsp.FFT(samplesBlock)

	f.smoothingBuffer[f.smoothingIndex] = data
	f.smoothingIndex = (f.smoothingIndex + 1) % len(f.smoothingBuffer)

	resultSize := toBin - fromBin
	offset := fromBin
	raw = make([]float64, resultSize)
	smoothed = make([]float64, resultSize)

	blockCenter := blockSize / 2
	for i := 0; i < blockSize; i++ {
		var resultIndex int
		if i < blockCenter {
			resultIndex = i + blockCenter - offset
		} else {
			resultIndex = i - blockCenter - offset
		}
		if resultIndex < 0 || resultIndex >= resultSize {
			continue
		}

		var re, im float64
		for j := 0; j < len(f.smoothingBuffer); j++ {
			if len(f.smoothingBuffer[j]) != len(data) {
				continue
			}
			pwr1 := math.Pow(im, 2) + math.Pow(re, 2)
			pwr2 := math.Pow(imag(f.smoothingBuffer[j][i]), 2) + math.Pow(real(f.smoothingBuffer[j][i]), 2)
			if pwr1 < pwr2 {
				re = real(f.smoothingBuffer[j][i])
				im = imag(f.smoothingBuffer[j][i])
			}
		}

		raw[resultIndex] = normalizeFFTValue(real(data[i]), imag(data[i]))
		smoothed[resultIndex] = normalizeFFTValue(re, im)
	}

	for len(raw) > f.maxResultSize {
		raw = reduce(raw)
		smoothed = reduce(smoothed)
	}

	return
}

func normalizeFFTValue(re, im float64) float64 {
	pwr := math.Pow(im, 2) + math.Pow(re, 2)
	return 10.0*math.Log10(pwr+1.0e-20) + 0.5
}

func reduce(data []float64) []float64 {
	result := make([]float64, (len(data)/2)-(len(data)%2))
	for i := 0; i < len(data); i += 2 {
		j := i / 2
		if j >= len(result) {
			break
		}
		switch {
		case i < 1:
			result[j] = data[i]
		case i < len(data)-1:
			result[j] = (data[i] + data[i+1]) / 2
		case i == len(data)-1:
			result[j] = data[i]
		}
	}
	return result
}

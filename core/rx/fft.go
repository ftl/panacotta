package rx

import (
	"math"

	dsp "github.com/mjibson/go-dsp/fft"
)

func newFFT() *fft {
	result := fft{
		smoothingBuffer: make([][]float64, 5),
		maxResultSize:   5000,
	}
	return &result
}

type fft struct {
	smoothingBuffer [][]float64
	smoothingIndex  int
	maxResultSize   int
}

func (f *fft) calculate(samplesBlock []complex128, fromBin, toBin int) (raw, smoothed []float64) {
	blockSize := len(samplesBlock)
	data := dsp.FFT(samplesBlock)

	resultSize := toBin - fromBin
	offset := fromBin

	if len(f.smoothingBuffer[f.smoothingIndex]) != blockSize {
		f.smoothingBuffer[f.smoothingIndex] = make([]float64, blockSize)
	}
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

		f.smoothingBuffer[f.smoothingIndex][i] = normalizeFFTValue(data[i])

		if resultIndex >= 0 && resultIndex < resultSize {
			var smoothedValue float64
			for j := 0; j < len(f.smoothingBuffer); j++ {
				if len(f.smoothingBuffer[j]) != len(data) {
					continue
				}
				smoothedValue = math.Max(smoothedValue, f.smoothingBuffer[j][i])
			}

			raw[resultIndex] = f.smoothingBuffer[f.smoothingIndex][i]
			smoothed[resultIndex] = smoothedValue
		}
	}
	f.smoothingIndex = (f.smoothingIndex + 1) % len(f.smoothingBuffer)

	for len(raw) > f.maxResultSize {
		raw = reduce(raw)
		smoothed = reduce(smoothed)
	}

	return
}

func normalizeFFTValue(v complex128) float64 {
	pwr := math.Pow(imag(v), 2) + math.Pow(real(v), 2)
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

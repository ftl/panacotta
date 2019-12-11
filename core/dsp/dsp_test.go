package dsp

import (
	"fmt"
	"math"
	"testing"
	"time"

	dsp "github.com/mjibson/go-dsp/fft"
	"github.com/stretchr/testify/assert"

	"github.com/ftl/panacotta/core"
)

func TestStop(t *testing.T) {
	dsp := New(1024, 70000, -256)
	stop := make(chan struct{})
	go dsp.Run(stop)

	select {
	case <-dsp.FFT():
		assert.Fail(t, "FFT should be open while running")
	default:
	}

	close(stop)

	select {
	case <-dsp.FFT():
	case <-time.After(10 * time.Millisecond):
		assert.Fail(t, "FFT should be closed when stopped")
	}
}

func TestSamplesRoundtrip(t *testing.T) {
	dsp := New(1024, 70000, -256)
	stop := make(chan struct{})
	defer close(stop)
	go dsp.Run(stop)

	dsp.ProcessSamples(make([]complex128, 512), core.FrequencyRange{From: 7050000, To: 7100000}, core.VFO{Frequency: 7075000})
	select {
	case <-dsp.FFT():
	case <-time.After(10 * time.Millisecond):
		assert.Fail(t, "missing result from processing samples")
	}
}

func TestFFTTonePeak(t *testing.T) {
	blockSize := 16

	for f := -0.5; f <= 0.5; f += 0.01 {
		t.Run(fmt.Sprintf("%.0f", f), func(t *testing.T) {
			samples := tone(blockSize, f)
			fft := dsp.FFT(samples)
			magnitudes := make([]float64, len(fft))
			for i, c := range fft {
				magnitudes[i] = fftValue2dBm(c, blockSize)
			}

			peak := peakIndex(f, blockSize)
			left := peak - 1
			if left < 0 {
				left = blockSize + left
			}
			right := (peak + 1) % blockSize
			for i, m := range magnitudes {
				if i == left || i == right {
					assert.Truef(t, (magnitudes[peak] > m) || math.Abs(m-magnitudes[peak]) < 1.0e-20,
						"close %d:%f !< %d:%f", i, m, peak, magnitudes[peak])
				} else if i != peak {
					assert.Truef(t, magnitudes[peak]-m > 0.4, "%d:%f !< %d:%f", i, m, peak, magnitudes[peak])
				}
			}
		})
	}
}

func TestShift(t *testing.T) {
	blockSize := 16
	for f := -0.5; f <= 0.5; f += 0.001 {
		samples := tone(blockSize, f)

		shifted := shift(samples, -f)
		fft := dsp.FFT(shifted)
		magnitudes := make([]float64, len(fft))
		for i, c := range fft {
			magnitudes[i] = fftValue2dBm(c, blockSize)
		}

		peak := 0
		left := blockSize - 1
		right := 1
		for i, m := range magnitudes {
			if i == left || i == right {
				assert.Truef(t, (magnitudes[peak] > m) || math.Abs(m-magnitudes[peak]) < 1.0e-20,
					"close %d:%f !< %d:%f", i, m, peak, magnitudes[peak])
			} else if i != peak {
				assert.Truef(t, magnitudes[peak]-m > 0.4, "%d:%f !< %d:%f", i, m, peak, magnitudes[peak])
			}
		}
	}
}

func TestFilter(t *testing.T) {
	testCases := []struct {
		samples  []complex128
		filter   []complex128
		expected []complex128
	}{
		{
			[]complex128{1},
			[]complex128{11},
			[]complex128{11},
		},
		{
			[]complex128{1, 2},
			[]complex128{11, 7},
			[]complex128{11, 29},
		},
		{
			[]complex128{1, 2, 3},
			[]complex128{11, 7},
			[]complex128{11, 29, 47},
		},
		{
			[]complex128{1, 2, 3, 4},
			[]complex128{11, 7},
			[]complex128{11, 29, 47, 65},
		},
		{
			[]complex128{1, 2, 3, 4, 5},
			[]complex128{11, 7},
			[]complex128{11, 29, 47, 65, 83},
		},
	}
	for _, tC := range testCases {
		t.Run(fmt.Sprintf("%v", tC.samples), func(t *testing.T) {
			actual := filter(tC.samples, tC.filter)
			assert.Equal(t, tC.expected, actual)
		})
	}
}

func BenchmarkFilter(b *testing.B) {
	for i := 0; i < b.N; i++ {
		filter([]complex128{1, 2, 3, 4, 5}, []complex128{11, 7})
	}
}

func BenchmarkShiftAndFilter(b *testing.B) {
	for i := 0; i < b.N; i++ {
		shiftAndFilter([]complex128{1, 2, 3, 4, 5}, 0.1, []complex128{11, 7})
	}
}

func TestDecimate(t *testing.T) {
	testCases := []struct {
		samples    []complex128
		decimation int
		expected   []complex128
	}{
		{
			[]complex128{1},
			1,
			[]complex128{1},
		},
		{
			[]complex128{1, 2, 3, 4},
			2,
			[]complex128{1, 3},
		},
		{
			[]complex128{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12},
			4,
			[]complex128{1, 5, 9},
		},
	}
	for _, tC := range testCases {
		t.Run(fmt.Sprintf("%v", tC.samples), func(t *testing.T) {
			filter := []complex128{1}
			actual := decimate(tC.samples, tC.decimation, filter)
			assert.Equal(t, tC.expected, actual)
		})
	}
}

func BenchmarkDecimate(b *testing.B) {
	for i := 0; i < b.N; i++ {
		decimate([]complex128{1, 2, 3, 4, 5, 6, 7, 8}, 2, []complex128{11, 7})
	}
}

func BenchmarkShiftAndDecimate(b *testing.B) {
	for i := 0; i < b.N; i++ {
		shiftAndDecimate([]complex128{1, 2, 3, 4, 5, 6, 7, 8}, 0.1, 2, []complex128{11, 7})
	}
}

func TestFIRLowpassGoldenMaster(t *testing.T) {
	order := 9
	cutOff := 0.25
	actual := firLowpass(order, cutOff)
	expected := []complex128{(2.7647317044901985e-34 + 0i), (-0.007206152757970339 + 0i), (6.773477102722069e-18 + 0i), (0.251676223158653 + 0i), (0.5110598591986345 + 0i), (0.2516762231586531 + 0i), (6.773477102722073e-18 + 0i), (-0.007206152757970341 + 0i), (2.7647317044901985e-34 + 0i)}

	assert.Equal(t, expected, actual)
}

func TestFindBlocksize(t *testing.T) {
	testCases := []struct {
		value    int
		max      int
		expected int
	}{
		{0, 16, 0},
		{1, 16, 1},
		{2, 16, 2},
		{7, 16, 8},
		{15, 8, 8},
		{2500, 8192, 4096},
		{2500, 2048, 2048},
	}
	for _, tC := range testCases {
		t.Run(fmt.Sprintf("%d", tC.value), func(t *testing.T) {
			actual := findBlocksize(tC.value, tC.max)
			assert.Equal(t, tC.expected, actual)
		})
	}
}

func peakIndex(frequencyRate float64, blockSize int) int {
	peak := int(math.Round(frequencyRate * float64(blockSize)))
	if peak < 0 {
		return blockSize + peak
	}
	return peak
}

func tone(blockSize int, frequencyRate float64) []complex128 {
	result := make([]complex128, blockSize)

	ω := 2 * math.Pi * frequencyRate
	for i := range result {
		t := float64(i)
		re := math.Cos(ω * t)
		im := math.Sin(ω * t)
		result[i] = complex(re, im)
	}

	return result
}

func tones(blockSize int, frequencyRates ...float64) []complex128 {
	result := make([]complex128, blockSize)

	for _, f := range frequencyRates {
		t := tone(blockSize, f)
		for i := range result {
			result[i] += t[i]
		}
	}

	return result
}

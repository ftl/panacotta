package dsp

import (
	"fmt"
	"log"
	"math"
	"math/cmplx"

	// dsp "github.com/mjibson/go-dsp/fft"
	"github.com/mjibson/go-dsp/dsputils"
	dsp "github.com/mjibson/go-dsp/fft"
	"github.com/mjibson/go-dsp/window"

	"github.com/ftl/panacotta/core"
)

const smoothingLength = 5 // 10 // TODO make configurable
type smoother interface {
	Put([]float64) []float64
}

func New(sampleRate int, ifFrequency, rxOffset core.Frequency) *DSP {
	result := &DSP{
		workInput: make(chan work, 1),
		fft:       make(chan core.FFT, 1),

		sampleRate:  sampleRate,
		ifCenter:    ifFrequency,
		rxCenter:    ifFrequency + rxOffset,
		filterCoeff: firLowpass(27, 1.0/4.8),
		smoother:    newAverager(smoothingLength, 0),
	}

	return result
}

func NewFullRange(sampleRate int, ifFrequency, rxOffset core.Frequency) *DSP {
	result := New(sampleRate, ifFrequency, rxOffset)
	result.fullRangeMode = true
	return result
}

type DSP struct {
	workInput chan work
	fft       chan core.FFT

	sampleRate int
	ifCenter   core.Frequency
	rxCenter   core.Frequency // actual receiving frequency

	vfo             core.VFO
	fftRange        core.FrequencyRange
	fftWindow       []complex128
	inputBlockSize  int
	outputBlockSize int
	Δf              float64
	decimation      int
	filterCoeff     []complex128
	filterWindow    []complex128
	fullRangeMode   bool

	smoother smoother
}

type work struct {
	samples  []complex128
	fftRange core.FrequencyRange
	vfo      core.VFO
}

func (d *DSP) Run(stop chan struct{}) {
	defer log.Print("DSP shutdown")
	for {
		select {
		case work := <-d.workInput:
			d.doWork(work)
		case <-stop:
			close(d.fft)
			return
		}
	}
}

func (d *DSP) ProcessSamples(samples []complex128, fftRange core.FrequencyRange, vfo core.VFO) {
	select {
	case d.workInput <- work{samples, fftRange, vfo}:
	default:
		log.Print("process samples hangs")
	}
}

func (d *DSP) FFT() chan core.FFT {
	return d.fft
}

func findBlocksize(width, max int) int {
	result := dsputils.NextPowerOf2(width)
	if result > max {
		return max
	}
	return result
}

func (d *DSP) doWork(work work) {
	if work.fftRange.Width() == 0 {
		return
	}

	oldOutputBlockSize := d.outputBlockSize
	var needReconfiguration bool
	if len(work.samples) != d.inputBlockSize || needReconfiguration {
		d.inputBlockSize = len(work.samples)
		needReconfiguration = true
	}
	if work.vfo.Frequency != d.vfo.Frequency || needReconfiguration {
		d.vfo = work.vfo
		needReconfiguration = true
	}
	if work.fftRange != d.fftRange || needReconfiguration {
		d.fftRange = work.fftRange
		d.Δf = float64(d.ifCenter - d.rxCenter - d.fftRange.Center() + d.vfo.Frequency)
		d.outputBlockSize = findBlocksize(int(float64(d.inputBlockSize)/(float64(d.sampleRate)/(2*float64(d.fftRange.Width())))), d.inputBlockSize)
		d.decimation = d.inputBlockSize / d.outputBlockSize
		d.filterWindow = fftLowpass(d.inputBlockSize, d.decimation)
		needReconfiguration = true
	}
	if oldOutputBlockSize != d.outputBlockSize {
		d.smoother = newAverager(smoothingLength, d.outputBlockSize)

		fftWindow := window.Hamming(d.outputBlockSize)
		d.fftWindow = make([]complex128, len(fftWindow))
		for i := range fftWindow {
			d.fftWindow[i] = complex(fftWindow[i], 0)
		}
	}
	if needReconfiguration {
		log.Printf("fftRange %f %f %f (%f) | vfo %f | if %f | rx %f", d.fftRange.From, d.fftRange.Center(), d.fftRange.To, d.fftRange.Width(), d.vfo.Frequency, d.ifCenter, d.rxCenter)
		log.Printf("reconfiguration: %d %d %f %f %f", d.decimation, d.outputBlockSize, d.fftRange.Width(), d.Δf, toRate(d.Δf, d.sampleRate))
	}

	var outputSamples []complex128
	if d.decimation == 1 {
		outputSamples = shift(work.samples, toRate(d.Δf, d.sampleRate))
	} else {
		outputSamples = fftShiftAndDecimate(work.samples, toRate(d.Δf, d.sampleRate), d.decimation, d.filterWindow)
	}

	for i := range outputSamples {
		outputSamples[i] *= d.fftWindow[i]
	}

	fft, mean := fft(outputSamples)
	if smoothingLength > 1 {
		fft = d.smoother.Put(fft)
	}
	peaks, threshold := peaks(fft, mean)

	center := d.fftRange.Center()
	sideband := core.Frequency(d.sampleRate / (2 * d.decimation))
	if d.fullRangeMode {
		fft = padZero(fft, d.inputBlockSize)
		sideband = core.Frequency(d.sampleRate / 2)
	}

	select {
	case d.fft <- core.FFT{
		Data:          fft,
		Range:         core.FrequencyRange{From: center - sideband, To: center + sideband},
		Mean:          mean,
		PeakThreshold: threshold,
		Peaks:         peaks,
	}:
	default:
		log.Print("return FFT hangs")
	}
}

func padZero(samples []float64, size int) []float64 {
	pad := make([]float64, (size-len(samples))/2)
	result := make([]float64, 0, size)
	result = append(result, pad...)
	result = append(result, samples...)
	result = append(result, pad...)
	if len(result) != size {
		panic(fmt.Errorf("wrong size %d != %d expected", len(result), size))
	}
	return result
}

func shift(samples []complex128, shiftRate float64) []complex128 {
	outputSamples := make([]complex128, len(samples))

	ω := 2 * math.Pi * shiftRate
	for i, s := range samples {
		t := float64(i)
		outputSamples[i] = s * cmplx.Exp(complex(0, ω*t)) // shift fftRange to center of fullRange
	}

	return outputSamples
}

func filter(samples []complex128, filterCoeff []complex128) []complex128 {
	blockSize := len(samples)
	filterOrder := len(filterCoeff)

	outputSamples := make([]complex128, blockSize)
	for i := range samples {
		j := (i - filterOrder + 1)
		for k := filterOrder - 1; k >= 0; k-- {
			if j >= 0 && j < blockSize {
				outputSamples[i] += samples[j] * filterCoeff[k]
			}
			j++
		}
	}

	return outputSamples
}

func shiftAndFilter(samples []complex128, shiftRate float64, filterCoeff []complex128) []complex128 {
	ω := 2 * math.Pi * shiftRate

	blockSize := len(samples)
	filterOrder := len(filterCoeff)

	shiftedSamples := make([]complex128, blockSize)
	lastShifted := -1
	outputSamples := make([]complex128, blockSize)

	for i := range samples {
		t := float64(i)

		j := (i - filterOrder + 1)
		for k := filterOrder - 1; k >= 0; k-- {
			if j >= 0 && j < blockSize {
				var s complex128
				if j <= lastShifted {
					s = shiftedSamples[j]
				} else {
					s = samples[j] * cmplx.Exp(complex(0, ω*t))
					shiftedSamples[j] = s
					lastShifted = j
				}
				outputSamples[i] += s * filterCoeff[k]
			}
			j++
		}
	}

	return outputSamples
}

func downsample(samples []complex128, decimation int) []complex128 {
	result := make([]complex128, len(samples)/decimation)
	for i := range result {
		result[i] = samples[i*decimation]
	}
	return result
}

func decimate(samples []complex128, decimation int, filterCoeff []complex128) []complex128 {
	blockSize := len(samples)
	filterOrder := len(filterCoeff)

	outputSamples := make([]complex128, blockSize/decimation)
	outputIndex := 0
	for i := 0; i < blockSize; i += decimation {
		j := (i - filterOrder + 1)
		for k := filterOrder - 1; k >= 0; k-- {
			if j >= 0 && j < blockSize {
				outputSamples[outputIndex] += samples[j] * filterCoeff[k]
			}
			j++
		}
		outputIndex++
	}

	return outputSamples
}

func shiftAndDecimate(samples []complex128, shiftRate float64, decimation int, filterCoeff []complex128) []complex128 {
	ω := 2 * math.Pi * shiftRate

	blockSize := len(samples)
	filterOrder := len(filterCoeff)

	shiftedSamples := make([]complex128, blockSize)
	lastShifted := -1
	outputSamples := make([]complex128, blockSize/decimation)
	outputIndex := 0

	for i := 0; i < blockSize; i += decimation {
		t := float64(i)

		j := (i - filterOrder + 1)
		for k := filterOrder - 1; k >= 0; k-- {
			if j >= 0 && j < blockSize {
				var s complex128
				if j <= lastShifted {
					s = shiftedSamples[j]
				} else {
					s = samples[j] * cmplx.Exp(complex(0, ω*t))
					shiftedSamples[j] = s
					lastShifted = j
				}
				outputSamples[outputIndex] += s * filterCoeff[k]
			}
			j++
		}
		outputIndex++
	}

	return outputSamples
}

func fftDecimate(samples []complex128, decimation int, filterWindow []complex128) []complex128 {
	frequencyDomain := dsp.FFT(samples)
	for i := range frequencyDomain {
		frequencyDomain[i] *= filterWindow[i]
	}
	timeDomain := dsp.IFFT(frequencyDomain)

	result := make([]complex128, len(samples)/decimation)
	for i := 0; i < len(timeDomain); i += decimation {
		result[i/decimation] = timeDomain[i]
	}

	return result
}

func fftShiftAndDecimate(samples []complex128, shiftRate float64, decimation int, filterWindow []complex128) []complex128 {
	frequencyDomain := dsp.FFT(samples)
	blockSize := len(frequencyDomain)
	shiftOffset := int(shiftRate * float64(blockSize))
	shifted := make([]complex128, blockSize)
	for i := range frequencyDomain {
		shiftedIndex := (blockSize + shiftOffset + i) % blockSize
		shifted[shiftedIndex] = frequencyDomain[i] * filterWindow[shiftedIndex]
	}
	timeDomain := dsp.IFFT(shifted)

	result := make([]complex128, len(samples)/decimation)
	for i := 0; i < len(timeDomain); i += decimation {
		result[i/decimation] = timeDomain[i]
	}

	return result
}

func fft(samples []complex128) ([]float64, float64) {
	cfft := dsp.FFT(samples)

	result := make([]float64, len(cfft))
	mean := 0.0
	blockSize := len(result)
	blockCenter := blockSize / 2
	for i, v := range cfft {
		var resultIndex int
		if i < blockCenter {
			resultIndex = i + blockCenter
		} else {
			resultIndex = i - blockCenter
		}

		result[resultIndex] = fftValue2dBm(v, blockSize)
		mean += result[resultIndex]
	}
	mean /= float64(blockSize)
	return result, mean
}

func fftValue2dBm(fftValue complex128, blockSize int) float64 {
	return 10.0 * math.Log10(20*(math.Pow(real(fftValue), 2)+math.Pow(imag(fftValue), 2))/math.Pow(float64(blockSize), 2))
}

func peaks(fft []float64, mean float64) ([]core.PeakIndexRange, float64) {
	if len(fft) == 0 {
		return []core.PeakIndexRange{}, 0
	}

	sum := 0.0
	for _, p := range fft {
		sum += math.Pow(p-mean, 2)
	}
	σ := math.Sqrt(sum / float64(len(fft)))

	threshold := mean + 2*σ
	startI := 0
	max := -200.0
	maxI := 0
	lastMax := -200.0
	lastMaxI := 0
	lastValue := -200.0
	wasRising := false
	wasAbove := false

	result := make([]core.PeakIndexRange, 0, len(fft)/4)
	for i, v := range fft {
		rising := v-lastValue >= 0
		turn := rising != wasRising
		above := v > threshold
		if turn && !rising {
			wasAbove = above
			if max < lastValue {
				max = lastValue
				maxI = i - 1
			}
		} else if turn && rising {
			if wasAbove {
				peak := core.PeakIndexRange{From: startI, To: i - 1, Max: maxI, Value: max}
				isClose := (maxI - lastMaxI) < 5 // this threshold value is arbitrary, it should be configurable
				if isClose && max > lastMax && len(result) > 0 {
					result[len(result)-1] = peak
				} else if isClose && max < lastMax {
					// ignore this peak
				} else {
					result = append(result, peak)
				}
			}
			startI = i - 1
			wasAbove = false
			lastMax = max
			lastMaxI = maxI
			max = v
			maxI = i
		}
		lastValue = v
		wasRising = rising
	}

	return result, threshold
}

func toRate(frequency float64, sampleRate int) float64 {
	return frequency / float64(sampleRate)
}

func firLowpass(order int, cutoffRate float64) []complex128 {
	if order%2 == 0 {
		panic("FIR order must be odd")
	}

	window := window.Blackman(order)
	order2 := (order - 1) / 2
	coeff := make([]float64, order)
	sum := 0.0
	for i := range coeff {
		t := float64(i - order2)
		coeff[i] = sinc(2.0*cutoffRate*t) * window[i]
		sum += coeff[i]
	}

	result := make([]complex128, len(coeff))
	for i := range result {
		result[i] = complex((coeff[i] / sum), 0)
	}
	return result
}

func sinc(x float64) float64 {
	if x == 0 {
		return 1.0
	}
	return math.Sin(math.Pi*x) / (math.Pi * x)
}

func fftLowpass(blockSize int, decimation int) []complex128 {
	impulseResponse := make([]complex128, blockSize)
	filter := firLowpass(blockSize/2+1, 1.0/float64(2*decimation))
	copy(impulseResponse, filter)
	return dsp.FFT(impulseResponse)
}

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

func New(sampleRate int, ifFrequency, rxOffset core.Frequency) *DSP {
	result := DSP{
		workInput: make(chan work, 1),
		fft:       make(chan core.FFT, 1),

		sampleRate: sampleRate,
		ifCenter:   ifFrequency,
		rxCenter:   ifFrequency + rxOffset,
	}

	return &result
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
	inputBlockSize  int
	outputBlockSize int
	Δf              float64
	decimation      int
	filterCoeff     []complex128
	fullRangeMode   bool
}

type work struct {
	samples  []complex128
	fftRange core.FrequencyRange
	vfo      core.VFO
}

func (d *DSP) Run(stop chan struct{}) {
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
	const filterOrder = 15
	if work.fftRange.Width() == 0 {
		return
	}

	needReconfiguration := work.fftRange != d.fftRange || work.vfo != d.vfo || len(work.samples) != d.inputBlockSize
	if needReconfiguration {
		d.fftRange = work.fftRange
		d.vfo = work.vfo
		d.inputBlockSize = len(work.samples)
		d.Δf = -(float64(d.fftRange.Center() - d.vfo.Frequency + d.ifCenter - d.rxCenter))
		d.outputBlockSize = findBlocksize(int(float64(d.inputBlockSize)/(float64(d.sampleRate)/(2*float64(d.fftRange.Width())))), d.inputBlockSize)
		d.decimation = d.inputBlockSize / d.outputBlockSize
		d.filterCoeff = firLowpass(filterOrder, 1.0/(2.0*float64(d.decimation)))
		log.Printf("fftRange %f %f %f (%f) | vfo %f | if %f | rx %f", d.fftRange.From, d.fftRange.Center(), d.fftRange.To, d.fftRange.Width(), d.vfo.Frequency, d.ifCenter, d.rxCenter)
		log.Printf("reconfiguration: %d %d %f %f", d.decimation, d.outputBlockSize, d.fftRange.Width(), d.Δf)
	}

	outputSamples := shiftAndDecimate(work.samples, toRate(d.Δf, d.sampleRate), d.decimation, d.filterCoeff)
	fft := fft(outputSamples)

	center := d.fftRange.Center()
	sideband := core.Frequency(d.sampleRate / (2 * d.decimation))

	if d.fullRangeMode {
		fft = padZero(fft, d.inputBlockSize)
		sideband = core.Frequency(d.sampleRate / 2)
	}

	select {
	case d.fft <- core.FFT{
		Data:  fft,
		Range: core.FrequencyRange{From: center - sideband, To: center + sideband},
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
	filterOrder := len(filterCoeff)
	filterBuf := make([]complex128, filterOrder)
	bufIndex := 0

	outputSamples := make([]complex128, len(samples))

	for i, s := range samples {
		var out complex128

		filterBuf[bufIndex] = s
		for j, c := range filterCoeff {
			bi := (filterOrder + bufIndex - j) % filterOrder
			out += filterBuf[bi] * c
		}
		bufIndex = (bufIndex + 1) % filterOrder

		outputSamples[i] = out
	}

	return outputSamples
}

func shiftAndFilter(samples []complex128, shiftRate float64, filterCoeff []complex128) []complex128 {
	ω := 2 * math.Pi * shiftRate

	filterOrder := len(filterCoeff)
	filterBuf := make([]complex128, filterOrder)
	bufIndex := 0

	outputSamples := make([]complex128, len(samples))

	for i, s := range samples {
		t := float64(i) // "time" for the shift operation
		var out complex128

		filterBuf[bufIndex] = s * cmplx.Exp(complex(0, ω*t)) // shift fftRange to center of fullRange
		for j, c := range filterCoeff {
			bi := (filterOrder + bufIndex - j) % filterOrder
			out += filterBuf[bi] * c
		}

		outputSamples[i] = out
		bufIndex = (bufIndex + 1) % filterOrder
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
	filterOrder := len(filterCoeff)
	filterBuf := make([]complex128, filterOrder)
	bufIndex := 0
	decimCountDown := 0

	outputIndex := 0
	outputSamples := make([]complex128, len(samples)/decimation)

	for _, s := range samples {
		filterBuf[bufIndex] = s
		if decimCountDown <= 0 {
			decimCountDown = decimation - 1

			var out complex128
			for j, c := range filterCoeff {
				bi := (filterOrder + bufIndex - j) % filterOrder
				out += filterBuf[bi] * c
			}
			outputSamples[outputIndex] = out
			outputIndex++
		} else {
			decimCountDown--
		}
		bufIndex = (bufIndex + 1) % filterOrder
	}

	return outputSamples
}

func shiftAndDecimate(samples []complex128, shiftRate float64, decimation int, filterCoeff []complex128) []complex128 {
	ω := 2 * math.Pi * shiftRate

	filterOrder := len(filterCoeff)
	filterBuf := make([]complex128, filterOrder)
	bufIndex := 0
	decimCountDown := 0

	outputIndex := 0
	outputSamples := make([]complex128, len(samples)/decimation)

	for i, s := range samples {
		t := float64(i) // "time" for the shift operation

		filterBuf[bufIndex] = s * cmplx.Exp(complex(0, ω*t)) // shift fftRange to center of fullRange
		if decimCountDown <= 0 {
			decimCountDown = decimation - 1

			var out complex128
			for j, c := range filterCoeff {
				bi := (filterOrder + bufIndex - j) % filterOrder
				out += filterBuf[bi] * c
			}
			outputSamples[outputIndex] = out
			outputIndex++
		} else {
			decimCountDown--
		}
		bufIndex = (bufIndex + 1) % filterOrder
	}

	return outputSamples
}

func fft(samples []complex128) []float64 {
	cfft := dsp.FFT(samples)
	result := make([]float64, len(cfft))
	blockSize := len(result)
	blockCenter := blockSize / 2
	for i, v := range cfft {
		var resultIndex int
		if i < blockCenter {
			resultIndex = i + blockCenter
		} else {
			resultIndex = i - blockCenter
		}
		result[resultIndex] = fftValueToDB(v, blockSize)
	}
	return result
}

func fftValueToDB(fftValue complex128, blockSize int) float64 {
	return 20.0 * math.Log10(2*math.Sqrt(math.Pow(real(fftValue), 2)+math.Pow(imag(fftValue), 2))/float64(blockSize))
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
	log.Print(result)
	return result
}

func sinc(x float64) float64 {
	if x == 0 {
		return 1.0
	}
	return math.Sin(math.Pi*x) / (math.Pi * x)
}

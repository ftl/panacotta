package rx

import (
	"log"
	"math"
	"sync"
	"time"

	"github.com/ftl/panacotta/core"
	"github.com/ftl/panacotta/core/bandplan"
)

// New instance of the receiver.
func New(in core.SamplesInput, blockSize int, ifCenter, rxCenter, rxBandwidth core.Frequency) *Receiver {
	result := Receiver{
		in:               in,
		readBlock:        readIQ8Block,
		samplesBlockSize: blockSize,
		fft:              newFFT(),

		vfoBand: bandplan.UnknownBand,

		ifCenter:    ifCenter,
		rxCenter:    rxCenter,
		rxOffset:    ifCenter - rxCenter,
		rxBandwidth: rxBandwidth,

		// viewMode:    ViewFullSpectrum,
		setViewMode: make(chan ViewMode),
	}
	return &result
}

// Receiver type
type Receiver struct {
	in               core.SamplesInput
	readBlock        samplesReader
	samplesBlockSize int
	fft              *fft

	vfoFrequency core.Frequency      // updated from outside
	vfoBand      bandplan.Band       // depends on the vfoFrequency
	vfoROI       core.FrequencyRange // depends on vfoFrequency
	vfoMode      string              // updated from outside
	vfoBandwidth core.Frequency      // updated from outside

	ifCenter           core.Frequency      // fix, corresponds to the vfoFrequency in the IF range
	rxCenter           core.Frequency      // fix
	rxOffset           core.Frequency      // == r.ifCenter - r.rxCenter
	rxBandwidth        core.Frequency      // == sample rate, fix
	rxROI              core.FrequencyRange // corresponds to the vfoROI in the IF range
	processedBandwidth core.Frequency      // depends on the DSP processing, < rxBandwidth

	fftAvailableCallbacks []FFTAvailable
	vfoChangedCallbacks   []VFOChanged

	viewMode    ViewMode
	setViewMode chan ViewMode
}

// FFTAvailable is called when new FFT data is available.
type FFTAvailable func([]float64)

// VFOChanged is called when the VFO setup (frequency, band, ROI, mode, bandwidth), changes.
type VFOChanged func(core.Frequency, bandplan.Band, core.FrequencyRange, string, core.Frequency)

// ViewMode of the panorama.
type ViewMode int

// All view modes.
const (
	ViewFixed ViewMode = iota
	ViewCentered
	ViewFullSpectrum
)

// SampleSource provides blocks with samples.
type SampleSource interface {
	ReadBlock() ([]complex128, error)
}

// Run this receiver.
func (r *Receiver) Run(stop chan struct{}, wait *sync.WaitGroup) {
	wait.Add(1)

	// samplesFrequencyScale := float64(r.samplesBlockSize) / float64(r.rxBandwidth)
	dspIn, dspOut := buildPipeline(
		r.samplesBlockSize,
		// cfir(shiftFIR(LPF180k, float64(r.rxOffset)*samplesFrequencyScale, r.samplesBlockSize)),
		// shift(float64(r.rxOffset)*samplesFrequencyScale),
		// downsample(4),
	)
	processedBlock := accumulateSamples(dspOut)
	downsamplingRate := cap(dspIn) / cap(dspOut)
	r.processedBandwidth = r.rxBandwidth / core.Frequency(downsamplingRate)

	log.Printf("shift by %v", float64(r.ifCenter-r.rxCenter))
	log.Printf("downsampling by %v", downsamplingRate)

	go func() {
		defer wait.Done()
		defer r.shutdown()
		defer close(dspIn)

		lastBlock := time.Now()

		for {
			select {
			case rawBlock := <-r.in.Samples():
				blockTime := time.Now().Sub(lastBlock)
				blocksPerSecond := int(time.Second / blockTime)
				if blocksPerSecond > 25 {
					continue
				}

				lastBlock = time.Now()
				// log.Printf("New block after %v %v", blockTime, time.Second/blockTime)

				// dspStart := time.Now()
				r.readBlock(dspIn, rawBlock)
				block := <-processedBlock
				// log.Printf("DSP %v", time.Now().Sub(dspStart))

				blockSize := len(block)
				hzPerBin := r.processedBandwidth / core.Frequency(blockSize)

				fromBin := int(r.rxROI.From / hzPerBin)
				toBin := int(r.rxROI.To / hzPerBin)

				// fftStart := time.Now()
				_, fftdata := r.fft.calculate(block, fromBin, toBin)
				// log.Printf("FFT %v %d", time.Now().Sub(fftStart), blockSize)
				for _, fftAvailable := range r.fftAvailableCallbacks {
					fftAvailable(fftdata)
				}
			case viewMode := <-r.setViewMode:
				r.viewMode = viewMode
				r.updateROI()
				r.notifyVFOChange()
			case <-stop:
				return
			}
		}
	}()
}

func (r *Receiver) shutdown() {
	r.in.Close()
	log.Print("Receiver shutdown")
}

// SetVFOFrequency sets the current VFO frequency.
func (r *Receiver) SetVFOFrequency(f core.Frequency) {
	r.vfoFrequency = f
	r.updateROI()
	r.notifyVFOChange()
}

// SetVFOMode sets the VFO's current mode and bandwidth
func (r *Receiver) SetVFOMode(mode string, bandwidth core.Frequency) {
	r.vfoMode = mode
	r.vfoBandwidth = bandwidth
	r.notifyVFOChange()
}

func (r *Receiver) updateROI() {
	f := r.vfoFrequency
	if !r.vfoBand.Contains(f) {
		band := bandplan.IARURegion1.ByFrequency(f)
		if band.Width() > 0 {
			r.vfoBand = band
		}
	}

	switch r.viewMode {
	case ViewFixed:
		r.vfoROI = core.FrequencyRange{From: r.vfoBand.From - 10000.0, To: r.vfoBand.To + 10000.0}
		r.rxROI = core.FrequencyRange{From: r.vfoToRx(r.vfoROI.From), To: r.vfoToRx(r.vfoROI.To)}

	case ViewCentered:
		r.vfoROI = core.FrequencyRange{From: f - 20000, To: f + 20000}
		r.rxROI = core.FrequencyRange{From: r.vfoToRx(r.vfoROI.From), To: r.vfoToRx(r.vfoROI.To)}

	case ViewFullSpectrum:
		r.rxROI = core.FrequencyRange{From: 0, To: r.processedBandwidth}
		r.vfoROI = core.FrequencyRange{From: r.rxToVFO(r.rxROI.From), To: r.rxToVFO(r.rxROI.To)}

	}
	log.Print(r.rxROI, r.vfoROI)
}

func (r *Receiver) notifyVFOChange() {
	for _, vfoChanged := range r.vfoChangedCallbacks {
		vfoChanged(r.vfoFrequency, r.vfoBand, r.vfoROI, r.vfoMode, r.vfoBandwidth)
	}
}

// SetViewMode of the receiver.
func (r *Receiver) SetViewMode(viewMode ViewMode) {
	r.setViewMode <- viewMode
}

// ViewMode of the receiver.
func (r *Receiver) ViewMode() ViewMode {
	return r.viewMode
}

func (r *Receiver) vfoToRx(f core.Frequency) core.Frequency {
	return core.Frequency(r.processedBandwidth/2) - (r.vfoFrequency - f) - r.rxOffset
}

func (r *Receiver) rxToVFO(f core.Frequency) core.Frequency {
	return f + r.vfoFrequency - core.Frequency(r.processedBandwidth/2) + r.rxOffset
}

// OnFFTAvailable registers the given callback to be notified when new FFT data is available.
func (r *Receiver) OnFFTAvailable(f FFTAvailable) {
	r.fftAvailableCallbacks = append(r.fftAvailableCallbacks, f)
}

// OnVFOChange registers the given callback to be notified when the VFO setup (frequency, ROI) changes.
func (r *Receiver) OnVFOChange(f VFOChanged) {
	r.vfoChangedCallbacks = append(r.vfoChangedCallbacks, f)
}

type samplesReader func(chan<- complex128, []byte)

func readIQ8Block(to chan<- complex128, block []byte) {
	// startTime := time.Now()
	// defer log.Printf("readIQ8Block %v", time.Now().Sub(startTime))

	if len(block)%2 != 0 {
		log.Printf("blocksize must be even")
	}

	for i := 0; i < len(block); i += 2 {
		qSample := normalizeSampleUint8(block[i])
		iSample := normalizeSampleUint8(block[i+1])
		to <- complex(iSample, qSample)
	}
}

func readI8Block(to chan<- complex128, block []byte) {
	// startTime := time.Now()
	// defer log.Printf("readI8Block %v", time.Now().Sub(startTime))

	if len(block)%2 != 0 {
		log.Printf("blocksize must be even")
	}

	for _, s := range block {
		sample := normalizeSampleUint8(s)
		to <- complex(sample, 0)
	}
}

func normalizeSampleUint8(s byte) float64 {
	return (float64(s) - float64(math.MaxInt8)) / float64(math.MaxInt8)
}

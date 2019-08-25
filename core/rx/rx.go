package rx

import (
	"io"
	"log"
	"math"
	"sync"
	"time"

	"github.com/ftl/panacotta/core"
	"github.com/ftl/panacotta/core/bandplan"
	"github.com/pkg/errors"
)

// New instance of the receiver.
func New(in io.ReadCloser, ifCenter, rxCenter, rxBandwidth core.Frequency) *Receiver {
	result := Receiver{
		in:      in,
		samples: &samplesInput{in: in, blockSize: 131072, readBlock: readIQBlock8},
		fft:     newFFT(),

		vfoBand: bandplan.UnknownBand,

		ifCenter:    ifCenter,
		rxCenter:    rxCenter,
		rxBandwidth: rxBandwidth,

		viewMode:    ViewFullSpectrum,
		setViewMode: make(chan ViewMode),
	}
	return &result
}

// Receiver type
type Receiver struct {
	in      io.ReadCloser
	samples SampleSource
	fft     *fft

	vfoFrequency core.Frequency      // updated from outside
	vfoBand      bandplan.Band       // depends on the vfoFrequency
	vfoROI       core.FrequencyRange // depends on vfoFrequency

	ifCenter    core.Frequency      // fix, corresponds to the vfoFrequency in the IF range
	rxCenter    core.Frequency      // fix
	rxBandwidth core.Frequency      // == sample rate, fix
	rxROI       core.FrequencyRange // corresponds to the vfoROI in the IF range

	fftAvailableCallbacks []FFTAvailable
	vfoChangedCallbacks   []VFOChanged

	viewMode    ViewMode
	setViewMode chan ViewMode
}

// FFTAvailable is called when new FFT data is available.
type FFTAvailable func([]float64)

// VFOChanged is called when the VFO setup (frequency, band, ROI), changes.
type VFOChanged func(core.Frequency, bandplan.Band, core.FrequencyRange)

// ViewMode of the panorama.
type ViewMode int

// All view modes.
const (
	ViewFullBand ViewMode = iota
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
	go func() {
		defer wait.Done()
		defer r.shutdown()

		for {
			select {
			case <-time.After(1 * time.Millisecond):
				block, err := r.samples.ReadBlock()
				if err == io.EOF {
					log.Print("Waiting for data")
					continue
				} else if err != nil {
					log.Print("Reading incoming data failed: ", err)
					continue
				}

				blockSize := len(block)
				hzPerBin := r.rxBandwidth / core.Frequency(blockSize)

				fromBin := int(r.rxROI.From / hzPerBin)
				toBin := int(r.rxROI.To / hzPerBin)

				_, fftdata := r.fft.calculate(block, fromBin, toBin)
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

func (r *Receiver) updateROI() {
	f := r.vfoFrequency
	if !r.vfoBand.Contains(f) {
		band := bandplan.IARURegion1.ByFrequency(f)
		if band.Width() > 0 {
			r.vfoBand = band
		}
	}

	switch r.viewMode {
	case ViewFullBand:
		r.vfoROI = core.FrequencyRange{From: r.vfoBand.From - 10000.0, To: r.vfoBand.To + 10000.0}
		r.rxROI = core.FrequencyRange{From: r.vfoToRx(r.vfoROI.From), To: r.vfoToRx(r.vfoROI.To)}

	case ViewCentered:
		r.vfoROI = core.FrequencyRange{From: f - 20000, To: f + 20000}
		r.rxROI = core.FrequencyRange{From: r.vfoToRx(r.vfoROI.From), To: r.vfoToRx(r.vfoROI.To)}

	case ViewFullSpectrum:
		r.rxROI = core.FrequencyRange{From: 0, To: r.rxBandwidth}
		r.vfoROI = core.FrequencyRange{From: r.rxToVFO(r.rxROI.From), To: r.rxToVFO(r.rxROI.To)}

	}
	log.Print(r.rxROI, r.vfoROI)
}

func (r *Receiver) notifyVFOChange() {
	for _, vfoChanged := range r.vfoChangedCallbacks {
		vfoChanged(r.vfoFrequency, r.vfoBand, r.vfoROI)
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
	return core.Frequency(r.rxBandwidth/2) - (r.vfoFrequency - f) - (r.ifCenter - r.rxCenter)
}

func (r *Receiver) rxToVFO(f core.Frequency) core.Frequency {
	return f - r.rxCenter + r.ifCenter + r.vfoFrequency - core.Frequency(r.rxBandwidth/2)
}

// OnFFTAvailable registers the given callback to be notified when new FFT data is available.
func (r *Receiver) OnFFTAvailable(f FFTAvailable) {
	r.fftAvailableCallbacks = append(r.fftAvailableCallbacks, f)
}

// OnVFOChange registers the given callback to be notified when the VFO setup (frequency, ROI) changes.
func (r *Receiver) OnVFOChange(f VFOChanged) {
	r.vfoChangedCallbacks = append(r.vfoChangedCallbacks, f)
}

type samplesInput struct {
	in        io.Reader
	blockSize int
	readBlock blockReader
}

type blockReader func(in io.Reader, blocksize int) ([]complex128, error)

func (s *samplesInput) ReadBlock() ([]complex128, error) {
	return s.readBlock(s.in, s.blockSize)
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

package rtlsdr

import (
	"log"
	"sync"
	"time"

	rtl "github.com/jpoirier/gortlsdr"
)

// Open the RTL-SDR dongle for reading.
func Open(centerFrequency int, sampleRate int, blockSize int, frequencyCorrection int) (*Dongle, error) {
	device, err := rtl.Open(0)
	if err != nil {
		return nil, err
	}

	err = device.SetSampleRate(sampleRate)
	if err != nil {
		device.Close()
		log.Print("SetSampleRate failed", err)
		return nil, err
	}
	log.Printf("GetSampleRate: %d\n", device.GetSampleRate())

	err = device.SetCenterFreq(centerFrequency)
	if err != nil {
		device.Close()
		log.Print("SetCenterFreq failed", err)
		return nil, err
	}

	err = device.ResetBuffer()
	if err != nil {
		device.Close()
		log.Print("ResetBuffer failed", err)
		return nil, err
	}

	err = device.SetFreqCorrection(frequencyCorrection)
	if err != nil {
		device.Close()
		log.Print("SetFreqCorrection failed", err)
		return nil, err
	}

	result := Dongle{
		device:    device,
		asyncRead: new(sync.WaitGroup),
		samples:   make(chan []byte, 1),
	}

	go func() {
		result.asyncRead.Add(1)
		result.device.ReadAsync(result.incomingData, nil, 0, blockSize*2)
		result.asyncRead.Done()
	}()

	return &result, nil
}

// Dongle represents the RTL-SDR dongle.
type Dongle struct {
	device    *rtl.Context
	blockSize int
	asyncRead *sync.WaitGroup
	lastInput time.Time
	samples   chan []byte
}

// Samples from the dongle
func (d *Dongle) Samples() <-chan []byte {
	return d.samples
}

// Close the dongle.
func (d *Dongle) Close() error {
	d.device.CancelAsync()
	d.asyncRead.Wait()
	close(d.samples)
	return d.device.Close()
}

func (d *Dongle) incomingData(data []byte) {
	select {
	case d.samples <- data:
		d.lastInput = time.Now()
	default:
		// log.Print("RTL buffer overflow, dropping incoming data")
	}
}

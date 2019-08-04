package rtlsdr

import (
	"bytes"
	"log"
	"sync"
	"time"

	rtl "github.com/jpoirier/gortlsdr"
)

// Open the RTL-SDR dongle for reading.
func Open(centerFrequency int, sampleRate int, frequencyCorrection int) (*Dongle, error) {
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
	}

	go func() {
		result.asyncRead.Add(1)
		result.device.ReadAsync(result.incomingData, nil, 0, 0)
		result.asyncRead.Done()
	}()

	return &result, nil
}

// Dongle represents the RTL-SDR dongle.
type Dongle struct {
	device    *rtl.Context
	buffer    bytes.Buffer
	asyncRead *sync.WaitGroup
	lastInput time.Time
}

// Read samples from the dongle.
func (d *Dongle) Read(p []byte) (n int, err error) {
	for d.buffer.Len() < len(p) {
		time.Sleep(1)
	}

	return d.buffer.Read(p)

}

// Close the dongle.
func (d *Dongle) Close() error {
	d.device.CancelAsync()
	d.asyncRead.Wait()
	return d.device.Close()
}

func (d *Dongle) incomingData(data []byte) {
	now := time.Now()
	d.lastInput = now
	_, err := d.buffer.Write(data)
	if err != nil {
		log.Print("Writing incoming data to buffer failed", err)
	}
}

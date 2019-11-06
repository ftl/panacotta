package vfo

import (
	"context"
	"fmt"
	"log"
	"net"
	"strconv"
	"sync"
	"time"

	"github.com/ftl/rigproxy/pkg/protocol"
	"github.com/pkg/errors"

	"github.com/ftl/panacotta/core"
)

// Open a connection to a hamlib VFO at the given network address. If address is empty, localhost:4534 is used.
func Open(address string) (*VFO, error) {
	if address == "" {
		address = "localhost:4532"
	}

	result := VFO{
		address:         address,
		trxTimeout:      100 * time.Millisecond,
		pollingInterval: 500 * time.Millisecond,
		setFrequency:    make(chan core.Frequency, 1),
		stateLock:       new(sync.RWMutex),
		data:            make(chan core.VFO, 1),
	}

	err := result.reconnect()
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// VFO type.
type VFO struct {
	address                   string
	trx                       *protocol.Transceiver
	trxTimeout                time.Duration
	pollingInterval           time.Duration
	setFrequency              chan core.Frequency
	currentFrequency          core.Frequency
	currentMode               string
	currentBandwidth          core.Frequency
	stateLock                 *sync.RWMutex
	frequencyChangedCallbacks []FrequencyChanged
	modeChangedCallbacks      []ModeChanged

	data chan core.VFO
}

// FrequencyChanged is called on frequency changes.
type FrequencyChanged func(f core.Frequency)

// ModeChanged is called on mode changes.
type ModeChanged func(mode string, bandwidth core.Frequency)

// Run the VFO.
func (v *VFO) Run(stop chan struct{}, wait *sync.WaitGroup) {
	wait.Add(1)
	go func() {
		defer wait.Done()
		defer v.shutdown()

		for {
			var err error
			select {
			case <-time.After(v.pollingInterval):
				err = v.pollFrequency()
				if err == nil {
					err = v.pollMode()
				}

			case f := <-v.setFrequency:
				err = v.sendFrequency(f)

			case <-stop:
				return
			}
			if err != nil {
				log.Print(err)
			}
		}
	}()
}

func (v *VFO) reconnect() error {
	if v.trx != nil {
		v.trx.Close()
	}

	out, err := net.Dial("tcp", v.address)
	if err != nil {
		return errors.Wrap(err, "cannot open VFO connection")
	}

	v.trx = protocol.NewTransceiver(out)
	v.trx.WhenDone(func() {
		out.Close()
	})

	return nil
}

func (v *VFO) shutdown() {
	v.trx.Close()
	log.Print("VFO shutdown")
}

func (v *VFO) pollFrequency() error {
	ctx, _ := context.WithTimeout(context.Background(), v.trxTimeout)
	request := protocol.Request{Command: protocol.ShortCommand("f")}
	response, err := v.trx.Send(ctx, request)
	if err != nil {
		log.Print("polling frequency failed: ", err)
		return err
	}

	if len(response.Data) < 1 {
		log.Printf("empty response %v", response)
		return errors.New("empty response")
	}

	f, err := hamlibToF(response.Data[0])
	if err != nil {
		log.Printf("wrong frequency format %s: %v", response.Data[0], err)
		return err
	}

	if v.updateCurrentFrequency(f) {
		v.data <- core.VFO{Frequency: f, Mode: v.currentMode, FilterWidth: v.currentBandwidth}

		for _, frequencyChanged := range v.frequencyChangedCallbacks {
			frequencyChanged(f)
		}
	}
	return nil
}

func (v *VFO) updateCurrentFrequency(f core.Frequency) bool {
	v.stateLock.Lock()
	defer v.stateLock.Unlock()
	if int(f) == int(v.currentFrequency) {
		return false
	}

	v.currentFrequency = f
	return true
}

func (v *VFO) pollMode() error {
	ctx, _ := context.WithTimeout(context.Background(), v.trxTimeout)
	request := protocol.Request{Command: protocol.ShortCommand("m")}
	response, err := v.trx.Send(ctx, request)
	if err != nil {
		log.Print("polling mode failed: ", err)
		return err
	}

	if len(response.Data) < 2 {
		log.Printf("empty response %v", response)
		return errors.New("empty response")
	}

	mode := response.Data[0]

	bandwidth, err := hamlibToF(response.Data[1])
	if err != nil {
		log.Printf("wrong frequency format %s: %v", response.Data[0], err)
		return err
	}

	if v.updateCurrentMode(mode, bandwidth) {
		v.data <- core.VFO{Frequency: v.currentFrequency, Mode: mode, FilterWidth: bandwidth}

		for _, modeChanged := range v.modeChangedCallbacks {
			modeChanged(mode, bandwidth)
		}
	}
	return nil
}

func (v *VFO) updateCurrentMode(mode string, bandwidth core.Frequency) bool {
	v.stateLock.Lock()
	defer v.stateLock.Unlock()
	if mode == v.currentMode && int(bandwidth) == int(v.currentBandwidth) {
		return false
	}

	v.currentMode = mode
	v.currentBandwidth = bandwidth
	return true
}

func (v *VFO) sendFrequency(f core.Frequency) error {
	ctx, _ := context.WithTimeout(context.Background(), v.trxTimeout)
	request := protocol.Request{Command: protocol.ShortCommand("F"), Args: []string{fToHamlib(f)}}
	_, err := v.trx.Send(ctx, request)
	if err != nil {
		log.Print("Sending frequency failed: ", err)
		return err
	}

	if v.updateCurrentFrequency(f) {
		for _, frequencyChanged := range v.frequencyChangedCallbacks {
			frequencyChanged(f)
		}
	}
	return nil
}

// Data of this VFO.
func (v *VFO) Data() <-chan core.VFO {
	return v.data
}

// TuneBy the given frequency delta.
func (v *VFO) TuneBy(Δf core.Frequency) {
	v.setFrequency <- v.CurrentFrequency() + Δf
}

// TuneTo the given frequency.
func (v *VFO) TuneTo(f core.Frequency) {
	v.setFrequency <- f
}

// SetFrequency sets the given frequency on the VFO.
func (v *VFO) SetFrequency(f core.Frequency) {
	v.TuneTo(f)
}

// MoveFrequency moves the VFO frequncy by the given delta.
func (v *VFO) MoveFrequency(Δf core.Frequency) {
	v.TuneBy(Δf)
}

// CurrentFrequency returns the current frequency of the VFO.
func (v *VFO) CurrentFrequency() core.Frequency {
	v.stateLock.RLock()
	defer v.stateLock.RUnlock()
	return v.currentFrequency
}

// OnFrequencyChange registers the given callback to be notified when the current frequency changes.
func (v *VFO) OnFrequencyChange(f FrequencyChanged) {
	v.frequencyChangedCallbacks = append(v.frequencyChangedCallbacks, f)
}

// CurrentMode returns the current mode and the current bandwidth of the VFO.
func (v *VFO) CurrentMode() (string, core.Frequency) {
	v.stateLock.RLock()
	defer v.stateLock.RUnlock()
	return v.currentMode, v.currentBandwidth
}

// OnModeChange registers the given callback to be notified when the current mode or bandwidth changes.
func (v *VFO) OnModeChange(m ModeChanged) {
	v.modeChangedCallbacks = append(v.modeChangedCallbacks, m)
}

func fToHamlib(f core.Frequency) string {
	return fmt.Sprintf("%d", int(f))
}

func hamlibToF(s string) (core.Frequency, error) {
	f, err := strconv.Atoi(s)
	if err != nil {
		return 0, nil
	}
	return core.Frequency(f), nil
}

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
	out, err := net.Dial("tcp", address)
	if err != nil {
		return nil, errors.Wrap(err, "cannot open VFO connection")
	}

	trx := protocol.NewTransceiver(out)
	trx.WhenDone(func() {
		out.Close()
	})

	result := VFO{
		trx:             trx,
		pollingInterval: 500 * time.Millisecond,
		setFrequency:    make(chan core.Frequency, 10),
		frequencyLock:   new(sync.RWMutex),
	}
	return &result, nil
}

// VFO type.
type VFO struct {
	trx                       *protocol.Transceiver
	pollingInterval           time.Duration
	setFrequency              chan core.Frequency
	currentFrequency          core.Frequency
	frequencyLock             *sync.RWMutex
	frequencyChangedCallbacks []FrequencyChanged
}

// FrequencyChanged is called on frequency changes.
type FrequencyChanged func(f core.Frequency)

// Run the VFO.
func (v *VFO) Run(stop chan struct{}, wait *sync.WaitGroup) {
	wait.Add(1)
	go func() {
		defer wait.Done()
		defer v.shutdown()

		for {
			select {
			case <-time.After(v.pollingInterval):
				v.pollFrequency()

			case f := <-v.setFrequency:
				v.sendFrequency(f)

			case <-stop:
				return
			}
		}
	}()
}

func (v *VFO) shutdown() {
	v.trx.Close()
	log.Print("VFO shutdown")
}

func (v *VFO) pollFrequency() {
	request := protocol.Request{Command: protocol.ShortCommand("f")}
	response, err := v.trx.Send(context.Background(), request)
	if err != nil {
		log.Print("Polling frequency failed: ", err)
		return
	}

	f, err := hamlibToF(response.Data[0])
	if err != nil {
		log.Printf("Wrong frequency format %s: %v", response.Data[0], err)
		return
	}

	if v.updateCurrentFrequency(f) {
		for _, frequencyChanged := range v.frequencyChangedCallbacks {
			frequencyChanged(f)
		}
	}
}

func (v *VFO) updateCurrentFrequency(f core.Frequency) bool {
	v.frequencyLock.Lock()
	defer v.frequencyLock.Unlock()
	if int(f) == int(v.currentFrequency) {
		return false
	}

	v.currentFrequency = f
	return true
}

func (v *VFO) sendFrequency(f core.Frequency) {
	request := protocol.Request{Command: protocol.ShortCommand("F"), Args: []string{fToHamlib(f)}}
	_, err := v.trx.Send(context.Background(), request)
	if err != nil {
		log.Print("Sending frequency failed: ", err)
	}
}

// SetFrequency sets the given frequency on the VFO.
func (v *VFO) SetFrequency(f core.Frequency) {
	v.setFrequency <- f
}

// CurrentFrequency returns the current frequency of the VFO.
func (v *VFO) CurrentFrequency() core.Frequency {
	v.frequencyLock.RLock()
	defer v.frequencyLock.RUnlock()
	return v.currentFrequency
}

// OnFrequencyChange registers the given callback to be notified if the current frequency changes.
func (v *VFO) OnFrequencyChange(f FrequencyChanged) {
	v.frequencyChangedCallbacks = append(v.frequencyChangedCallbacks, f)
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

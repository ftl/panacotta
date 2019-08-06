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
		setFrequency:    make(chan core.Frequency, 10),
		frequencyLock:   new(sync.RWMutex),
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
			var err error
			select {
			case <-time.After(v.pollingInterval):
				err = v.pollFrequency()

			case f := <-v.setFrequency:
				err = v.sendFrequency(f)

			case <-stop:
				return
			}
			if err == nil {
				continue
			}

			time.Sleep(500 * time.Millisecond)
			err = v.reconnect()
			if err != nil {
				log.Print("cannot reconnect to hamlib server: ", err)
				return
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
		for _, frequencyChanged := range v.frequencyChangedCallbacks {
			frequencyChanged(f)
		}
	}
	return nil
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

// SetFrequency sets the given frequency on the VFO.
func (v *VFO) SetFrequency(f core.Frequency) {
	v.setFrequency <- f
}

// MoveFrequency moves the VFO frequncy by the given delta.
func (v *VFO) MoveFrequency(delta core.Frequency) {
	v.setFrequency <- v.CurrentFrequency() + delta
}

// CurrentFrequency returns the current frequency of the VFO.
func (v *VFO) CurrentFrequency() core.Frequency {
	v.frequencyLock.RLock()
	defer v.frequencyLock.RUnlock()
	return v.currentFrequency
}

// OnFrequencyChange registers the given callback to be notified when the current frequency changes.
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

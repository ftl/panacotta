package vfo

import (
	"context"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/ftl/rigproxy/pkg/protocol"
	"github.com/pkg/errors"

	"github.com/ftl/panacotta/core"
)

// Open a connection to a hamlib VFO at the given network address. If address is empty, localhost:4532 is used.
func Open(address string) (*VFO, error) {
	if address == "" {
		address = "localhost:4532"
	}

	result := VFO{
		address:         address,
		trxTimeout:      100 * time.Millisecond,
		pollingInterval: 500 * time.Millisecond,
		tuneTo:          make(chan core.Frequency, 1),
		tuneBy:          make(chan core.Frequency, 1),
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
	address         string
	trx             *protocol.Transceiver
	trxTimeout      time.Duration
	pollingInterval time.Duration
	tuneTo          chan core.Frequency
	tuneBy          chan core.Frequency
	state           core.VFO
	stateLock       *sync.RWMutex

	data chan core.VFO
}

// Run the VFO.
func (v *VFO) Run(stop chan struct{}) {
	go func() {
		defer v.shutdown()

		for {
			var err error
			select {
			case <-time.After(v.pollingInterval):
				err = v.pollState()

			case f := <-v.tuneTo:
				err = v.sendFrequency(f)

			case Δf := <-v.tuneBy:
				currentFrequency := v.CurrentFrequency()
				err = v.sendFrequency(currentFrequency + Δf)

			case <-stop:
				return
			}
			if err != nil {
				log.Print(err)
				err = nil
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

func (v *VFO) pollState() error {
	commands := []string{"v", "f", "m"}
	handlers := []func(protocol.Response) error{
		v.handleNameResponse,
		v.handleFrequencyResponse,
		v.handleModeResponse,
	}
	for i, command := range commands {
		err := v.poll(handlers[i], command)
		if err != nil {
			return err
		}
	}
	return nil
}

func (v *VFO) poll(handleResponse func(protocol.Response) error, shortCommand string, args ...string) error {
	ctx, _ := context.WithTimeout(context.Background(), v.trxTimeout)
	request := protocol.Request{Command: protocol.ShortCommand(shortCommand), Args: args}
	response, err := v.trx.Send(ctx, request)
	if err != nil {
		log.Printf("polling %s failed: %v", shortCommand, err)
		return err
	}

	return handleResponse(response)
}

func (v *VFO) handleFrequencyResponse(response protocol.Response) error {
	if len(response.Data) < 1 {
		log.Printf("empty response %v", response)
		return errors.New("empty response")
	}

	f, err := hamlibToF(response.Data[0])
	if err != nil {
		log.Printf("wrong frequency format %s: %v", response.Data[0], err)
		return err
	}

	v.updateState(setFrequency(f))

	return nil
}

func setFrequency(f core.Frequency) func(*core.VFO) {
	return func(state *core.VFO) {
		state.Frequency = f
	}
}

func (v *VFO) handleModeResponse(response protocol.Response) error {
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

	v.updateState(setMode(mode, bandwidth))
	return nil
}

func setMode(mode string, bandwidth core.Frequency) func(*core.VFO) {
	return func(state *core.VFO) {
		state.Mode = mode
		state.FilterWidth = bandwidth
	}
}

func (v *VFO) handleNameResponse(response protocol.Response) error {
	if len(response.Data) < 1 {
		log.Printf("empty response %v", response)
		return errors.New("empty response")
	}

	name := response.Data[0]
	if strings.HasPrefix(name, "VFO") && len(name) > 3 {
		name = name[3:len(name)]
	}
	v.updateState(setName(name))

	return nil
}

func setName(name string) func(*core.VFO) {
	return func(state *core.VFO) {
		state.Name = name
	}
}

func (v *VFO) updateState(updater func(*core.VFO)) {
	v.stateLock.Lock()
	defer v.stateLock.Unlock()

	oldState := v.state
	newState := v.state
	updater(&newState)

	if oldState != newState {
		v.state = newState
		v.data <- newState
	}
}

func (v *VFO) sendFrequency(f core.Frequency) error {
	ctx, _ := context.WithTimeout(context.Background(), v.trxTimeout)
	request := protocol.Request{Command: protocol.ShortCommand("F"), Args: []string{fToHamlib(f)}}
	_, err := v.trx.Send(ctx, request)
	if err != nil {
		log.Print("Sending frequency failed: ", err)
		return err
	}

	v.updateState(setFrequency(f))
	return nil
}

// Data of this VFO.
func (v *VFO) Data() <-chan core.VFO {
	return v.data
}

// TuneTo the given frequency.
func (v *VFO) TuneTo(f core.Frequency) {
	select {
	case v.tuneTo <- f:
	default:
		log.Print("VFO.TuneTo hangs")
	}
}

// TuneBy the given frequency delta.
func (v *VFO) TuneBy(Δf core.Frequency) {
	select {
	case v.tuneBy <- Δf:
	default:
		log.Print("VFO.TuneBy hangs")
	}
}

// CurrentFrequency returns the current frequency of the VFO.
func (v *VFO) CurrentFrequency() core.Frequency {
	v.stateLock.RLock()
	defer v.stateLock.RUnlock()
	return v.state.Frequency
}

// CurrentMode returns the current mode and the current bandwidth of the VFO.
func (v *VFO) CurrentMode() (string, core.Frequency) {
	v.stateLock.RLock()
	defer v.stateLock.RUnlock()
	return v.state.Mode, v.state.FilterWidth
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

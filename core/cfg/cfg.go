package cfg

import (
	"github.com/ftl/hamradio/cfg"

	"github.com/ftl/panacotta/core"
)

const (
	testmode            cfg.Key = "panacotta.testmode"
	frequencyCorrection cfg.Key = "panacotta.frequencyCorrection"
	vfoHost             cfg.Key = "panacotta.vfoHost"
	fftPerSecond        cfg.Key = "panacotta.fftPerSecond"
	dynamicRangeFrom    cfg.Key = "panacotta.dynamicRange.from"
	dynamicRangeTo      cfg.Key = "panacotta.dynamicRange.to"
)

func Load() (core.Configuration, error) {
	configuration, err := cfg.LoadDefault()
	if err != nil {
		return core.Configuration{}, err
	}

	result := core.Configuration{
		Testmode:            configuration.Get(testmode, false).(bool),
		FrequencyCorrection: int(configuration.Get(frequencyCorrection, 0.0).(float64)),
		VFOHost:             configuration.Get(vfoHost, "").(string),
		FFTPerSecond:        int(configuration.Get(fftPerSecond, 25.0).(float64)),
		DynamicRange: core.DBRange{
			From: core.DB(configuration.Get(dynamicRangeFrom, -105.0).(float64)),
			To:   core.DB(configuration.Get(dynamicRangeTo, 15.0).(float64)),
		}.Normalized(),
	}

	return result, nil
}

func Static() core.Configuration {
	return core.Configuration{}
}

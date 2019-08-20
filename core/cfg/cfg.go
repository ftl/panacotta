package cfg

import (
	"github.com/ftl/hamradio/cfg"

	"github.com/ftl/panacotta/core"
)

const (
	testmode            cfg.Key = "panacotta.testmode"
	frequencyCorrection cfg.Key = "panacotta.frequencyCorrection"
	vfoHost             cfg.Key = "panacotta.vfoHost"
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
	}

	return result, nil
}

func Static() core.Configuration {
	return core.Configuration{}
}

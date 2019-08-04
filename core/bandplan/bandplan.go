package bandplan

import "github.com/ftl/panacotta/core"

// Band represents a frequency band.
type Band struct {
	core.FrequencyRange
	Name BandName
}

// Contains indicates if the band contains the given frequency.
func (b *Band) Contains(f core.Frequency) bool {
	return f >= b.From && f <= b.To
}

// UnknownBand is the unknown band that contains no frequency.
var UnknownBand = Band{Name: BandUnknown}

// BandName is the name of a frequency band.
type BandName string

// All HF bands.
const (
	BandUnknown BandName = "Unknown"
	Band160m    BandName = "160m"
	Band80m     BandName = "80m"
	Band60m     BandName = "60m"
	Band40m     BandName = "40m"
	Band30m     BandName = "30m"
	Band20m     BandName = "20m"
	Band17m     BandName = "17m"
	Band15m     BandName = "15m"
	Band12m     BandName = "12m"
	Band10m     BandName = "10m"
)

// Mode type
type Mode string

// All modes.
const (
	ModeCW      Mode = "CW"
	ModeSSB     Mode = "SSB"
	ModeFM      Mode = "FM"
	ModeDigital Mode = "Digital"
	ModeBeacon  Mode = "Beacon"
	ModeContest Mode = "Contest"
)

// Bandplan type.
type Bandplan map[BandName]Band

// ByFrequency returns the band for the matching frequency.
func (p Bandplan) ByFrequency(f core.Frequency) Band {
	for _, b := range p {
		if b.Contains(f) {
			return b
		}
	}
	return UnknownBand
}

// IARURegion1 is the bandplan for IARU Region 1
var IARURegion1 = Bandplan{
	Band160m: Band{
		Name: Band160m,
		FrequencyRange: core.FrequencyRange{
			From: 1810000.0,
			To:   2000000.0,
		},
	},
	Band80m: Band{
		Name: Band80m,
		FrequencyRange: core.FrequencyRange{
			From: 3500000.0,
			To:   3800000.0,
		},
	},
	Band60m: Band{
		Name: Band60m,
		FrequencyRange: core.FrequencyRange{
			From: 5351500.0,
			To:   5366500.0,
		},
	},
	Band40m: Band{
		Name: Band40m,
		FrequencyRange: core.FrequencyRange{
			From: 7000000.0,
			To:   7200000.0,
		},
	},
	Band30m: Band{
		Name: Band30m,
		FrequencyRange: core.FrequencyRange{
			From: 10100000.0,
			To:   10150000.0,
		},
	},
	Band20m: Band{
		Name: Band20m,
		FrequencyRange: core.FrequencyRange{
			From: 14000000.0,
			To:   14350000.0,
		},
	},
	Band17m: Band{
		Name: Band17m,
		FrequencyRange: core.FrequencyRange{
			From: 18068000.0,
			To:   18168000.0,
		},
	},
	Band15m: Band{
		Name: Band15m,
		FrequencyRange: core.FrequencyRange{
			From: 21000000.0,
			To:   21450000.0,
		},
	},
	Band12m: Band{
		Name: Band12m,
		FrequencyRange: core.FrequencyRange{
			From: 24890000.0,
			To:   24990000.0,
		},
	},
	Band10m: Band{
		Name: Band10m,
		FrequencyRange: core.FrequencyRange{
			From: 28000000.0,
			To:   29700000.0,
		},
	},
}

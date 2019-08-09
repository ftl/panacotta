package bandplan

import "github.com/ftl/panacotta/core"

// Band represents a frequency band.
type Band struct {
	core.FrequencyRange
	Name     BandName
	Portions []Portion
}

// Portion is a part of a frequency band with a preferred mode.
type Portion struct {
	core.FrequencyRange
	Mode Mode
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
	ModePhone   Mode = "Phone"
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
		Portions: []Portion{
			{
				Mode: ModeCW,
				FrequencyRange: core.FrequencyRange{
					From: 1810000.0,
					To:   1838000.0,
				},
			},
			{
				Mode: ModeDigital,
				FrequencyRange: core.FrequencyRange{
					From: 1838000.0,
					To:   1843000.0,
				},
			},
			{
				Mode: ModePhone,
				FrequencyRange: core.FrequencyRange{
					From: 1843000.0,
					To:   2000000.0,
				},
			},
		},
	},
	Band80m: Band{
		Name: Band80m,
		FrequencyRange: core.FrequencyRange{
			From: 3500000.0,
			To:   3800000.0,
		},
		Portions: []Portion{
			{
				Mode: ModeCW,
				FrequencyRange: core.FrequencyRange{
					From: 3500000.0,
					To:   3570000.0,
				},
			},
			{
				Mode: ModeDigital,
				FrequencyRange: core.FrequencyRange{
					From: 3570000.0,
					To:   3620000.0,
				},
			},
			{
				Mode: ModePhone,
				FrequencyRange: core.FrequencyRange{
					From: 3620000.0,
					To:   3800000.0,
				},
			},
			{
				Mode: ModeContest,
				FrequencyRange: core.FrequencyRange{
					From: 3510000.0,
					To:   3569000.0,
				},
			},
			{
				Mode: ModeContest,
				FrequencyRange: core.FrequencyRange{
					From: 3600000.0,
					To:   3650000.0,
				},
			},
			{
				Mode: ModeContest,
				FrequencyRange: core.FrequencyRange{
					From: 3700000.0,
					To:   3800000.0,
				},
			},
		},
	},
	Band60m: Band{
		Name: Band60m,
		FrequencyRange: core.FrequencyRange{
			From: 5351500.0,
			To:   5366500.0,
		},
		Portions: []Portion{
			{
				Mode: ModeCW,
				FrequencyRange: core.FrequencyRange{
					From: 5351500.0,
					To:   5354000.0,
				},
			},
			{
				Mode: ModePhone,
				FrequencyRange: core.FrequencyRange{
					From: 5354000.0,
					To:   5366000.0,
				},
			},
			{
				Mode: ModeDigital,
				FrequencyRange: core.FrequencyRange{
					From: 5366000.0,
					To:   5366500.0,
				},
			},
		},
	},
	Band40m: Band{
		Name: Band40m,
		FrequencyRange: core.FrequencyRange{
			From: 7000000.0,
			To:   7200000.0,
		},
		Portions: []Portion{
			{
				Mode: ModeCW,
				FrequencyRange: core.FrequencyRange{
					From: 7000000.0,
					To:   7040000.0,
				},
			},
			{
				Mode: ModeDigital,
				FrequencyRange: core.FrequencyRange{
					From: 7040000.0,
					To:   7053000.0,
				},
			},
			{
				Mode: ModePhone,
				FrequencyRange: core.FrequencyRange{
					From: 7053000.0,
					To:   7200000.0,
				},
			},
			{
				Mode: ModeContest,
				FrequencyRange: core.FrequencyRange{
					From: 7000000.0,
					To:   7040000.0,
				},
			},
			{
				Mode: ModeContest,
				FrequencyRange: core.FrequencyRange{
					From: 7130000.0,
					To:   7200000.0,
				},
			},
		},
	},
	Band30m: Band{
		Name: Band30m,
		FrequencyRange: core.FrequencyRange{
			From: 10100000.0,
			To:   10150000.0,
		},
		Portions: []Portion{
			{
				Mode: ModeCW,
				FrequencyRange: core.FrequencyRange{
					From: 10100000.0,
					To:   10130000.0,
				},
			},
			{
				Mode: ModeDigital,
				FrequencyRange: core.FrequencyRange{
					From: 10130000.0,
					To:   10150000.0,
				},
			},
		},
	},
	Band20m: Band{
		Name: Band20m,
		FrequencyRange: core.FrequencyRange{
			From: 14000000.0,
			To:   14350000.0,
		},
		Portions: []Portion{
			{
				Mode: ModeCW,
				FrequencyRange: core.FrequencyRange{
					From: 14000000.0,
					To:   14070000.0,
				},
			},
			{
				Mode: ModeDigital,
				FrequencyRange: core.FrequencyRange{
					From: 14070000.0,
					To:   14099000.0,
				},
			},
			{
				Mode: ModeBeacon,
				FrequencyRange: core.FrequencyRange{
					From: 14099000.0,
					To:   14101000.0,
				},
			},
			{
				Mode: ModeDigital,
				FrequencyRange: core.FrequencyRange{
					From: 14101000.0,
					To:   14112000.0,
				},
			},
			{
				Mode: ModePhone,
				FrequencyRange: core.FrequencyRange{
					From: 14112000.0,
					To:   14350000.0,
				},
			},
			{
				Mode: ModeContest,
				FrequencyRange: core.FrequencyRange{
					From: 14000000.0,
					To:   14060000.0,
				},
			},
			{
				Mode: ModeContest,
				FrequencyRange: core.FrequencyRange{
					From: 14125000.0,
					To:   14300000.0,
				},
			},
		},
	},
	Band17m: Band{
		Name: Band17m,
		FrequencyRange: core.FrequencyRange{
			From: 18068000.0,
			To:   18168000.0,
		},
		Portions: []Portion{
			{
				Mode: ModeCW,
				FrequencyRange: core.FrequencyRange{
					From: 18068000.0,
					To:   18095000.0,
				},
			},
			{
				Mode: ModeDigital,
				FrequencyRange: core.FrequencyRange{
					From: 18095000.0,
					To:   18109000.0,
				},
			},
			{
				Mode: ModeBeacon,
				FrequencyRange: core.FrequencyRange{
					From: 18109000.0,
					To:   18111000.0,
				},
			},
			{
				Mode: ModeDigital,
				FrequencyRange: core.FrequencyRange{
					From: 18111000.0,
					To:   18120000.0,
				},
			},
			{
				Mode: ModePhone,
				FrequencyRange: core.FrequencyRange{
					From: 18120000.0,
					To:   18168000.0,
				},
			},
		},
	},
	Band15m: Band{
		Name: Band15m,
		FrequencyRange: core.FrequencyRange{
			From: 21000000.0,
			To:   21450000.0,
		},
		Portions: []Portion{
			{
				Mode: ModeCW,
				FrequencyRange: core.FrequencyRange{
					From: 21000000.0,
					To:   21070000.0,
				},
			},
			{
				Mode: ModeDigital,
				FrequencyRange: core.FrequencyRange{
					From: 21070000.0,
					To:   21149000.0,
				},
			},
			{
				Mode: ModeBeacon,
				FrequencyRange: core.FrequencyRange{
					From: 21149000.0,
					To:   21151000.0,
				},
			},
			{
				Mode: ModePhone,
				FrequencyRange: core.FrequencyRange{
					From: 21151000.0,
					To:   21450000.0,
				},
			},
			{
				Mode: ModeContest,
				FrequencyRange: core.FrequencyRange{
					From: 21000000.0,
					To:   21070000.0,
				},
			},
			{
				Mode: ModeContest,
				FrequencyRange: core.FrequencyRange{
					From: 21151000.0,
					To:   21450000.0,
				},
			},
		},
	},
	Band12m: Band{
		Name: Band12m,
		FrequencyRange: core.FrequencyRange{
			From: 24890000.0,
			To:   24990000.0,
		},
		Portions: []Portion{
			{
				Mode: ModeCW,
				FrequencyRange: core.FrequencyRange{
					From: 24890000.0,
					To:   24915000.0,
				},
			},
			{
				Mode: ModeDigital,
				FrequencyRange: core.FrequencyRange{
					From: 24915000.0,
					To:   24929000.0,
				},
			},
			{
				Mode: ModeBeacon,
				FrequencyRange: core.FrequencyRange{
					From: 24929000.0,
					To:   24931000.0,
				},
			},
			{
				Mode: ModeDigital,
				FrequencyRange: core.FrequencyRange{
					From: 24931000.0,
					To:   24940000.0,
				},
			},
			{
				Mode: ModePhone,
				FrequencyRange: core.FrequencyRange{
					From: 24940000.0,
					To:   24990000.0,
				},
			},
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

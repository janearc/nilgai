package band

import (
	"fmt"
	"strings"

	"github.com/janearc/libtheme-css/primitives/si"
	"github.com/janearc/libtheme-css/spaces/radiation"
	"github.com/janearc/nilgai"
)

// Quantity is a value written with its unit, as a paper writes it, such
// as "1 keV" or "21 cm". In data/profiles.json it is an object with a
// "value" and a "unit".
type Quantity struct {
	// Value is the number the paper gives.
	Value float64 `json:"value"`
	// Unit is the unit the paper gives it in, such as "keV" or "cm".
	Unit string `json:"unit"`
}

// prefixes are the si prefixes nilgai reads in front of a unit, by their
// symbol, with u standing for micro. The functions are libtheme's. A unit
// with no prefix is not in the map; it is used as it is.
var prefixes = map[string]func(float64) float64{
	"k": si.Kilo, "M": si.Mega, "G": si.Giga, "T": si.Tera, "P": si.Peta,
	"c": si.Centi, "m": si.Milli, "u": si.Micro, "n": si.Nano,
}

// bases are the units a band's end is written in, and how one becomes a
// wavelength in nm: a photon's energy in eV, a frequency in Hz, or a
// wavelength in metres. The conversions are libtheme's.
var bases = map[string]func(float64) float64{
	"eV": radiation.FromEnergy,
	"Hz": radiation.HertzToNm,
	"m":  radiation.MetresToNm,
}

// Nanometres is the quantity brought to a wavelength in nm: an energy in
// eV, a frequency in Hz or a length in m, each with an si prefix or none.
//
// Nanometres reads these prefixes, and refuses any other:
//
//   - k, M, G, T and P: kilo to peta
//   - c: centi
//   - m: milli
//   - u: micro
//   - n: nano
//
// Nanometres refuses a value that is not finite and above zero, and a
// wavelength that comes out not finite and above zero. On a refusal it
// returns zero with the error.
func (quantity Quantity) Nanometres() (float64, error) {
	if !nilgai.Positive(quantity.Value) {
		return none, fmt.Errorf("%v %s must be finite and above zero",
			quantity.Value, quantity.Unit)
	}
	for base, toNanometres := range bases {
		prefix, found := strings.CutSuffix(quantity.Unit, base)
		if !found {
			continue
		}
		scaled := quantity.Value
		if prefix != "" {
			scale, known := prefixes[prefix]
			if !known {
				continue
			}
			scaled = scale(quantity.Value)
		}
		nm := toNanometres(scaled)
		if !nilgai.Positive(nm) {
			return none, fmt.Errorf("%v %s comes to %v nm, "+
				"which must be finite and above zero",
				quantity.Value, quantity.Unit, nm)
		}
		return nm, nil
	}
	return none, fmt.Errorf("%q is not a unit nilgai reads", quantity.Unit)
}

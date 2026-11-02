package band

import (
	"fmt"
	"strings"

	"github.com/janearc/libtheme-css/primitives/si"
	"github.com/janearc/libtheme-css/spaces/radiation"
	"github.com/janearc/nilgai"
)

// Quantity pairs a value with a published unit, such as 1 keV or 21 cm. In
// data/profiles.json it has keys "value" and "unit".
type Quantity struct {
	// Value is the number the paper gives.
	Value float64 `json:"value"`
	// Unit is the unit the paper gives it in, such as "keV" or "cm".
	Unit string `json:"unit"`
}

// prefixes maps supported SI prefix symbols to libtheme functions, using u for
// micro. Units without a prefix use their base conversion directly.
var prefixes = map[string]func(float64) float64{
	"k": si.Kilo, "M": si.Mega, "G": si.Giga, "T": si.Tera, "P": si.Peta,
	"c": si.Centi, "m": si.Milli, "u": si.Micro, "n": si.Nano,
}

// bases maps eV, Hz, and m to libtheme’s wavelength conversions. Each
// conversion returns nanometres.
var bases = map[string]func(float64) float64{
	"eV": radiation.FromEnergy,
	"Hz": radiation.HertzToNm,
	"m":  radiation.MetresToNm,
}

// Nanometres converts a photon energy in eV, frequency in Hz, or length in m to
// wavelength. Supported prefixes are k, M, G, T, P, c, m, u, and n; unprefixed
// units are also accepted.
//
// The input and resulting wavelength must be finite and positive. Unknown units
// or invalid values return zero and an error.
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

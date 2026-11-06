package band

import (
	"fmt"

	"github.com/janearc/celestial/radiation"
)

// Quantity is a value written with its unit, as a paper writes it, such
// as "1 keV" or "84 GHz". In data/profiles.json it is an object with a
// "value" and a "unit".
type Quantity struct {
	// Value is the number the paper gives.
	Value float64 `json:"value"`
	// Unit is the unit the paper gives it in, such as "keV" or "GHz".
	Unit string `json:"unit"`
}

// units are the units nilgai's data writes a band's ends in, each with
// celestial's named function that brings it to nanometres. A unit comes
// here when the data uses it.
var units = map[string]func(float64) (*float64, error){
	"keV": radiation.KeVToNm,
	"MeV": radiation.MeVToNm,
	"GeV": radiation.GeVToNm,
	"TeV": radiation.TeVToNm,
	"MHz": radiation.MHzToNm,
	"GHz": radiation.GHzToNm,
	"um":  radiation.MicronsToNm,
}

// Nanometres is the quantity brought to a wavelength in nm, by the named
// function for its unit. returns nil and error for a unit nilgai does
// not read, and for a value that gives no wavelength.
func (quantity Quantity) Nanometres() (*float64, error) {
	toNanometres, known := units[quantity.Unit]
	if !known {
		return nil, fmt.Errorf("%q is not a unit nilgai reads",
			quantity.Unit)
	}
	return toNanometres(quantity.Value)
}

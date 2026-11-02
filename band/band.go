// Package band maps catalogue filters onto libtheme’s spectral bands and
// supplies false-colour profiles for instruments.
//
// Filter edges and their citations live in data/filters.json, with flags
// identifying approximate bounds. data/profiles.json records instrument
// coverage in its published units and the visible range used to display it.
package band

import (
	"fmt"

	"github.com/janearc/libtheme-css/primitives/bands"
	"github.com/janearc/nilgai"
)

// none is the zero count.
const none = 0

// one is the increment for the next band.
const one = 1

// BandsBetween returns the first and last libtheme bands covered by a filter
// with wavelength bounds lo and hi in nanometres. Nilgai uses libtheme’s
// bands.Band directly.
//
// Both bands and filters include their low edge and exclude their high edge. If
// hi falls inside a band, that band is included. If hi is exactly on a band
// boundary, the band starting there is excluded.
//
// Bounds must be finite and positive, with lo < hi. Convert frequencies,
// energies, or other lengths to nanometres with libtheme’s radiation package
// before calling.
func BandsBetween(lo, hi float64) (first, last bands.Band, err error) {
	if !nilgai.Positive(lo) || !nilgai.Positive(hi) {
		return first, last, fmt.Errorf("edges %v nm and %v nm must be "+
			"finite and above zero", lo, hi)
	}
	if lo >= hi {
		return first, last, fmt.Errorf("low edge %v nm must be below "+
			"high edge %v nm", lo, hi)
	}
	first, last = bands.Of(lo), bands.Of(hi)
	if start, _ := last.Edges(); start == hi {
		last--
	}
	return first, last, nil
}

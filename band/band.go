// Package band maps catalogue filters and instruments onto libtheme's
// bands of the spectrum. It holds:
//   - the first and last libtheme band that a catalogue's filter covers
//   - the filters themselves
//   - the false-colour profile of each instrument
//
// data/filters.json lists the filters, with a citation for each and a flag
// saying whether its edges are approximate.
//
// data/profiles.json lists the profiles: each instrument's band in its own
// units, and the slice of visible light that band is drawn in.
//
// Papers give these numbers and nobody computes them, so they live in data
// and not in code.
package band

import (
	"fmt"

	"github.com/janearc/libtheme-css/primitives/bands"
	"github.com/janearc/nilgai"
)

// none is the count of nothing, zero.
const none = 0

// one is the count of a single thing.
const one = 1

// BandsBetween returns the first and last libtheme band that a filter
// covers, given the filter's edges as wavelengths in nm. nilgai has no
// band type of its own: a slice of the spectrum is libtheme's bands.Band,
// and a filter is the first and last of them.
//
// BandsBetween follows libtheme: a band includes its low edge but not its
// high edge, and so does a filter. BandsBetween's first band is the one lo
// falls in, and its last is the one just below hi.
//
// When hi lands inside a band, that band is the last. When hi lands
// exactly on a band's low edge, that band is left out, because none of
// the filter reaches into it.
//
// BandsBetween refuses:
//   - an edge that is not a finite number above zero
//   - a low edge that is not below the high edge
//
// Convert a frequency, an energy or a length to nanometres with
// libtheme's radiation package first.
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

package nilgai

import (
	"fmt"

	"github.com/janearc/libtheme-css/primitives/bands"
)

// BandsBetween is the run of libtheme bands a catalogue's filter covers,
// given the filter's edges as wavelengths in nm. nilgai has no band of
// its own: a slice of the spectrum is libtheme's bands.Band, and a
// filter is the first and last of them.
//
// libtheme's bands hold their low edge and not their high one, and so
// does a filter: first is the band lo falls in, and last is the band
// just below hi.
//
// When hi lands inside a band, that band is last; when hi lands exactly
// on a band's low edge, that band is left out, because none of the
// filter reaches into it.
//
// It refuses an edge that is not a finite number above zero, and edges
// that are not low then high. Bring a frequency, an energy or a length
// into nanometres with libtheme's radiation package first.
func BandsBetween(lo, hi float64) (first, last bands.Band, err error) {
	if !positive(lo) || !positive(hi) {
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

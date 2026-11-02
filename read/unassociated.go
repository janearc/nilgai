package read

import (
	"iter"
	"math"

	"github.com/janearc/nilgai"
)

// Unassociated filters a Fermi sequence to sources with a blank CLASS1 and |b|
// strictly below limit degrees from the Galactic plane. Row and input errors
// pass through unchanged.
//
// Classes unk (unknown nature) and bcu (blazar candidate of uncertain type) are
// associations for this selection. Sources from other catalogues or without a
// CLASS1 flag are excluded.
//
// Latitude comes from Position.GalacticLatitude. The boundary is exclusive; a
// nonpositive limit selects nothing.
func Unassociated(sources iter.Seq2[*nilgai.Source, error],
	limit float64) iter.Seq2[*nilgai.Source, error] {
	selecting := unassociation{sources: sources, limit: limit}
	return selecting.each
}

// unassociation holds one call of Unassociated: the sources it reads and
// how near the Galactic plane they must lie.
type unassociation struct {
	sources iter.Seq2[*nilgai.Source, error]
	limit   float64
}

// each yields what the sources yield, minus the sources that are not
// unassociated within the limit, until the sources end or the caller
// stops.
func (selecting unassociation) each(yield func(*nilgai.Source, error) bool) {
	for source, err := range selecting.sources {
		if err == nil && !unassociatedWithin(*source, selecting.limit) {
			continue
		}
		if !yield(source, err) {
			return
		}
	}
}

// unassociatedWithin says whether the source is a 4FGL source with a
// blank CLASS1 flag, and within limit degrees of the Galactic plane.
func unassociatedWithin(source nilgai.Source, limit float64) bool {
	if source.Catalogue() != nilgai.Fermi ||
		math.Abs(source.Position().GalacticLatitude()) >= limit {
		return false
	}
	for _, measurement := range source.Measurements() {
		for _, flag := range measurement.Flags() {
			if flag.Name() == fermiClassColumn {
				return flag.Value() == ""
			}
		}
	}
	return false
}

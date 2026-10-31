package read

import (
	"iter"
	"math"

	"github.com/janearc/nilgai"
)

// Unassociated yields what a Fermi read yields, but keeps only the sources
// that 4FGL has no association for and that lie within limit degrees of
// the Galactic plane. Rows the reader could not read come through as the
// errors they are, and are not dropped.
//
// Unassociated selects a 4FGL source whose CLASS1 is blank. The classes unk
// (unknown nature) and bcu (blazar candidate of uncertain type) are
// associations, so Unassociated does not select them.
//
// Unassociated never selects:
//   - a source of another catalogue
//   - a 4FGL source with no CLASS1 flag, because 4FGL has not said it is
//     unassociated
//
// Unassociated takes |b| from the source's position, and the limit is
// exclusive: a source at |b| equal to limit is not selected. A limit of
// zero or less selects nothing.
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

package nilgai

import (
	"errors"
	"fmt"
	"slices"
)

// Catalogue names one of the four inputs.
type Catalogue string

// The four catalogues nilgai reads.
const (
	MeerKAT Catalogue = "MeerKAT"
	Fermi   Catalogue = "Fermi-4FGL"
	Chandra Catalogue = "Chandra-CSC"
	NEOWISE Catalogue = "NEOWISE"
)

// valid says whether the catalogue is one of the four.
func (catalogue Catalogue) valid() bool {
	switch catalogue {
	case MeerKAT, Fermi, Chandra, NEOWISE:
		return true
	}
	return false
}

// Source is one catalogue row in the common record. It comes from
// NewSource, and a Source that exists has passed its checks.
type Source struct {
	catalogue    Catalogue
	id           string
	position     Position
	ellipse      ErrorEllipse
	epoch        When
	measurements []Measurement
}

// NewSource makes a Source. It refuses an unknown catalogue, no
// identifier, an ellipse that did not come from NewErrorEllipse, no
// measurements, or a measurement that did not come from a constructor.
//
// The identifier is the catalogue's own, verbatim. The position error
// is at one sigma. The epoch is the position's, when the catalogue
// gives one, and the zero When when it does not.
func NewSource(catalogue Catalogue, id string, position Position,
	ellipse ErrorEllipse, epoch When,
	measurements []Measurement) (source Source, err error) {
	if !catalogue.valid() {
		return source, fmt.Errorf("catalogue %q is not one of the four",
			string(catalogue))
	}
	if id == "" {
		return source, errors.New("source has no identifier")
	}
	if !ellipse.valid() {
		return source, errors.New("source has no error ellipse")
	}
	if len(measurements) == none {
		return source, errors.New("source has no measurements")
	}
	for place, each := range measurements {
		if !each.valid() {
			return source, fmt.Errorf(
				"measurement %d was not made by a "+
					"constructor", place)
		}
	}
	return Source{catalogue: catalogue, id: id, position: position,
		ellipse: ellipse, epoch: epoch,
		measurements: slices.Clone(measurements)}, nil
}

// Catalogue is the catalogue the row came from.
func (source Source) Catalogue() Catalogue { return source.catalogue }

// ID is the catalogue's own identifier, verbatim.
func (source Source) ID() string { return source.id }

// Position is the ICRS position in degrees.
func (source Source) Position() Position { return source.position }

// Error is the position error at one sigma.
func (source Source) Error() ErrorEllipse { return source.ellipse }

// Epoch is the epoch of the position, if the catalogue gave one.
func (source Source) Epoch() When { return source.epoch }

// Measurements are the source's measurements, one or more.
func (source Source) Measurements() []Measurement {
	return slices.Clone(source.measurements)
}

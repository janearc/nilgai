package nilgai

import (
	"errors"
	"fmt"
	"slices"
)

// Catalogue names one of the inputs.
type Catalogue string

// Catalogue's values are the catalogues nilgai reads. AllWISE is WISE's
// catalogue of sources, one row for each; Neowise is WISE's catalogue of
// single looks, which nilgai reads as light curves.
const (
	MeerKAT Catalogue = "MeerKAT"
	Fermi   Catalogue = "Fermi-4FGL"
	Chandra Catalogue = "Chandra-CSC"
	NEOWISE Catalogue = "NEOWISE"
	AllWISE Catalogue = "AllWISE"
)

// valid says whether nilgai reads the catalogue.
func (catalogue Catalogue) valid() bool {
	switch catalogue {
	case MeerKAT, Fermi, Chandra, NEOWISE, AllWISE:
		return true
	}
	return false
}

// Source is one catalogue row in the common record. Make one with
// NewSource. A Source that exists has passed NewSource's checks.
type Source struct {
	catalogue    Catalogue
	id           string
	position     Position
	ellipse      ErrorEllipse
	epoch        When
	measurements []Measurement
}

// NewSource returns a Source. It refuses:
//   - a catalogue nilgai does not read
//   - an empty identifier
//   - an ellipse that did not come from NewErrorEllipse
//   - no measurements
//   - a measurement that no constructor made
//
// id is the catalogue's own identifier, verbatim. The position error is
// at one sigma. The epoch is the position's, when the catalogue gives
// one, and the zero When when it does not. On a refusal the Source is
// nil.
func NewSource(catalogue Catalogue, id string, position Position,
	ellipse ErrorEllipse, epoch When,
	measurements []Measurement) (*Source, error) {
	if !catalogue.valid() {
		return nil, fmt.Errorf("catalogue %q is not one nilgai "+
			"reads", string(catalogue))
	}
	if id == "" {
		return nil, errors.New("source has no identifier")
	}
	if !ellipse.valid() {
		return nil, errors.New("source has no error ellipse")
	}
	if len(measurements) == none {
		return nil, errors.New("source has no measurements")
	}
	for place, each := range measurements {
		if !each.valid() {
			return nil, fmt.Errorf(
				"measurement %d was not made by a "+
					"constructor", place)
		}
	}
	return &Source{catalogue: catalogue, id: id, position: position,
		ellipse: ellipse, epoch: epoch,
		measurements: slices.Clone(measurements)}, nil
}

// Catalogue returns the catalogue the row came from.
func (source Source) Catalogue() Catalogue { return source.catalogue }

// ID returns the catalogue's own identifier, verbatim.
func (source Source) ID() string { return source.id }

// Position returns the ICRS position in degrees.
func (source Source) Position() Position { return source.position }

// Error returns the position error at one sigma.
func (source Source) Error() ErrorEllipse { return source.ellipse }

// Epoch returns the epoch of the position, if the catalogue gave one.
func (source Source) Epoch() When { return source.epoch }

// Measurements returns the source's measurements, one or more.
func (source Source) Measurements() []Measurement {
	return slices.Clone(source.measurements)
}

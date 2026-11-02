package nilgai

import (
	"errors"
	"fmt"
	"slices"
)

// Catalogue names one of the inputs.
type Catalogue string

// Catalogue’s values identify the supported record formats. AllWISE has one row
// per source; Neowise single exposures can be grouped into light curves. Reader
// implementations are listed in package read.
const (
	MeerKAT Catalogue = "MeerKAT"
	Fermi   Catalogue = "Fermi-4FGL"
	Chandra Catalogue = "Chandra-CSC"
	NEOWISE Catalogue = "NEOWISE"
	AllWISE Catalogue = "AllWISE"
)

// valid reports whether the catalogue has a defined record format.
func (catalogue Catalogue) valid() bool {
	switch catalogue {
	case MeerKAT, Fermi, Chandra, NEOWISE, AllWISE:
		return true
	}
	return false
}

// Source holds one catalogue row in the common format. NewSource validates the
// identifier, error ellipse, and measurements before constructing it.
type Source struct {
	catalogue    Catalogue
	id           string
	position     Position
	ellipse      ErrorEllipse
	epoch        When
	measurements []Measurement
}

// NewSource constructs a catalogue record. It returns nil and an error for an
// unknown catalogue, empty identifier, invalid ellipse, no measurements, or a
// measurement not initialized by its constructor.
//
// The identifier is preserved verbatim. The ellipse is at one sigma, and epoch
// is the position’s epoch when known; a zero When means unknown.
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

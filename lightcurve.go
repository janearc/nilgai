package nilgai

import (
	"cmp"
	"errors"
	"fmt"
	"slices"
)

// LightCurve holds every look at one position, in time order. Neowise
// rows are looks and not sources, so they are grouped here and not folded
// into one number. Make one with NewLightCurve.
type LightCurve struct {
	id       string
	position Position
	ellipse  ErrorEllipse
	points   []Measurement
}

// byTime orders two looks by their MJD. It is a comparison function for
// the slices package.
func byTime(first, second Measurement) int {
	firstMJD, _ := first.Time().MJD()
	secondMJD, _ := second.Time().MJD()
	return cmp.Compare(firstMJD, secondMJD)
}

// NewLightCurve returns a LightCurve of the looks given, which it copies
// and sorts by time. The id names the position, as the reader grouped it.
//
// NewLightCurve refuses, and returns nil on a refusal:
//   - an empty identifier
//   - an ellipse that did not come from NewErrorEllipse
//   - no points
//   - a point that no constructor made
//   - a point with no time
func NewLightCurve(id string, position Position, ellipse ErrorEllipse,
	points []Measurement) (*LightCurve, error) {
	if id == "" {
		return nil, errors.New("light curve has no identifier")
	}
	if !ellipse.valid() {
		return nil, errors.New("light curve has no error ellipse")
	}
	if len(points) == none {
		return nil, errors.New("light curve has no points")
	}
	for place, each := range points {
		if _, known := each.Time().MJD(); !known || !each.valid() {
			return nil, fmt.Errorf(
				"point %d has no time or was not "+
					"made by a constructor", place)
		}
	}
	own := slices.Clone(points)
	slices.SortStableFunc(own, byTime)
	return &LightCurve{id: id, position: position, ellipse: ellipse,
		points: own}, nil
}

// ID returns the name of the position.
func (curve LightCurve) ID() string { return curve.id }

// Position returns the ICRS position in degrees.
func (curve LightCurve) Position() Position { return curve.position }

// Error returns the position error at one sigma.
func (curve LightCurve) Error() ErrorEllipse { return curve.ellipse }

// Points returns the looks, earliest first.
func (curve LightCurve) Points() []Measurement {
	return slices.Clone(curve.points)
}

// Len returns the number of looks.
func (curve LightCurve) Len() int { return len(curve.points) }

// InFilter returns the looks through the filter of that name, in time
// order.
func (curve LightCurve) InFilter(name string) []Measurement {
	var within []Measurement
	for _, each := range curve.points {
		if each.Filter() == name {
			within = append(within, each)
		}
	}
	return within
}

// Span returns the MJD of the first and last look. The ok result is false
// for a curve that NewLightCurve did not make, which has no looks.
func (curve LightCurve) Span() (first, last float64, ok bool) {
	for _, each := range curve.points {
		mjd, _ := each.Time().MJD()
		if !ok {
			first, ok = mjd, true
		}
		last = mjd
	}
	return first, last, ok
}

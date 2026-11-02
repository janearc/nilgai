package nilgai

import (
	"cmp"
	"errors"
	"fmt"
	"slices"
)

// LightCurve groups measurements at one position in time order. Neowise single
// exposures remain separate observations, preserving variability instead of
// collapsing them into one brightness. Use NewLightCurve.
type LightCurve struct {
	id       string
	position Position
	ellipse  ErrorEllipse
	points   []Measurement
}

// byTime compares observation MJDs for sorting with slices.
func byTime(first, second Measurement) int {
	firstMJD, _ := first.Time().MJD()
	secondMJD, _ := second.Time().MJD()
	return cmp.Compare(firstMJD, secondMJD)
}

// NewLightCurve copies the measurements and sorts them stably by time. The
// identifier names the position grouped by the reader.
//
// It returns nil and an error for an empty identifier, invalid ellipse, no
// points, an uninitialized measurement, or a measurement without a time.
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

// Points returns a copy of the observations, earliest first.
func (curve LightCurve) Points() []Measurement {
	return slices.Clone(curve.points)
}

// Len returns the number of observations.
func (curve LightCurve) Len() int { return len(curve.points) }

// InFilter returns observations through the named filter in time order.
func (curve LightCurve) InFilter(name string) []Measurement {
	var within []Measurement
	for _, each := range curve.points {
		if each.Filter() == name {
			within = append(within, each)
		}
	}
	return within
}

// Span returns the earliest and latest MJD. The ok result is false for an
// uninitialized curve with no observations.
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

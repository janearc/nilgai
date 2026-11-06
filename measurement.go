package nilgai

import (
	"errors"
	"fmt"
	"slices"

	"github.com/janearc/libtheme-css/primitives/bands"
)

// Kind says whether a Measurement is a detection or an upper limit.
type Kind int

const (
	// kindUnset is the zero Kind, which is neither a detection nor an
	// upper limit.
	kindUnset Kind = iota
	// Detection is a measured value with an error.
	Detection
	// UpperLimit says the flux is at most the value. It has no error.
	UpperLimit
)

// String returns the kind in words, and "invalid" for anything else.
func (kind Kind) String() string {
	switch kind {
	case Detection:
		return "detection"
	case UpperLimit:
		return "upper limit"
	}
	return "invalid"
}

// Measurement records brightness through one filter for a source or a single
// exposure. It keeps the catalogue’s filter name, such as W1, and the
// corresponding span of libtheme bands; see band.BandsBetween.
//
// Construct measurements with NewDetection or NewUpperLimit. The With methods
// return modified copies.
type Measurement struct {
	filter  string
	first   bands.Band
	last    bands.Band
	kind    Kind
	value   Flux
	err     Flux
	when    When
	flags   []Flag
	quality Quality
	raw     string
}

// NewDetection returns a Measurement of a measured flux and its
// one-sigma error, through the named filter covering libtheme bands first
// to last. The flux and its error are one type of Flux, so a flux density
// with an energy-flux error does not compile.
//
// NewDetection refuses, and returns nil on a refusal:
//   - a flux or error below zero, NaN or infinite
//   - a filter with no name
//   - a first band above the last
//
// NewDetection's measurement starts with Quality Good; WithReading sets
// it from the flags.
func NewDetection[Of oneFlux](
	filter string,
	first, last bands.Band,
	value, sigma Of,
) (*Measurement, error) {
	if why := whyNotAFlux(float64(value)); why != nil {
		return nil, fmt.Errorf("value %w", why)
	}
	if why := whyNotAFlux(float64(sigma)); why != nil {
		return nil, fmt.Errorf("error %w", why)
	}
	if why := whyNotAFilter(filter, first, last); why != nil {
		return nil, why
	}
	return &Measurement{filter: filter, first: first, last: last,
		kind: Detection, value: value, err: sigma, quality: Good}, nil
}

// NewUpperLimit returns a Measurement that says the flux through the
// named filter, covering libtheme bands first to last, is at most the
// limit. The Quality starts as Good.
//
// NewUpperLimit refuses what NewDetection refuses of a flux and a filter,
// and returns nil on a refusal.
func NewUpperLimit[Of oneFlux](
	filter string,
	first, last bands.Band,
	limit Of,
) (*Measurement, error) {
	if why := whyNotAFlux(float64(limit)); why != nil {
		return nil, fmt.Errorf("limit %w", why)
	}
	if why := whyNotAFilter(filter, first, last); why != nil {
		return nil, why
	}
	return &Measurement{filter: filter, first: first, last: last,
		kind: UpperLimit, value: limit, quality: Good}, nil
}

// whyNotAFilter returns why a filter cannot label a measurement: it has
// no name, or its first band is above its last. It returns nil for a
// filter that can.
func whyNotAFilter(filter string, first, last bands.Band) error {
	if filter == "" {
		return errors.New("measurement has no filter")
	}
	if first > last {
		return fmt.Errorf("filter %q: first band %d is above last %d",
			filter, int(first), int(last))
	}
	return nil
}

// WithTime returns the measurement taken at the time given.
func (measurement Measurement) WithTime(when When) Measurement {
	measurement.when = when
	return measurement
}

// WithRaw returns the measurement with the catalogue's own text for the
// value, before any conversion. For example, it can hold the magnitude
// that a flux density was converted from.
func (measurement Measurement) WithRaw(raw string) Measurement {
	measurement.raw = raw
	return measurement
}

// WithReading returns a copy with the catalogue’s flags and their quality
// assessment. Flags are copied. An undefined quality returns nil and an error;
// valid qualities are Good, Doubtful, and Bad.
func (measurement Measurement) WithReading(quality Quality,
	flags []Flag) (*Measurement, error) {
	if !quality.valid() {
		return nil, fmt.Errorf("quality %d is not "+
			"good, doubtful or bad", int(quality))
	}
	read := measurement.readAs(quality, flags)
	return &read, nil
}

// readAs attaches a quality assessment and a copy of the flags. Callers supply
// a known quality constant, so it does not repeat validation.
func (measurement Measurement) readAs(quality Quality,
	flags []Flag) Measurement {
	measurement.quality = quality
	measurement.flags = slices.Clone(flags)
	return measurement
}

// Filter returns the catalogue's name for the filter the measurement was
// made through, such as "W1".
func (measurement Measurement) Filter() string { return measurement.filter }

// Bands returns the first and last libtheme band that the filter covers.
func (measurement Measurement) Bands() (first, last bands.Band) {
	return measurement.first, measurement.last
}

// Kind returns Detection or UpperLimit.
func (measurement Measurement) Kind() Kind { return measurement.kind }

// IsUpperLimit says whether the measurement is an upper limit.
func (measurement Measurement) IsUpperLimit() bool {
	return measurement.kind == UpperLimit
}

// Value returns the flux of a detection, and the limit of an upper limit.
func (measurement Measurement) Value() Flux { return measurement.value }

// Error returns the one-sigma error of a detection, in the unit of Value.
// Error returns nil for an upper limit.
func (measurement Measurement) Error() Flux { return measurement.err }

// Time returns the observation time, when the measurement has one.
func (measurement Measurement) Time() When { return measurement.when }

// Flags returns the catalogue's flags for this measurement, verbatim.
func (measurement Measurement) Flags() []Flag {
	return slices.Clone(measurement.flags)
}

// Quality returns Nilgai’s interpretation of the catalogue flags.
func (measurement Measurement) Quality() Quality { return measurement.quality }

// Raw returns the catalogue's own text for the value, if it was kept.
func (measurement Measurement) Raw() string { return measurement.raw }

// valid reports whether a constructor initialized the measurement.
func (measurement Measurement) valid() bool {
	return measurement.kind != kindUnset
}

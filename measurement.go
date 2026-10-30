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
	// kindUnset is the zero Kind: neither.
	kindUnset Kind = iota
	// Detection is a measured value with an error.
	Detection
	// UpperLimit says the flux is at most the value; there is no error.
	UpperLimit
)

// String is the kind in words, and "invalid" for anything else.
func (kind Kind) String() string {
	switch kind {
	case Detection:
		return "detection"
	case UpperLimit:
		return "upper limit"
	}
	return "invalid"
}

// Measurement is the brightness through one filter for one source or
// one look. The filter is the catalogue's name for it, such as "W1",
// and the run of libtheme bands it covers; see BandsBetween.
//
// Every Measurement comes from NewDetection or NewUpperLimit and is
// changed only by the With methods, each of which returns a new one.
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

// NewDetection makes a Measurement of a measured flux and its one-sigma
// error, which must be in the same unit, through the named filter
// covering libtheme bands first to last.
//
// A flux or error below zero, or one that is NaN or infinite, is
// refused, and so is a filter with no name or a first band above the
// last. Its Quality starts as Good; WithReading sets it from the flags.
func NewDetection(
	filter string,
	first, last bands.Band,
	value, sigma Flux,
) (measurement Measurement, err error) {
	if err = checkFlux("error", sigma); err != nil {
		return measurement, err
	}
	if err = checkFlux("value", value); err != nil {
		return measurement, err
	}
	if sigma.Unit() != value.Unit() {
		return measurement, fmt.Errorf(
			"value is in %s but error is in %s",
			value.Unit(), sigma.Unit())
	}
	if err = checkFilter(filter, first, last); err != nil {
		return measurement, err
	}
	return Measurement{filter: filter, first: first, last: last,
		kind: Detection, value: value, err: sigma, quality: Good}, nil
}

// NewUpperLimit makes a Measurement that says the flux through the named
// filter, covering libtheme bands first to last, is at most the limit.
// Its Quality starts as Good.
func NewUpperLimit(
	filter string,
	first, last bands.Band,
	limit Flux,
) (measurement Measurement, err error) {
	if err = checkFlux("limit", limit); err != nil {
		return measurement, err
	}
	if err = checkFilter(filter, first, last); err != nil {
		return measurement, err
	}
	return Measurement{filter: filter, first: first, last: last,
		kind: UpperLimit, value: limit, quality: Good}, nil
}

// checkFilter refuses a filter with no name, and one whose first band
// is above its last.
func checkFilter(filter string, first, last bands.Band) error {
	if filter == "" {
		return errors.New("measurement has no filter")
	}
	if first > last {
		return fmt.Errorf("filter %q: first band %d is above last %d",
			filter, int(first), int(last))
	}
	return nil
}

// WithTime is the measurement taken at the time given.
func (measurement Measurement) WithTime(when When) Measurement {
	measurement.when = when
	return measurement
}

// WithRaw is the measurement with the catalogue's own text for the
// value, before any conversion, such as the magnitude a flux density
// came from.
func (measurement Measurement) WithRaw(raw string) Measurement {
	measurement.raw = raw
	return measurement
}

// WithReading is the measurement with the catalogue's flags, kept as
// given, and nilgai's reading of them. It refuses a reading that is not
// good, doubtful or bad.
func (measurement Measurement) WithReading(quality Quality,
	flags []Flag) (Measurement, error) {
	if !quality.valid() {
		return Measurement{}, fmt.Errorf("quality %d is not "+
			"good, doubtful or bad", int(quality))
	}
	measurement.quality = quality
	measurement.flags = slices.Clone(flags)
	return measurement, nil
}

// Filter is the catalogue's name for what the measurement was made
// through, such as "W1".
func (measurement Measurement) Filter() string { return measurement.filter }

// Bands are the first and last libtheme bands the filter covers.
func (measurement Measurement) Bands() (first, last bands.Band) {
	return measurement.first, measurement.last
}

// Kind is Detection or UpperLimit.
func (measurement Measurement) Kind() Kind { return measurement.kind }

// IsUpperLimit says whether the measurement is an upper limit.
func (measurement Measurement) IsUpperLimit() bool {
	return measurement.kind == UpperLimit
}

// Value is the flux of a detection, and the limit of an upper limit.
func (measurement Measurement) Value() Flux { return measurement.value }

// Error is the one-sigma error of a detection, in the unit of Value.
// It is nil for an upper limit.
func (measurement Measurement) Error() Flux { return measurement.err }

// Time is the time of a single look, when the measurement is one.
func (measurement Measurement) Time() When { return measurement.when }

// Flags are the catalogue's flags for this measurement, verbatim.
func (measurement Measurement) Flags() []Flag {
	return slices.Clone(measurement.flags)
}

// Quality is nilgai's reading of Flags.
func (measurement Measurement) Quality() Quality { return measurement.quality }

// Raw is the catalogue's own text for the value, if it was kept.
func (measurement Measurement) Raw() string { return measurement.raw }

// valid says whether the measurement was made by a constructor.
func (measurement Measurement) valid() bool {
	return measurement.kind != kindUnset
}

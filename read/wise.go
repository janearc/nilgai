package read

import (
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"math"

	"github.com/janearc/nilgai"
)

// WiseBand names one of WISE's four bands.
type WiseBand int

const (
	_ WiseBand = iota // the zero WiseBand is none of them
	// W1 is the 3.4 micron band.
	W1
	// W2 is the 4.6 micron band.
	W2
	// W3 is the 12 micron band.
	W3
	// W4 is the 22 micron band.
	W4
)

// zeroPointData holds the zero points of the WISE Vega magnitude system,
// in Jy, with the source they come from. Other people measured them, so
// they live in data/wise-zero-points.json and not in code.
//
//go:embed data/wise-zero-points.json
var zeroPointData []byte

// WiseZeroPoints holds the flux densities of magnitude zero in W1 to W4.
type WiseZeroPoints struct{ w1, w2, w3, w4 nilgai.FluxDensity }

// ReadWiseZeroPoints reads zero points in Jy from JSON shaped like
// data/wise-zero-points.json, with keys "W1" to "W4". Each must be above
// zero. A missing key reads as zero and is refused.
func ReadWiseZeroPoints(data []byte) (points WiseZeroPoints, err error) {
	var read struct{ W1, W2, W3, W4 float64 }
	if err = json.Unmarshal(data, &read); err != nil {
		return points, fmt.Errorf("WISE zero points: %w", err)
	}
	given := []float64{read.W1, read.W2, read.W3, read.W4}
	for place, value := range given {
		if !nilgai.Positive(value) {
			return points, fmt.Errorf("WISE zero point W%d is %v "+
				"and must be above zero", place+int(W1), value)
		}
	}
	return WiseZeroPoints{
		w1: nilgai.Jansky(read.W1),
		w2: nilgai.Jansky(read.W2),
		w3: nilgai.Jansky(read.W3),
		w4: nilgai.Jansky(read.W4),
	}, nil
}

// For returns the zero point of the band. It refuses any band but W1 to
// W4.
func (points WiseZeroPoints) For(band WiseBand) (zero nilgai.FluxDensity,
	err error) {
	switch band {
	case W1:
		return points.w1, nil
	case W2:
		return points.w2, nil
	case W3:
		return points.w3, nil
	case W4:
		return points.w4, nil
	}
	return zero, fmt.Errorf("WISE band %d is not W1, W2, W3 or W4",
		int(band))
}

// ZeroPoint returns the flux density of magnitude zero in the band.
// ZeroPoint's values come from data/wise-zero-points.json, which takes
// them from the WISE All-Sky Release Explanatory Supplement, section
// IV.4.h, Table 1. ZeroPoint refuses any band but W1 to W4.
func (band WiseBand) ZeroPoint() (zero nilgai.FluxDensity, err error) {
	points, readErr := ReadWiseZeroPoints(zeroPointData)
	zero, err = points.For(band)
	return zero, errors.Join(readErr, err)
}

// pogson is the number of magnitudes in a factor of ten in flux, 2.5,
// because Pogson's scale makes five magnitudes a factor of 100 (Pogson
// 1856; IAU).
const pogson = 2.5

// tenth is ten, the base of the magnitude scale's decimal logarithm.
const tenth = 10.0

// WiseFlux converts a WISE Vega magnitude and its error to a flux
// density and its error.
//
// WiseFlux computes F = F0 * 10^(-m / 2.5), where F0 is the band's zero
// point (Jarrett et al. 2011, ApJ 735, 112, as the Explanatory Supplement
// gives it).
//
// WiseFlux propagates the error to first order: dF/dm = -F * ln(10) / 2.5,
// so sigma_F = F * ln(10) / 2.5 * sigma_m. First order holds when sigma_m
// is small, as it is for a detection; an error near one magnitude makes a
// lopsided flux error that this does not describe.
//
// WiseFlux refuses a magnitude that is not a finite number, and an error
// that is below zero or not finite.
func WiseFlux(band WiseBand, magnitude, magnitudeError float64) (flux,
	fluxError nilgai.FluxDensity, err error) {
	zero, err := band.ZeroPoint()
	if err != nil {
		return flux, fluxError, err
	}
	if !nilgai.Finite(magnitude) {
		return flux, fluxError, fmt.Errorf(
			"magnitude %v is not a finite "+
				"number", magnitude)
	}
	if !nilgai.NotBelowNone(magnitudeError) {
		return flux, fluxError, fmt.Errorf("magnitude error "+
			"%v must be finite and not below zero", magnitudeError)
	}
	density := float64(zero) * math.Pow(tenth, -magnitude/pogson)
	if !nilgai.Finite(density) {
		return flux, fluxError, fmt.Errorf("magnitude %v gives a flux "+
			"that does not fit a float", magnitude)
	}
	slope := math.Ln10 / pogson
	return nilgai.FluxDensity(
			density,
		), nilgai.FluxDensity(
			density * slope * magnitudeError,
		),
		nil
}

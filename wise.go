package nilgai

import (
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"math"
)

// WiseBand is one of the two WISE bands Neowise keeps working in.
type WiseBand int

const (
	_ WiseBand = iota // the zero WiseBand is neither
	// W1 is the 3.4 micron band.
	W1
	// W2 is the 4.6 micron band.
	W2
)

// zeroPointData holds the zero points of the WISE Vega magnitude
// system, in Jy, with the paper they come from. They are measured, so
// they live in data/wise-zero-points.json and not in code.
//
//go:embed data/wise-zero-points.json
var zeroPointData []byte

// WiseZeroPoints are the flux densities of magnitude zero in W1 and W2, in
// Jy.
type WiseZeroPoints struct{ w1, w2 FluxDensity }

// ReadWiseZeroPoints reads zero points from JSON shaped as
// data/wise-zero-points.json is, with keys "W1" and "W2". Both must be
// above zero.
func ReadWiseZeroPoints(data []byte) (points WiseZeroPoints, err error) {
	var read struct{ W1, W2 float64 }
	if err = json.Unmarshal(data, &read); err != nil {
		return points, fmt.Errorf("WISE zero points: %w", err)
	}
	if !positive(read.W1) || !positive(read.W2) {
		return points, fmt.Errorf("WISE zero points %v and %v must be "+
			"above zero", read.W1, read.W2)
	}
	return WiseZeroPoints{w1: FluxDensity(read.W1),
		w2: FluxDensity(read.W2)}, nil
}

// For is the zero point of the band. Any band but W1 and W2 is refused.
func (points WiseZeroPoints) For(band WiseBand) (zero FluxDensity,
	err error) {
	switch band {
	case W1:
		return points.w1, nil
	case W2:
		return points.w2, nil
	}
	return zero, fmt.Errorf("WISE band %d is not W1 or W2", int(band))
}

// ZeroPoint is the flux density of magnitude zero in the band, in Jy,
// from Jarrett et al. 2011, ApJ 735, 112, as data/wise-zero-points.json
// gives it. Any other band is refused.
func (band WiseBand) ZeroPoint() (zero FluxDensity, err error) {
	points, readErr := ReadWiseZeroPoints(zeroPointData)
	zero, err = points.For(band)
	return zero, errors.Join(readErr, err)
}

// pogson is defined by the magnitude scale: five magnitudes are a
// factor of 100 in flux, so a factor of ten is 2.5 magnitudes (Pogson
// 1856; IAU).
const pogson = 2.5

// tenth is defined: the base of the decimal logarithm in the
// magnitude scale.
const tenth = 10.0

// WiseFlux converts a WISE Vega magnitude and its error to a flux
// density and its error, both in Jy.
//
// The flux is F = F0 * 10^(-m / 2.5), with F0 the band's zero point
// (Jarrett et al. 2011, ApJ 735, 112).
//
// The error is propagated to first order: dF/dm = -F * ln(10) / 2.5,
// so sigma_F = F * ln(10) / 2.5 * sigma_m.
//
// That holds when sigma_m is small, as it is for a detection;
// a magnitude with an error near one has a lopsided flux error that
// this does not describe. The magnitude may be any finite number; the
// error may not be below zero.
func WiseFlux(band WiseBand, magnitude, magnitudeError float64) (flux,
	fluxError FluxDensity, err error) {
	zero, err := band.ZeroPoint()
	if err != nil {
		return flux, fluxError, err
	}
	if !finite(magnitude) {
		return flux, fluxError, fmt.Errorf(
			"magnitude %v is not a finite "+
				"number", magnitude)
	}
	if !notBelowNone(magnitudeError) {
		return flux, fluxError, fmt.Errorf("magnitude error "+
			"%v must be finite and not below zero", magnitudeError)
	}
	density := float64(zero) * math.Pow(tenth, -magnitude/pogson)
	if !finite(density) {
		return flux, fluxError, fmt.Errorf("magnitude %v gives a flux "+
			"that does not fit a float", magnitude)
	}
	slope := math.Ln10 / pogson
	return FluxDensity(
			density,
		), FluxDensity(
			density * slope * magnitudeError,
		),
		nil
}

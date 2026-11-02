package nilgai

import (
	"fmt"
	"math"
)

// Confidence is the catalogue’s probability of the true position lying inside
// its ellipse, such as 0.95. It must be strictly between zero and one; neither
// endpoint defines a finite, positive Gaussian error radius.
type Confidence float64

// PercentToConfidence converts a catalogue percentage, such as 95, to a
// Confidence. Values outside the open interval (0, 100) return an error.
func PercentToConfidence(percent float64) (confidence Confidence, err error) {
	confidence = Confidence(percent / percentPerWhole)
	if err = confidence.check(); err != nil {
		return Confidence(none), err
	}
	return confidence, nil
}

// percentPerWhole is the number of percent in a whole, 100.
const percentPerWhole = 100.0

// check rejects a confidence outside the open interval (0, 1).
func (confidence Confidence) check() error {
	level := float64(confidence)
	if !Positive(level) || level >= whole {
		return fmt.Errorf(
			"confidence %v must be above 0 and below 1", level)
	}
	return nil
}

// whole is the probability of the whole sample space, 1.
const whole = 1.0

// halfSquare is the 1/2 in exp(-1/2). For a Gaussian, the density at one
// sigma is exp(-1/2) of its peak.
const halfSquare = 0.5

// radiusLawTwo is the 2 in -2 ln(1 - p), in the radius law that
// RadiusFactor uses. It is also the 2 in 2 sigma^2.
const radiusLawTwo = 2.0

// OneSigma returns the probability inside a one-sigma ellipse of a 2-D
// Gaussian, 1 - exp(-1/2), about 0.3935. MeerKAT states one-sigma errors,
// which are at this level. RadiusFactor gives a radius of one at this
// level.
func OneSigma() Confidence {
	return Confidence(-math.Expm1(-halfSquare))
}

// RadiusFactor returns the radius, in sigma, of the circle that holds the
// given confidence, for a 2-D Gaussian with equal sigma on both axes.
//
// RadiusFactor solves for r in 1 - exp(-r^2 / (2 sigma^2)) = c, the
// chance of falling inside radius r: r = sigma * sqrt(-2 ln(1 - c)). For
// 95% the factor is about 2.4477.
//
// RadiusFactor's formula is the Rayleigh distribution's quantile function
// (Evans, Hastings and Peacock, Statistical Distributions, 4th ed.).
//
// RadiusFactor scales both semi-axes of an ellipse stated at level c,
// because each axis is an independent Gaussian.
func RadiusFactor(confidence Confidence) (factor float64, err error) {
	if err = confidence.check(); err != nil {
		return factor, err
	}
	logOutside := math.Log1p(-float64(confidence))
	return math.Sqrt(-radiusLawTwo * logOutside), nil
}

// ToOneSigma converts a length stated at the confidence given to one
// sigma, by dividing by RadiusFactor.
func ToOneSigma(stated float64, confidence Confidence) (length float64,
	err error) {
	if !NotBelowNone(stated) {
		return length, fmt.Errorf("length %v must be finite and not "+
			"below zero", stated)
	}
	factor, err := RadiusFactor(confidence)
	if err != nil {
		return length, err
	}
	return stated / factor, nil
}

// ErrorEllipse stores one-sigma positional semi-axes alongside the catalogue’s
// original axes and confidence level, so the conversion can be checked.
type ErrorEllipse struct {
	semiMajor, semiMinor             float64
	positionAngle                    float64
	statedAt                         Confidence
	statedSemiMajor, statedSemiMinor float64
}

// NewErrorEllipse converts catalogue semi-axes in arcseconds at the given
// confidence to one sigma, retaining the stated values.
//
// The position angle is in degrees east of north, in [0, 360). Axes must be
// finite and positive, with the major axis at least as long as the minor.
func NewErrorEllipse(major, minor, angle float64,
	stated Confidence) (ellipse ErrorEllipse, err error) {
	if !Positive(major) || !Positive(minor) {
		return ellipse, fmt.Errorf("semi-axes %v and %v arcsec "+
			"must be finite and above zero", major, minor)
	}
	if major < minor {
		return ellipse, fmt.Errorf("semi-major %v is shorter than "+
			"semi-minor %v", major, minor)
	}
	if !NotBelowNone(angle) || angle >= FullCircle {
		return ellipse, fmt.Errorf("position angle %v deg is outside "+
			"[0, 360)", angle)
	}
	factor, err := RadiusFactor(stated)
	if err != nil {
		return ellipse, err
	}
	return ErrorEllipse{semiMajor: major / factor,
		semiMinor:     minor / factor,
		positionAngle: angle, statedAt: stated, statedSemiMajor: major,
		statedSemiMinor: minor}, nil
}

// SemiMajor returns the one-sigma semi-major axis, in arcseconds.
func (ellipse ErrorEllipse) SemiMajor() float64 { return ellipse.semiMajor }

// SemiMinor returns the one-sigma semi-minor axis, in arcseconds.
func (ellipse ErrorEllipse) SemiMinor() float64 { return ellipse.semiMinor }

// PositionAngle returns the angle of the major axis, in degrees east of
// north.
func (ellipse ErrorEllipse) PositionAngle() float64 {
	return ellipse.positionAngle
}

// StatedAt returns the confidence that the catalogue's semi-axes were
// stated at.
func (ellipse ErrorEllipse) StatedAt() Confidence { return ellipse.statedAt }

// StatedSemiMajor returns the catalogue's own semi-major axis in
// arcseconds, at StatedAt.
func (ellipse ErrorEllipse) StatedSemiMajor() float64 {
	return ellipse.statedSemiMajor
}

// StatedSemiMinor returns the catalogue's own semi-minor axis in
// arcseconds, at StatedAt.
func (ellipse ErrorEllipse) StatedSemiMinor() float64 {
	return ellipse.statedSemiMinor
}

// valid reports whether NewErrorEllipse initialized the ellipse.
func (ellipse ErrorEllipse) valid() bool {
	return ellipse.statedAt.check() == nil
}

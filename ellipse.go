package nilgai

import (
	"fmt"
	"math"
)

// Confidence is the probability that the true position lies inside an
// error ellipse, as the catalogue stated it: 0.95 for a 95% ellipse. It
// excludes both ends, because 0 and 1 mean nothing for a Gaussian.
type Confidence float64

// PercentToConfidence converts a stated percentage to a Confidence, so
// that a reader can pass the catalogue's "95" as the catalogue wrote it.
// PercentToConfidence refuses a percentage that is not above 0 and
// below 100.
func PercentToConfidence(percent float64) (confidence Confidence, err error) {
	confidence = Confidence(percent / percentPerWhole)
	if err = confidence.check(); err != nil {
		return Confidence(none), err
	}
	return confidence, nil
}

// percentPerWhole is the number of percent in a whole, 100.
const percentPerWhole = 100.0

// check refuses a confidence that is not strictly between 0 and 1.
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

// ErrorEllipse holds the error in a position, as one-sigma semi-axes. It
// also keeps what the catalogue stated, so a reader of the record can see
// what was converted and from what.
type ErrorEllipse struct {
	semiMajor, semiMinor             float64
	positionAngle                    float64
	statedAt                         Confidence
	statedSemiMajor, statedSemiMinor float64
}

// NewErrorEllipse takes the catalogue's semi-axes in arcseconds, stated
// at the confidence given, and returns the ellipse at one sigma with the
// stated values kept.
//
// NewErrorEllipse takes the position angle in degrees east of north,
// from 0 up to but not including 360. Both axes must be finite and above
// zero, and the major axis must not be shorter than the minor.
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

// valid says whether NewErrorEllipse made the ellipse. A zero
// ErrorEllipse was not made by it.
func (ellipse ErrorEllipse) valid() bool {
	return ellipse.statedAt.check() == nil
}

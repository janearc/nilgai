package nilgai

import (
	"fmt"
	"math"
)

// Confidence is the probability that the true position lies inside an
// error ellipse, as the catalogue stated it: 0.95 for a 95% ellipse. It
// is open at both ends: 0 and 1 mean nothing for a Gaussian.
type Confidence float64

// PercentToConfidence turns a stated percentage into a Confidence, so
// that a reader passes the catalogue's "95" as the catalogue wrote it.
// It refuses a percentage that is not above 0 and below 100.
func PercentToConfidence(percent float64) (confidence Confidence, err error) {
	confidence = Confidence(percent / percentPerWhole)
	if err = confidence.check(); err != nil {
		return Confidence(none), err
	}
	return confidence, nil
}

// percentPerWhole is defined: a percentage is parts per hundred.
const percentPerWhole = 100.0

// check refuses a confidence that is not strictly between 0 and 1.
func (confidence Confidence) check() error {
	level := float64(confidence)
	if !positive(level) || level >= whole {
		return fmt.Errorf(
			"confidence %v must be above 0 and below 1", level)
	}
	return nil
}

// whole is defined: the probability of the whole sample space.
const whole = 1.0

// halfSquare is defined by the Gaussian: the density at one sigma is
// exp(-1/2) of its peak.
const halfSquare = 0.5

// radiusLawTwo is defined by the radius law in RadiusFactor: the 2 in
// -2 ln(1 - p), which is the 2 in 2 sigma^2.
const radiusLawTwo = 2.0

// OneSigma is the probability inside a one-sigma ellipse of a 2-D
// Gaussian, 1 - exp(-1/2), about 0.3935. MeerKAT states one-sigma
// errors, which are this level. It is the level at which RadiusFactor
// gives a radius of one.
func OneSigma() Confidence {
	return Confidence(-math.Expm1(-halfSquare))
}

// RadiusFactor is how many sigma wide the circle holding probability c
// is, for a 2-D Gaussian with equal sigma on both axes. The probability
// inside radius r is 1 - exp(-r^2 / (2 sigma^2)); solving for r gives
// r = sigma * sqrt(-2 ln(1 - c)). For 95% this is about 2.4477.
//
// This is
// the quantile function of the Rayleigh distribution; see e.g. Evans,
// Hastings and Peacock, Statistical Distributions, 4th ed., "Rayleigh
// distribution".
//
// An ellipse stated at level c has both of its semi-axes scaled by this
// factor, since each axis is an independent Gaussian.
func RadiusFactor(confidence Confidence) (factor float64, err error) {
	if err = confidence.check(); err != nil {
		return factor, err
	}
	logOutside := math.Log1p(-float64(confidence))
	return math.Sqrt(-radiusLawTwo * logOutside), nil
}

// ToOneSigma takes a length stated at the confidence given to one
// sigma, by dividing by RadiusFactor.
func ToOneSigma(stated float64, confidence Confidence) (length float64,
	err error) {
	if !notBelowNone(stated) {
		return length, fmt.Errorf("length %v must be finite and not "+
			"below zero", stated)
	}
	factor, err := RadiusFactor(confidence)
	if err != nil {
		return length, err
	}
	return stated / factor, nil
}

// ErrorEllipse is the error in a position, as one-sigma semi-axes. It
// also carries what the catalogue stated, so a reader of the record can
// see what was converted and from what.
type ErrorEllipse struct {
	semiMajor, semiMinor             float64
	positionAngle                    float64
	statedAt                         Confidence
	statedSemiMajor, statedSemiMinor float64
}

// NewErrorEllipse takes the catalogue's semi-axes in arcseconds, stated
// at the confidence given, and returns the ellipse at one sigma with
// the stated values kept.
//
// The position angle is in degrees east of
// north, from 0 up to but not including 360. The major axis must not be
// shorter than the minor, and the minor must be above zero.
func NewErrorEllipse(major, minor, angle float64,
	stated Confidence) (ellipse ErrorEllipse, err error) {
	if !positive(major) || !positive(minor) {
		return ellipse, fmt.Errorf("semi-axes %v and %v arcsec "+
			"must be finite and above zero", major, minor)
	}
	if major < minor {
		return ellipse, fmt.Errorf("semi-major %v is shorter than "+
			"semi-minor %v", major, minor)
	}
	if !notBelowNone(angle) || angle >= fullCircle {
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

// SemiMajor is the one-sigma semi-major axis, in arcseconds.
func (ellipse ErrorEllipse) SemiMajor() float64 { return ellipse.semiMajor }

// SemiMinor is the one-sigma semi-minor axis, in arcseconds.
func (ellipse ErrorEllipse) SemiMinor() float64 { return ellipse.semiMinor }

// PositionAngle is the angle of the major axis, in degrees east of north.
func (ellipse ErrorEllipse) PositionAngle() float64 {
	return ellipse.positionAngle
}

// StatedAt is the confidence the catalogue's semi-axes were stated at.
func (ellipse ErrorEllipse) StatedAt() Confidence { return ellipse.statedAt }

// StatedSemiMajor is the catalogue's own semi-major axis in arcseconds,
// at StatedAt.
func (ellipse ErrorEllipse) StatedSemiMajor() float64 {
	return ellipse.statedSemiMajor
}

// StatedSemiMinor is the catalogue's own semi-minor axis in arcseconds,
// at StatedAt.
func (ellipse ErrorEllipse) StatedSemiMinor() float64 {
	return ellipse.statedSemiMinor
}

// valid says whether the ellipse was made by NewErrorEllipse: a zero
// ErrorEllipse was not.
func (ellipse ErrorEllipse) valid() bool {
	return ellipse.statedAt.check() == nil
}

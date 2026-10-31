package nilgai

import "math"

// none is the count of nothing, zero. It is the origin of every scale in
// this package: a length, a flux and a count all start here.
const none = 0

// ArcsecondsPerDegree and the constants with it are the units of angle
// and of division that nilgai's packages share.
const (
	// ArcsecondsPerDegree is the arcseconds in a degree: sixty minutes
	// of sixty seconds.
	ArcsecondsPerDegree = 3600
	// DegreesInHalfACircle is the degrees in half a circle, which is pi
	// radians.
	DegreesInHalfACircle = 180
	// Halves is the number of equal parts in a whole that a half is one
	// of.
	Halves = 2
)

// Half returns half of value. A width between two limits is twice the
// error on either side, and the reach of a pen is half its width.
func Half(value float64) float64 { return value / Halves }

// Finite reports whether value is neither NaN nor an infinity. The
// packages beside this one use it so that they refuse the same inputs.
func Finite(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, none)
}

// Positive reports whether value is a finite number above zero.
func Positive(value float64) bool {
	return Finite(value) && value > none
}

// NotBelowNone reports whether value is a finite number that is not under
// zero.
func NotBelowNone(value float64) bool {
	return Finite(value) && value >= none
}

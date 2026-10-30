package nilgai

import "math"

// none is defined as a count of nothing, and the origin of every scale
// in this package: a length, a flux and a count start here.
const none = 0

// finite reports whether value is neither NaN nor an infinity.
func finite(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, none)
}

// positive reports whether value is a finite number above none.
func positive(value float64) bool {
	return finite(value) && value > none
}

// notBelowNone reports whether value is a finite number, not under none.
func notBelowNone(value float64) bool {
	return finite(value) && value >= none
}

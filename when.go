package nilgai

import "fmt"

// When is a time as a modified Julian date, or not known. The zero
// When is not known; that is how a source with no single look says so.
type When struct {
	mjd   float64
	known bool
}

// NewWhen makes a When from an MJD, refusing one that is NaN, infinite
// or below zero.
func NewWhen(mjd float64) (when When, err error) {
	if !notBelowNone(mjd) {
		return when, fmt.Errorf(
			"MJD %v must be finite and not below zero",
			mjd)
	}
	return When{mjd: mjd, known: true}, nil
}

// MJD is the date, and whether it is known.
func (when When) MJD() (mjd float64, known bool) { return when.mjd, when.known }

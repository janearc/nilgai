package nilgai

import "fmt"

// When holds a time as a modified Julian date, or records that the time
// is not known. The zero When is not known, which is how a source with no
// single look says so.
type When struct {
	mjd   float64
	known bool
}

// NewWhen returns a When from an MJD. It refuses an MJD that is NaN,
// infinite or below zero.
func NewWhen(mjd float64) (when When, err error) {
	if !NotBelowNone(mjd) {
		return when, fmt.Errorf(
			"MJD %v must be finite and not below zero",
			mjd)
	}
	return When{mjd: mjd, known: true}, nil
}

// MJD returns the date, and whether it is known.
func (when When) MJD() (mjd float64, known bool) { return when.mjd, when.known }

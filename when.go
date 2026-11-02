package nilgai

import "fmt"

// When stores a modified Julian date or an unknown time. The zero value means
// unknown, as for a source with no single observation time.
type When struct {
	mjd   float64
	known bool
}

// NewWhen constructs a time from a finite, nonnegative MJD.
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

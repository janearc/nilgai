package nilgai

import "fmt"

// Degrees in a full turn and in a right angle. Both are defined.
const (
	// fullCircle is defined: degrees in a full turn.
	fullCircle = 360.0
	// quarter is defined: degrees in a right angle.
	quarter = 90.0
)

// Position is a place on the sky in the ICRS frame, in degrees.
type Position struct {
	ra, dec float64
}

// NewPosition makes a Position, refusing a right ascension outside
// [0, 360), a declination outside [-90, 90], or a value that is NaN or
// infinite.
func NewPosition(ra, dec float64) (position Position, err error) {
	if !finite(ra) || ra < none || ra >= fullCircle {
		return position, fmt.Errorf(
			"right ascension %v deg is outside "+
				"[0, 360)", ra)
	}
	if !finite(dec) || dec < -quarter || dec > quarter {
		return position, fmt.Errorf("declination %v deg is outside "+
			"[-90, 90]", dec)
	}
	return Position{ra: ra, dec: dec}, nil
}

// RA is the right ascension in degrees, from 0 up to but not including 360.
func (position Position) RA() float64 { return position.ra }

// Dec is the declination in degrees, from -90 to 90.
func (position Position) Dec() float64 { return position.dec }

package nilgai

import (
	"fmt"
	"math"
)

// FullCircle and RightAngle are in degrees: a full turn and a right angle.
const (
	// FullCircle is the degrees in a full turn. The readers wrap position
	// angles by it.
	FullCircle = 360.0
	// RightAngle is the degrees in a right angle, which is the turn from
	// north to east.
	RightAngle = 90.0
)

// Position is a place on the sky in the ICRS frame, in degrees.
type Position struct {
	ra, dec float64
}

// NewPosition returns a Position. It refuses:
//   - a right ascension outside [0, 360)
//   - a declination outside [-90, 90]
//   - a value that is NaN or infinite
func NewPosition(ra, dec float64) (position Position, err error) {
	if !Finite(ra) || ra < none || ra >= FullCircle {
		return position, fmt.Errorf(
			"right ascension %v deg is outside "+
				"[0, 360)", ra)
	}
	if !Finite(dec) || dec < -RightAngle || dec > RightAngle {
		return position, fmt.Errorf("declination %v deg is outside "+
			"[-90, 90]", dec)
	}
	return Position{ra: ra, dec: dec}, nil
}

// RA returns the right ascension in degrees, from 0 up to but not
// including 360.
func (position Position) RA() float64 { return position.ra }

// Dec returns the declination in degrees, from -90 to 90.
func (position Position) Dec() float64 { return position.dec }

// galacticPoleRA and galacticPoleDec give the north Galactic pole in the
// ICRS, in degrees. The pole defines the Galactic frame. The IAU fixed
// the frame in 1958, and the Hipparcos catalogue restates it for the ICRS
// (ESA SP-1200, vol. 1, section 1.5.3).
//
// galacticPoleRA and galacticPoleDec's digits were recalled, not read
// from a source. TestGalacticLatitude checks them against the latitude
// 4FGL gives its sources, and they agree to 1.6e-5 degrees.
const (
	// galacticPoleRA is the pole's right ascension.
	galacticPoleRA = 192.85948
	// galacticPoleDec is the pole's declination.
	galacticPoleDec = 27.12825
	// sineBound is the largest magnitude a sine can have, 1.
	sineBound = 1.0
)

// GalacticLatitude returns the Galactic latitude b of the position in
// degrees,
// from -90 at the south Galactic pole to 90 at the north one. It is zero
// on the plane of the Milky Way.
func (position Position) GalacticLatitude() float64 {
	toward := position.dec * math.Pi / DegreesInHalfACircle
	pole := galacticPoleDec * math.Pi / DegreesInHalfACircle
	apart := (position.ra - galacticPoleRA) * math.Pi / DegreesInHalfACircle
	sine := math.Sin(toward)*math.Sin(pole) +
		math.Cos(toward)*math.Cos(pole)*math.Cos(apart)
	// Rounding can leave the sine a hair past one at a pole.
	sine = min(sineBound, max(-sineBound, sine))
	return math.Asin(sine) * DegreesInHalfACircle / math.Pi
}

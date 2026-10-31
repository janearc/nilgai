package nilgai_test

import (
	"testing"

	"github.com/janearc/nilgai"
)

// odd is five, an odd number that two does not divide into a whole number.
const odd = 5.0

// minutesInDegree and secondsInMinute make up a degree of arc: sixty
// minutes of sixty seconds.
const (
	// minutesInDegree is the minutes of arc in a degree.
	minutesInDegree = 60
	// secondsInMinute is the seconds of arc in a minute.
	secondsInMinute = 60
)

// TestHalfIsOneOfTwoEqualParts checks that Half divides a number into two
// equal parts. The half taken twice is the whole, and the half is not the
// whole.
func TestHalfIsOneOfTwoEqualParts(test *testing.T) {
	half := nilgai.Half(odd)
	if half+half != odd || half == odd {
		test.Errorf("half of %v is %v", odd, half)
	}
	if nilgai.Halves != nilgai.Half(nilgai.Halves)*nilgai.Halves {
		test.Errorf("Halves is %d", nilgai.Halves)
	}
}

// TestTheSharedUnitsMakeAFullCircle checks the units that the readers and
// the sky convert by. Two half circles and four right angles each make a
// full circle.
func TestTheSharedUnitsMakeAFullCircle(test *testing.T) {
	if nilgai.Halves*nilgai.DegreesInHalfACircle != nilgai.FullCircle {
		test.Error("two half circles are not a full circle")
	}
	if nilgai.RightAngle*nilgai.Halves*nilgai.Halves != nilgai.FullCircle {
		test.Error("four right angles are not a full circle")
	}
	if nilgai.ArcsecondsPerDegree != minutesInDegree*secondsInMinute {
		test.Errorf("%d arcseconds in a degree",
			nilgai.ArcsecondsPerDegree)
	}
}

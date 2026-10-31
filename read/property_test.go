package read_test

import (
	"fmt"
	"math"
	"strings"
	"testing"

	"github.com/janearc/nilgai"
	"github.com/janearc/nilgai/read"
	"pgregory.net/rapid"
)

// span is a range that a property draws numbers from.
type span struct{ Low, High float64 }

// rangesFixture holds testdata/rapid.json.
type rangesFixture struct {
	Sigma, Share, Radius, Angle span
	Rel                         Num
}

// loadRanges reads the property ranges.
func loadRanges(test *testing.T) rangesFixture {
	test.Helper()
	var read rangesFixture
	load(test, "rapid", &read)
	return read
}

// properties holds the checks that must pass for any input in range, with
// the readers they run on.
type properties struct {
	ranges  rangesFixture
	allwise *read.AllWISE
	chandra *read.Chandra
}

// newProperties returns the properties. It fails the test if a reader
// cannot be made.
func newProperties(test *testing.T) properties {
	test.Helper()
	return properties{ranges: loadRanges(test), allwise: allwiseTable(test),
		chandra: chandraTable(test)}
}

// drawIn returns a number drawn from the span.
func drawIn(property *rapid.T, from span, name string) float64 {
	return rapid.Float64Range(from.Low, from.High).Draw(property, name)
}

// ellipseOfRow reads the one row of the table and returns its ellipse.
func ellipseOfRow(property *rapid.T, table catalogueReader, header,
	row string) nilgai.ErrorEllipse {
	property.Helper()
	sources, errs := collect(table, strings.NewReader(header+row+"\n"))
	if len(sources) != one || len(errs) != none {
		property.Fatalf("%d sources and %v from %q", len(sources), errs, row)
	}
	return sources[none].Error()
}

// allwiseEllipse checks the ellipse for any errors that make a real
// ellipse:
//   - the axes are in order and above zero
//   - the angle is in 0 to 360
//   - the sum of the squared axes is the sum of the squared errors, which
//     is the covariance's trace
//   - the product of the axes, squared, is the covariance's determinant
func (check properties) allwiseEllipse(property *rapid.T) {
	read := check.ranges
	sigra := drawIn(property, read.Sigma, "sigra")
	sigdec := drawIn(property, read.Sigma, "sigdec")
	share := drawIn(property, read.Share, "share")
	sigradec := share * math.Sqrt(sigra*sigdec)
	cosigma := sigradec * math.Abs(sigradec)
	row := fmt.Sprintf("J1,266.4,-29.0,%v,%v,%v,15,0.1,0000,AAAA", sigra,
		sigdec, sigradec)
	header := "designation,ra,dec,sigra,sigdec,sigradec,w1mpro," +
		"w1sigmpro,cc_flags,ph_qual\n"
	ellipse := ellipseOfRow(property, check.allwise, header, row)
	major, minor := ellipse.SemiMajor(), ellipse.SemiMinor()
	trace := sigra*sigra + sigdec*sigdec
	det := sigra*sigra*sigdec*sigdec - cosigma*cosigma
	angle := ellipse.PositionAngle()
	rel := float64(read.Rel)
	if major < minor || minor <= none || angle < none ||
		angle >= halves*degreesInHalfACircle ||
		!near(Num(major*major+minor*minor), Num(trace), read.Rel) ||
		math.Abs(major*major*minor*minor-det) > rel*trace*trace {
		property.Fatalf("%s: %v by %v at %v", row, major, minor, angle)
	}
}

// chandraSwap checks that the same ellipse, named with its radii the other
// way round and its angle a quarter turn on, reads as the same ellipse. The
// order the CSC writes its radii in then does not change what is read.
func (check properties) chandraSwap(property *rapid.T) {
	read := check.ranges
	first := drawIn(property, read.Radius, "radius")
	second := drawIn(property, read.Radius, "other radius")
	angle := drawIn(property, read.Angle, "angle")
	smaller, larger := min(first, second), max(first, second)
	turned := math.Mod(angle+quarterTurn, halves*degreesInHalfACircle)
	ordered := check.chandraEllipse(property, smaller, larger, angle)
	reversed := check.chandraEllipse(property, larger, smaller, turned)
	firstMajor, secondMajor := ordered.SemiMajor(), reversed.SemiMajor()
	firstMinor, secondMinor := ordered.SemiMinor(), reversed.SemiMinor()
	major := near(Num(firstMajor), Num(secondMajor), read.Rel)
	minor := near(Num(firstMinor), Num(secondMinor), read.Rel)
	apart := turnsApart(ordered.PositionAngle(), reversed.PositionAngle())
	if !major || !minor || (smaller != larger &&
		apart > float64(read.Rel)) {
		property.Fatalf("radii %v and %v at %v: %v at %v, then %v at %v",
			smaller, larger, angle, ordered.SemiMajor(),
			ordered.PositionAngle(), reversed.SemiMajor(),
			reversed.PositionAngle())
	}
}

// chandraEllipse returns the ellipse of a CSC row with the radii and
// angle given.
func (check properties) chandraEllipse(property *rapid.T, r0, r1,
	angle float64) nilgai.ErrorEllipse {
	property.Helper()
	row := fmt.Sprintf("2CXO J3\t266.4\t-29.0\t%v\t%v\t%v\t1e-13\t"+
		"0.9e-13\t1.1e-13", r0, r1, angle)
	return ellipseOfRow(property, check.chandra, chandraHeader, row)
}

// quarterTurn is a right angle, in degrees.
const quarterTurn = 90

// TestAllWISEEllipseProperty checks that any errors that make an ellipse
// give one with its axes in order, its angle in range, and the
// covariance's trace and determinant.
func TestAllWISEEllipseProperty(test *testing.T) {
	rapid.Check(test, newProperties(test).allwiseEllipse)
}

// TestChandraSwapProperty checks that the order of the CSC's radii does
// not change the ellipse read.
func TestChandraSwapProperty(test *testing.T) {
	rapid.Check(test, newProperties(test).chandraSwap)
}

// TestChandraSwapPropertySeesAnUnturnedAngle is an anti-test. It checks
// that the properties fail on a wrong ellipse. A reader that left the
// angle alone on a swap would read the two radii orders a quarter turn
// apart, so the two orders differ by a quarter turn if nothing is turned.
func TestChandraSwapPropertySeesAnUnturnedAngle(test *testing.T) {
	check := newProperties(test)
	rapid.Check(test, check.unturnedDiffers)
}

// unturnedDiffers checks that reading the radii the other way round at the
// same angle gives a different ellipse, a quarter turn apart.
func (check properties) unturnedDiffers(property *rapid.T) {
	read := check.ranges
	first := drawIn(property, read.Radius, "radius")
	gap := drawIn(property, read.Radius, "gap")
	angle := drawIn(property, read.Angle, "angle")
	smaller, larger := first, first+gap
	ordered := check.chandraEllipse(property, smaller, larger, angle)
	reversed := check.chandraEllipse(property, larger, smaller, angle)
	apart := turnsApart(ordered.PositionAngle(), reversed.PositionAngle())
	if math.Abs(apart-quarterTurn) > float64(read.Rel) {
		property.Fatalf("the same radii the other way round are %v "+
			"degrees apart, not a quarter turn", apart)
	}
}

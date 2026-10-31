package read_test

import (
	"bytes"
	"errors"
	"fmt"
	"math"
	"strings"
	"testing"

	"github.com/janearc/nilgai"
	"github.com/janearc/nilgai/band"
	"github.com/janearc/nilgai/read"
)

// sgraFixture holds testdata/sgra.json. It has Sgr A*'s radio position,
// what the real Chandra search around it holds, and tables and rows to
// refuse.
type sgraFixture struct {
	SameTo  Num
	RA, Dec Num
	Chandra struct {
		File    string
		Sources int
		ID      string
		Filter  string
		Flux    Num
		Stated  Num
		Sigma   Num
		Pin     Num
		PinWhy  string

		Offset, OffsetWithin Num
		OffsetWhy            string
	}
	Ellipses []struct {
		Name                    string
		R0, R1, Angle           Num
		Major, Minor, WantAngle Num
	}
	BadChandra []struct{ Name, Table string }
	BadRows    []struct{ Column, Row string }
}

// loadSgrA reads the Sgr A* fixture.
func loadSgrA(test *testing.T) sgraFixture {
	test.Helper()
	var read sgraFixture
	load(test, "sgra", &read)
	return read
}

// sameTo returns how close two computed numbers must be to count as the
// same. The fixture picks it.
func sameTo(test *testing.T) Num {
	test.Helper()
	return loadSgrA(test).SameTo
}

// one is the count of a single thing.
const one = 1

// arcsecondsPerDegree is the arcseconds in a degree: sixty minutes of
// sixty seconds.
const arcsecondsPerDegree = 3600

// degreesInHalfACircle is the degrees in half a circle, which is pi
// radians.
const degreesInHalfACircle = 180

// separation returns the angle between two positions in arcseconds,
// treating the sky as flat. That is accurate enough for the few
// arcseconds that a pin is about.
func separation(ra, dec, otherRA, otherDec float64) float64 {
	radians := dec * math.Pi / degreesInHalfACircle
	across := (ra - otherRA) * math.Cos(radians)
	along := dec - otherDec
	return math.Hypot(across, along) * arcsecondsPerDegree
}

// chandraTable returns a Chandra reader built on nilgai's own profiles.
func chandraTable(test *testing.T) *read.Chandra {
	test.Helper()
	profiles, err := band.Profiles()
	if err != nil {
		test.Fatal(err)
	}
	table, err := read.NewChandra(profiles)
	if err != nil {
		test.Fatal(err)
	}
	return table
}

// chandraSources reads a Chandra table and fails the test on any error.
func chandraSources(test *testing.T, table []byte) []nilgai.Source {
	test.Helper()
	var got []nilgai.Source
	reading := chandraTable(test).Read("t", bytes.NewReader(table))
	for source, err := range reading {
		if err != nil {
			test.Fatal(err)
		}
		got = append(got, *source)
	}
	return got
}

// realSearch returns the real CSC search around Sgr A*, from testdata.
func realSearch(test *testing.T, read sgraFixture) []byte {
	test.Helper()
	table, err := fixtures.ReadFile("testdata/" + read.Chandra.File)
	if err != nil {
		test.Fatal(err)
	}
	return table
}

// TestChandraReadsTheRealSearch checks that the real search around Sgr A*
// reads whole, with every source and none lost.
func TestChandraReadsTheRealSearch(test *testing.T) {
	read := loadSgrA(test)
	got := chandraSources(test, realSearch(test, read))
	if len(got) != read.Chandra.Sources {
		test.Errorf("%d sources, want %d", len(got), read.Chandra.Sources)
	}
}

// TestChandraPinsSgrA is the pin. It checks that Chandra's Sgr A* sits
// within the fixture's bound of the radio position of Sgr A*, which is a
// fraction of one sigma.
//
// TestChandraPinsSgrA checks the sigma as well, so that a pass for the
// wrong reason shows. Sigma must match the fixture's hand-worked value for
// the stated 95% radius.
func TestChandraPinsSgrA(test *testing.T) {
	read := loadSgrA(test)
	for _, source := range chandraSources(test, realSearch(test, read)) {
		if source.ID() == read.Chandra.ID {
			pinned(test, read, source)
			return
		}
	}
	test.Errorf("no source is %s", read.Chandra.ID)
}

// pinned checks one source against the radio position and the fixture.
// TestChandraPinsSgrA makes this check on each source.
func pinned(test *testing.T, read sgraFixture, source nilgai.Source) {
	test.Helper()
	at := source.Position()
	off := separation(at.RA(), at.Dec(), float64(read.RA),
		float64(read.Dec))
	ellipse := source.Error()
	if off > float64(read.Chandra.Pin) {
		test.Errorf("Sgr A* is %v arcsec from the radio, past the "+
			"fixture's %v: %s", off, read.Chandra.Pin,
			read.Chandra.PinWhy)
	}
	if math.Abs(off-float64(read.Chandra.Offset)) >
		float64(read.Chandra.OffsetWithin) {
		test.Errorf("Sgr A* is %v arcsec from the radio, not the %v the "+
			"fixture works out: %s", off, read.Chandra.Offset,
			read.Chandra.OffsetWhy)
	}
	sigma, stated := ellipse.SemiMajor(), ellipse.StatedSemiMajor()
	if !near(Num(stated), read.Chandra.Stated, sameTo(test)) ||
		!near(Num(sigma), read.Chandra.Sigma, sameTo(test)) {
		test.Errorf("one sigma %v of stated %v, want %v of %v", sigma,
			stated, read.Chandra.Sigma, read.Chandra.Stated)
	}
	flux := source.Measurements()[none]
	value, isEnergy := ergs(flux.Value())
	if !isEnergy || !near(value, read.Chandra.Flux, sameTo(test)) ||
		flux.Filter() != read.Chandra.Filter {
		test.Errorf("Sgr A*'s flux reads as %+v", flux)
	}
}

// TestChandraFluxErrorIsHalfTheLimitsWidth checks that the flux error is
// half the width between the 68% limits. Limits of 0.9e-13 and 1.1e-13
// give an error of 1e-14, not their 2e-14 width.
func TestChandraFluxErrorIsHalfTheLimitsWidth(test *testing.T) {
	got := chandraSources(test, []byte(chandraHeader+goodRow+"\n"))
	if len(got) != one {
		test.Fatalf("%d sources from the good row", len(got))
	}
	flux := got[none].Measurements()[none]
	sigma, isEnergy := ergs(flux.Error())
	if !isEnergy || !near(sigma, halfTheLimitsWidth, sameTo(test)) {
		test.Errorf("flux error %v, want %v", sigma, halfTheLimitsWidth)
	}
}

// halfTheLimitsWidth is half the width between goodRow's 68% limits,
// 0.9e-13 and 1.1e-13, which are 2e-14 apart.
const halfTheLimitsWidth = 1e-14

// TestNewChandraNeedsItsProfile checks that no profiles, and so no chandra
// profile, give no reader.
func TestNewChandraNeedsItsProfile(test *testing.T) {
	if _, err := read.NewChandra(nil); err == nil {
		test.Error("a reader was made with no profiles")
	}
}

// TestChandraRefusesTables checks that input that is not a table, or lacks
// a column nilgai needs, is yielded once as an error that is not a row's,
// and ends the read.
func TestChandraRefusesTables(test *testing.T) {
	table := chandraTable(test)
	for _, each := range loadSgrA(test).BadChandra {
		var errs []error
		for _, err := range table.Read("t", strings.NewReader(each.Table)) {
			errs = append(errs, err)
		}
		var row *nilgai.RowError
		if len(errs) != one || errs[none] == nil ||
			errors.As(errs[none], &row) {
			test.Errorf("%s: %v, want one error that is not a row's",
				each.Name, errs)
		}
	}
}

// chandraHeader is the header row of a CSC search made for nilgai.
const chandraHeader = "name\tra\tdec\terr_ellipse_r0\terr_ellipse_r1\t" +
	"err_ellipse_ang\tflux_aper_b\tflux_aper_lolim_b\tflux_aper_hilim_b\n"

// goodRow is a row that reads, to show that the reader carries on.
const goodRow = "2CXO J2\t266.4\t-29.0\t0.7\t0.7\t0\t1e-13\t0.9e-13\t1.1e-13"

// TestChandraNamesTheBadCell checks that a row nilgai cannot read is a
// RowError naming its column, and that the row after it is still read.
func TestChandraNamesTheBadCell(test *testing.T) {
	table := chandraTable(test)
	for _, each := range loadSgrA(test).BadRows {
		input := chandraHeader + each.Row + "\n" + goodRow + "\n"
		var errs []error
		var good []nilgai.Source
		for source, err := range table.Read("t", strings.NewReader(input)) {
			if err != nil {
				if source != nil {
					test.Errorf("%q: a source came beside %v",
						each.Row, err)
				}
				errs = append(errs, err)
				continue
			}
			good = append(good, *source)
		}
		var row *nilgai.RowError
		if len(errs) != one || !errors.As(errs[none], &row) ||
			row.Column() != each.Column || row.Row() != one ||
			len(good) != one {
			test.Errorf("%q: %v, want column %q then a good row",
				each.Row, errs, each.Column)
		}
	}
}

// TestChandraStopsWhenAsked checks that the reader does not yield again
// to a caller that has stopped early.
func TestChandraStopsWhenAsked(test *testing.T) {
	input := chandraHeader + goodRow + "\n" + goodRow + "\n"
	seen := none
	for range chandraTable(test).Read("t", strings.NewReader(input)) {
		seen++
		break
	}
	if seen != one {
		test.Errorf("yielded %d times after a stop", seen)
	}
}

// ellipseOf reads one row with the radii and angle given and returns its
// ellipse. It fails the test if the row does not read.
func ellipseOf(test *testing.T, r0, r1, angle Num) nilgai.ErrorEllipse {
	test.Helper()
	row := fmt.Sprintf("2CXO J3\t266.4\t-29.0\t%v\t%v\t%v\t1e-13\t0.9e-13\t"+
		"1.1e-13\n", r0, r1, angle)
	got := chandraSources(test, []byte(chandraHeader+row))
	if len(got) != one {
		test.Fatalf("%d sources from %q", len(got), row)
	}
	return got[none].Error()
}

// TestChandraTurnsTheAngleWithTheAxes checks that the larger radius is the
// major axis, and that when the swap puts the major axis a quarter turn
// from where r0 pointed, the angle turns with it.
//
// TestChandraTurnsTheAngleWithTheAxes uses angles worked by hand: 30 and
// a quarter turn is 120; 300 and a quarter turn is 390, which is 30.
func TestChandraTurnsTheAngleWithTheAxes(test *testing.T) {
	for _, each := range loadSgrA(test).Ellipses {
		ellipse := ellipseOf(test, each.R0, each.R1, each.Angle)
		major := Num(ellipse.StatedSemiMajor())
		minor := Num(ellipse.StatedSemiMinor())
		angle := Num(ellipse.PositionAngle())
		closely := sameTo(test)
		if !near(major, each.Major, closely) ||
			!near(minor, each.Minor, closely) ||
			!near(angle, each.WantAngle, closely) {
			test.Errorf("%s: major %v minor %v angle %v; want %v %v %v",
				each.Name, major, minor, angle,
				each.Major, each.Minor, each.WantAngle)
		}
	}
}

// TestChandraDoesNotTurnWhatItDidNotSwap is an anti-test. A swapped pair
// is not left at the angle it came with, and an unswapped pair is not
// turned.
func TestChandraDoesNotTurnWhatItDidNotSwap(test *testing.T) {
	for _, each := range loadSgrA(test).Ellipses {
		ellipse := ellipseOf(test, each.R0, each.R1, each.Angle)
		swapped := each.R1 > each.R0
		unmoved := Num(ellipse.PositionAngle()) == each.Angle
		if swapped == unmoved {
			test.Errorf("%s: swapped %v but the angle %v came from %v",
				each.Name, swapped, ellipse.PositionAngle(), each.Angle)
		}
	}
}

// TestChandraOffsetSeesCosDec is an anti-test. The offset pin can see
// what the physical bound cannot. The same row measured on a flat sky,
// with no cos(dec), is outside the offset pin and inside the bound, so
// the pin is not too loose to tell the two apart.
func TestChandraOffsetSeesCosDec(test *testing.T) {
	read := loadSgrA(test)
	source := chandraSources(test, realSearch(test, read))
	for _, each := range source {
		if each.ID() != read.Chandra.ID {
			continue
		}
		at := each.Position()
		across := (at.RA() - float64(read.RA)) * arcsecondsPerDegree
		along := (at.Dec() - float64(read.Dec)) * arcsecondsPerDegree
		flat := math.Hypot(across, along)
		if math.Abs(flat-float64(read.Chandra.Offset)) <=
			float64(read.Chandra.OffsetWithin) ||
			flat > float64(read.Chandra.Pin) {
			test.Errorf("a flat-sky offset of %v is not outside the pin "+
				"and inside the bound", flat)
		}
	}
}

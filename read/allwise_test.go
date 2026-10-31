package read_test

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/janearc/nilgai"
	"github.com/janearc/nilgai/band"
	"github.com/janearc/nilgai/read"
)

// allwiseFixture holds testdata/allwise.json.
type allwiseFixture struct {
	Header     string
	Good       string
	ID         string
	RA, Dec    Num
	Filter     string
	EllipseRel Num
	AngleApart Num
	Sgra       struct {
		File    string
		IDs     []string
		Quality string
	}
	Flux struct {
		Mag, Err, Value, Error, Rel Num
	}
	Ellipses []struct {
		Name                    string
		Sigra, Sigdec, Sigradec Num
		Major, Minor, Angle     Num
		AnyAngle                bool
	}
	Flags []struct {
		Name, CC, PH, Quality string
	}
	Limit struct {
		Name, Row  string
		Value, Rel Num
	}
	BadRows []struct {
		Name, Row, Column string
	}
	BadTables []struct{ Name, Table string }
	Stray     string
}

// loadAllwise reads the AllWISE fixture.
func loadAllwise(test *testing.T) allwiseFixture {
	test.Helper()
	var read allwiseFixture
	load(test, "allwise", &read)
	return read
}

// allwiseTable returns an AllWISE reader. allwiseTable fails the test if
// the reader cannot be made.
func allwiseTable(test *testing.T) *read.AllWISE {
	test.Helper()
	table, err := read.NewAllWISE()
	if err != nil {
		test.Fatal(err)
	}
	return table
}

// readAllwise reads a table made of the fixture's header and the rows
// given. It returns the sources and the errors that came of it.
func readAllwise(test *testing.T, read allwiseFixture,
	rows ...string) ([]nilgai.Source, []error) {
	test.Helper()
	text := read.Header + strings.Join(rows, "\n") + "\n"
	return collect(allwiseTable(test), strings.NewReader(text))
}

// onlySource reads the rows and returns the one source they hold. It
// fails the test if they hold anything else.
func onlySource(test *testing.T, read allwiseFixture,
	rows ...string) nilgai.Source {
	test.Helper()
	sources, errs := readAllwise(test, read, rows...)
	if len(sources) != one || len(errs) != none {
		test.Fatalf("%d sources and %v from %q", len(sources), errs, rows)
	}
	return sources[none]
}

// TestAllWISEReadsARow checks that a good row is a source of the AllWISE
// catalogue with:
//   - its position
//   - its designation
//   - its ellipse at one sigma
//   - W1 through the W1 filter's bands, and not W2's
func TestAllWISEReadsARow(test *testing.T) {
	read := loadAllwise(test)
	source := onlySource(test, read, read.Good)
	first, last, err := band.FilterBands(read.Filter)
	if err != nil {
		test.Fatal(err)
	}
	flux := source.Measurements()
	if source.Catalogue() != nilgai.AllWISE || source.ID() != read.ID ||
		Num(source.Position().RA()) != read.RA ||
		Num(source.Position().Dec()) != read.Dec ||
		source.Error().StatedAt() != nilgai.OneSigma() ||
		len(flux) != one || flux[none].Filter() != read.Filter {
		test.Fatalf("read as %+v", source)
	}
	gotFirst, gotLast := flux[none].Bands()
	if gotFirst != first || gotLast != last {
		test.Errorf("W1 covers bands %v to %v, want %v to %v", gotFirst,
			gotLast, first, last)
	}
}

// TestAllWISENeedsItsFilter is an anti-test. The filter edges come from
// data/filters.json, so making a reader is refused when the edges cannot
// be read, and is not refused when they can.
func TestAllWISENeedsItsFilter(test *testing.T) {
	for _, filters := range []string{"{", `{"filters": []}`} {
		_, err := read.ReadAllWISE([]byte(filters))
		if err == nil {
			test.Errorf("a reader was made with filters %q", filters)
		}
	}
	if _, err := read.ReadAllWISE([]byte(goodFilters)); err != nil {
		test.Error(err)
	}
	if _, err := read.NewAllWISE(); err != nil {
		test.Error(err)
	}
}

// TestAllWISEEllipses checks that an ellipse worked from sigra, sigdec
// and the signed square of sigradec comes to the fixture's hand-worked
// axes and angle.
func TestAllWISEEllipses(test *testing.T) {
	read := loadAllwise(test)
	for _, each := range read.Ellipses {
		row := fmt.Sprintf("J1,266.4,-29.0,%v,%v,%v,15,0.1,0000,AAAA",
			each.Sigra, each.Sigdec, each.Sigradec)
		ellipse := onlySource(test, read, row).Error()
		major, minor := Num(ellipse.SemiMajor()), Num(ellipse.SemiMinor())
		angle := ellipse.PositionAngle()
		if !near(major, each.Major, read.EllipseRel) ||
			!near(minor, each.Minor, read.EllipseRel) ||
			(!each.AnyAngle &&
				turnsApart(angle, float64(each.Angle)) >
					float64(read.AngleApart)) {
			test.Errorf("%s: %v by %v at %v; want %v by %v at %v",
				each.Name, major, minor, angle, each.Major, each.Minor,
				each.Angle)
		}
	}
}

// TestAllWISEEllipsesPointTheRightWay is an anti-test. A needle along the
// sky's north is not one along its east, and a tilt to the north-east is
// not one to the south-east.
func TestAllWISEEllipsesPointTheRightWay(test *testing.T) {
	read := loadAllwise(test)
	angles := map[string]float64{}
	for _, each := range read.Ellipses {
		row := fmt.Sprintf("J1,266.4,-29.0,%v,%v,%v,15,0.1,0000,AAAA",
			each.Sigra, each.Sigdec, each.Sigradec)
		angles[each.Name] = onlySource(test, read, row).Error().
			PositionAngle()
	}
	apart := turnsApart(angles["north-south needle"],
		angles["east-west needle"])
	tilt := turnsApart(angles["tilted north-east"],
		angles["tilted south-east"])
	if apart < float64(read.AngleApart) || tilt < float64(read.AngleApart) {
		test.Errorf("needles %v apart and tilts %v apart", apart, tilt)
	}
}

// TestAllWISEDetection checks that a magnitude with its error is a
// detection, with the flux and error worked by hand, in Jy.
func TestAllWISEDetection(test *testing.T) {
	read := loadAllwise(test)
	flux := onlySource(test, read, read.Good).Measurements()[none]
	value, valueIsDensity := janskys(flux.Value())
	sigma, sigmaIsDensity := janskys(flux.Error())
	if flux.IsUpperLimit() || flux.Kind() != nilgai.Detection ||
		!valueIsDensity || !sigmaIsDensity {
		test.Fatalf("W1 read as %+v, not a flux-density detection", flux)
	}
	if !near(value, read.Flux.Value, read.Flux.Rel) ||
		!near(sigma, read.Flux.Error, read.Flux.Rel) {
		test.Errorf("W1 read as %+v, want %v and %v", flux,
			read.Flux.Value, read.Flux.Error)
	}
}

// TestAllWISEUpperLimit is an anti-test. A magnitude with no error is an
// upper limit and has no error, and its limit is the flux of that
// magnitude.
func TestAllWISEUpperLimit(test *testing.T) {
	read := loadAllwise(test)
	flux := onlySource(test, read, read.Limit.Row).Measurements()[none]
	limit, isDensity := janskys(flux.Value())
	if !flux.IsUpperLimit() || flux.Kind() != nilgai.UpperLimit ||
		flux.Error() != nil || !isDensity ||
		!near(limit, read.Limit.Value, read.Limit.Rel) {
		test.Errorf("%s read as %+v, want a limit of %v", read.Limit.Name,
			flux, read.Limit.Value)
	}
}

// TestAllWISEReadsItsFlags checks that WISE's flags are kept as they
// came, and that W1's contamination flag, the first character of
// cc_flags, sets the quality. An artefact in W1 is doubtful and a clean W1
// is good.
func TestAllWISEReadsItsFlags(test *testing.T) {
	read := loadAllwise(test)
	for _, each := range read.Flags {
		row := fmt.Sprintf("J1,266.4,-29.0,0.2,0.2,0,15,0.1,%s,%s",
			each.CC, each.PH)
		flux := onlySource(test, read, row).Measurements()[none]
		flags := flux.Flags()
		if flux.Quality().String() != each.Quality || len(flags) != halves ||
			flags[none].Name() != "cc_flags" ||
			flags[none].Value() != each.CC ||
			flags[one].Name() != "ph_qual" ||
			flags[one].Value() != each.PH {
			test.Errorf("%s: %v with flags %v, want %s %q %q", each.Name,
				flux.Quality(), flags, each.Quality, each.CC, each.PH)
		}
	}
}

// TestAllWISENamesTheBadCell checks that a row that cannot be read is a
// RowError naming the column, and that the row after it is still read.
func TestAllWISENamesTheBadCell(test *testing.T) {
	read := loadAllwise(test)
	for _, each := range read.BadRows {
		sources, errs := readAllwise(test, read, each.Row, read.Good)
		var row *nilgai.RowError
		if len(errs) != one || !errors.As(errs[none], &row) ||
			row.Column() != each.Column || row.Row() != one ||
			len(sources) != one {
			test.Errorf("%s: %v, want column %q then a good row",
				each.Name, errs, each.Column)
		}
	}
}

// TestAllWISERefusesTables checks that a table without a column nilgai
// needs, or with no columns at all, is yielded once as an error that is
// not a row's, and ends the read.
func TestAllWISERefusesTables(test *testing.T) {
	read := loadAllwise(test)
	for _, each := range read.BadTables {
		sources, errs := collect(allwiseTable(test),
			strings.NewReader(each.Table))
		var row *nilgai.RowError
		if len(errs) != one || errors.As(errs[none], &row) ||
			len(sources) != none {
			test.Errorf("%s: %v, want one error that is not a row's",
				each.Name, errs)
		}
	}
}

// TestAllWISEStopsAtATornRow checks that a stray quote in the middle of a
// table ends the read with one error that is not a row's. The rows before
// it are read and none after.
func TestAllWISEStopsAtATornRow(test *testing.T) {
	read := loadAllwise(test)
	sources, errs := readAllwise(test, read, read.Good, read.Stray,
		read.Good)
	var row *nilgai.RowError
	if len(sources) != one || len(errs) != one ||
		errors.As(errs[none], &row) {
		test.Errorf("%d sources and %v, want one then an error that is "+
			"not a row's", len(sources), errs)
	}
}

// TestAllWISEStopsWhenAsked checks that the reader does not yield again
// to a caller that has stopped early.
func TestAllWISEStopsWhenAsked(test *testing.T) {
	read := loadAllwise(test)
	text := read.Header + read.Good + "\n" + read.Good + "\n"
	seen := none
	for range allwiseTable(test).Read("t", strings.NewReader(text)) {
		seen++
		break
	}
	if seen != one {
		test.Errorf("yielded %d times after a stop", seen)
	}
}

// TestAllWISEStopsAtABadRowWhenAsked is an anti-test. The reader does not
// yield again to a caller that stops at a row it cannot read, either.
func TestAllWISEStopsAtABadRowWhenAsked(test *testing.T) {
	read := loadAllwise(test)
	text := read.Header + read.BadRows[none].Row + "\n" + read.Good + "\n"
	seen := none
	for range allwiseTable(test).Read("t", strings.NewReader(text)) {
		seen++
		break
	}
	if seen != one {
		test.Errorf("yielded %d times after a stop", seen)
	}
}

// TestAllWISEReadsTheRealSearch checks that the real search around Sgr A*
// reads whole, and that its sources are doubtful in W1, as their
// contamination flags say.
func TestAllWISEReadsTheRealSearch(test *testing.T) {
	read := loadAllwise(test)
	file, err := fixtures.ReadFile("testdata/" + read.Sgra.File)
	if err != nil {
		test.Fatal(err)
	}
	search := bytes.NewReader(file)
	sources, errs := collect(allwiseTable(test), search)
	if len(errs) != none || len(sources) != len(read.Sgra.IDs) {
		test.Fatalf("%d sources and %v", len(sources), errs)
	}
	for place, source := range sources {
		quality := source.Measurements()[none].Quality().String()
		if source.ID() != read.Sgra.IDs[place] ||
			quality != read.Sgra.Quality {
			test.Errorf("%s is %s, want %s %s", source.ID(), quality,
				read.Sgra.IDs[place], read.Sgra.Quality)
		}
	}
}

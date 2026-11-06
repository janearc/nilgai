package read_test

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/janearc/nilgai"
	"github.com/janearc/nilgai/band"
)

// bandsFixture holds testdata/allwise-bands.json.
type bandsFixture struct {
	Header, Good string
	Rel          Num
	Bands        []struct {
		Filter       string
		Value, Error Num
	}
	LimitRow   string
	LimitValue Num
	Absent     []struct {
		Name, Row string
		Filters   []string
	}
	Qualities []struct {
		Name, CC string
		Want     []string
	}
	OnlyW2 struct {
		Header, Row string
		Filters     []string
	}
	BadCells []struct{ Name, Column, Cell string }
	Real     struct {
		File    string
		Rel     Num
		Sources []struct {
			ID          string
			W2Limit, W3 Num
			W3Error     Num
		}
	}
}

// bandsFirstAt and bandsLastAt are the places in a row of the first and
// last extra columns a search for four bands selects, counting from the
// designation at zero.
const (
	// bandsFirstAt is the place of the W2 magnitude, the ninth cell.
	bandsFirstAt = 8
	// bandsLastAt is the place of the W4 magnitude's error, the
	// fourteenth cell.
	bandsLastAt = 13
)

// loadBands reads the four-band fixture.
func loadBands(test *testing.T) bandsFixture {
	test.Helper()
	var read bandsFixture
	load(test, "allwise-bands", &read)
	return read
}

// readBands reads a table made of the header and the rows given.
func readBands(test *testing.T, header string,
	rows ...string) ([]nilgai.Source, []error) {
	test.Helper()
	text := header + strings.Join(rows, "\n") + "\n"
	return collect(allwiseTable(test), strings.NewReader(text))
}

// filtersOf returns the filters of a source's measurements, in order.
func filtersOf(source nilgai.Source) []string {
	var names []string
	for _, each := range source.Measurements() {
		names = append(names, each.Filter())
	}
	return names
}

// TestAllWISEReadsFourBands checks that a row with four bands is a source
// with four measurements, W1 first. Each measurement has:
//   - its own filter's libtheme bands
//   - its flux and error worked by hand, in Jy
//
// W1's bands and W1's zero point must not be used on every band.
func TestAllWISEReadsFourBands(test *testing.T) {
	read := loadBands(test)
	sources, errs := readBands(test, read.Header, read.Good)
	if len(sources) != one || len(errs) != none {
		test.Fatalf("%d sources and %v", len(sources), errs)
	}
	got := sources[none].Measurements()
	if len(got) != len(read.Bands) {
		test.Fatalf("%d measurements, want %d", len(got), len(read.Bands))
	}
	for place, want := range read.Bands {
		first, last, err := band.FilterBands(want.Filter)
		if err != nil {
			test.Fatal(err)
		}
		gotFirst, gotLast := got[place].Bands()
		value, valueIsDensity := janskys(got[place].Value())
		sigma, sigmaIsDensity := janskys(got[place].Error())
		if got[place].Filter() != want.Filter || gotFirst != first ||
			gotLast != last || got[place].Kind() != nilgai.Detection ||
			!valueIsDensity || !sigmaIsDensity ||
			!near(value, want.Value, read.Rel) ||
			!near(sigma, want.Error, read.Rel) {
			test.Errorf("%s read as %+v, want %v and %v in bands %v to %v",
				want.Filter, got[place], want.Value, want.Error, first,
				last)
		}
	}
}

// TestAllWISEBandsAreEachTheirOwn is an anti-test. It checks that no band
// is read through another band's libtheme bands.
func TestAllWISEBandsAreEachTheirOwn(test *testing.T) {
	read := loadBands(test)
	source := readOne(test, read.Header, read.Good)
	seen := map[string]string{}
	for _, each := range source.Measurements() {
		first, last := each.Bands()
		key := fmt.Sprint(first, last)
		if other, twice := seen[key]; twice {
			test.Errorf("%s and %s cover the same bands", each.Filter(),
				other)
		}
		seen[key] = each.Filter()
	}
}

// readOne reads the rows and returns the one source they hold. It fails
// the test if they hold anything else.
func readOne(test *testing.T, header string, rows ...string) nilgai.Source {
	test.Helper()
	sources, errs := readBands(test, header, rows...)
	if len(sources) != one || len(errs) != none {
		test.Fatalf("%d sources and %v from %q", len(sources), errs, rows)
	}
	return sources[none]
}

// TestAllWISEBandUpperLimit checks that a band with a magnitude and no
// error is an upper limit with no error, in W4 as in W1.
func TestAllWISEBandUpperLimit(test *testing.T) {
	read := loadBands(test)
	got := readOne(test, read.Header, read.LimitRow).Measurements()
	last := got[len(got)-one]
	limit, isDensity := janskys(last.Value())
	if last.Filter() != "W4" || !last.IsUpperLimit() || !isDensity ||
		last.Error() != nil || !near(limit, read.LimitValue, read.Rel) {
		test.Errorf("W4 read as %+v, want a limit of %v", last,
			read.LimitValue)
	}
}

// An optional band with both cells empty is omitted without rejecting the row.
// Later populated bands must still be read.
func TestAllWISELeavesOutABandItHasNoMeasurementIn(test *testing.T) {
	read := loadBands(test)
	for _, each := range read.Absent {
		got := filtersOf(readOne(test, read.Header, each.Row))
		if strings.Join(got, " ") != strings.Join(each.Filters, " ") {
			test.Errorf("%s: measurements in %v, want %v", each.Name, got,
				each.Filters)
		}
	}
}

// Each band’s quality uses its own cc_flags character. Contamination in W2 must
// not change W1’s quality, and each measurement retains the full flag strings.
func TestAllWISEReadsEachBandsContamination(test *testing.T) {
	read := loadBands(test)
	for _, each := range read.Qualities {
		row := withCell(read.Good, ",", bandsLastAt+one, each.CC)
		got := readOne(test, read.Header, row).Measurements()
		for place, want := range each.Want {
			flags := got[place].Flags()
			if got[place].Quality().String() != want || len(flags) != halves ||
				flags[none].Value() != each.CC {
				test.Errorf("%s: %s is %v with flags %v, want %s",
					each.Name, got[place].Filter(), got[place].Quality(),
					flags, want)
			}
		}
	}
}

// W2 can be selected without W3 or W4. This preserves support for searches
// exported before the reader handled all four bands.
func TestAllWISEReadsASearchOfOnlyW2(test *testing.T) {
	read := loadBands(test)
	got := filtersOf(readOne(test, read.OnlyW2.Header, read.OnlyW2.Row))
	if strings.Join(got, " ") != strings.Join(read.OnlyW2.Filters, " ") {
		test.Errorf("measurements in %v, want %v", got,
			read.OnlyW2.Filters)
	}
}

// TestAllWISENamesEachBandColumnItLacks is an anti-test. A search that
// selects one of a band's two columns and not the other is refused, and
// the error names the column it lacks.
func TestAllWISENamesEachBandColumnItLacks(test *testing.T) {
	read := loadBands(test)
	header := strings.TrimSuffix(read.Header, "\n")
	columns := strings.Split(header, ",")
	for place := bandsFirstAt; place <= bandsLastAt; place++ {
		input := withoutCell(header, ",", place) + "\n" +
			withoutCell(read.Good, ",", place) + "\n"
		sources, errs := collect(allwiseTable(test), strings.NewReader(input))
		if !namesAColumn(sources, errs, columns[place]) {
			test.Errorf("without %s: %d sources and %v", columns[place],
				len(sources), errs)
		}
	}
}

// TestAllWISENamesTheBadBandCell is an anti-test. Each of these band cells
// is a row error that names the column, and the row after it is still
// read:
//   - a cell that is not a number
//   - a cell that is empty while its partner is not
//   - an error cell that is negative
//   - a cell that is not finite
func TestAllWISENamesTheBadBandCell(test *testing.T) {
	read := loadBands(test)
	columns := strings.Split(strings.TrimSuffix(read.Header, "\n"), ",")
	for _, each := range read.BadCells {
		place := indexOf(columns, each.Column)
		row := withCell(read.Good, ",", place, each.Cell)
		sources, errs := readBands(test, read.Header, row, read.Good)
		var rowErr *nilgai.RowError
		if len(errs) != one || !errors.As(errs[none], &rowErr) ||
			rowErr.Column() != each.Column || rowErr.Row() != one ||
			len(sources) != one {
			test.Errorf("%s: %v, want column %q then a good row", each.Name,
				errs, each.Column)
		}
	}
}

// indexOf returns the place of the name in the names, or -1.
func indexOf(names []string, name string) int {
	for place, each := range names {
		if each == name {
			return place
		}
	}
	return -one
}

// The real Sgr A* search must produce W2 upper limits, W3 detections, and no W4
// measurements, following the populated cells in each row.
func TestAllWISEReadsTheRealFourBandSearch(test *testing.T) {
	read := loadBands(test)
	file, err := fixtures.ReadFile("testdata/" + read.Real.File)
	if err != nil {
		test.Fatal(err)
	}
	sources, errs := collect(allwiseTable(test), bytes.NewReader(file))
	if len(errs) != none || len(sources) != len(read.Real.Sources) {
		test.Fatalf("%d sources and %v", len(sources), errs)
	}
	for place, want := range read.Real.Sources {
		got := sources[place].Measurements()
		if sources[place].ID() != want.ID ||
			strings.Join(filtersOf(sources[place]), " ") != "W1 W2 W3" {
			test.Errorf("%s: measurements in %v", sources[place].ID(),
				filtersOf(sources[place]))
			continue
		}
		w2, w3 := got[one], got[halves]
		limit, limitIsDensity := janskys(w2.Value())
		value, valueIsDensity := janskys(w3.Value())
		sigma, sigmaIsDensity := janskys(w3.Error())
		rel := read.Real.Rel
		if !limitIsDensity || !valueIsDensity || !sigmaIsDensity ||
			!w2.IsUpperLimit() || !near(limit, want.W2Limit, rel) ||
			w3.IsUpperLimit() || !near(value, want.W3, rel) ||
			!near(sigma, want.W3Error, rel) {
			test.Errorf("%s: W2 %+v, W3 %+v", want.ID, w2, w3)
		}
	}
}

// The earlier Sgr A* export includes W2 columns. They must now be read as upper
// limits because their uncertainty cells are empty.
func TestAllWISEReadsTheEarlierRealSearchsW2(test *testing.T) {
	read := loadAllwise(test)
	file, err := fixtures.ReadFile("testdata/" + read.Sgra.File)
	if err != nil {
		test.Fatal(err)
	}
	sources, errs := collect(allwiseTable(test), bytes.NewReader(file))
	if len(errs) != none || len(sources) == none {
		test.Fatalf("%d sources and %v", len(sources), errs)
	}
	for _, source := range sources {
		got := source.Measurements()
		if len(got) != halves || got[one].Filter() != "W2" ||
			!got[one].IsUpperLimit() {
			test.Errorf("%s: measurements in %v", source.ID(),
				filtersOf(source))
		}
	}
}

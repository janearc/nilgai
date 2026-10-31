package read_test

import (
	"errors"
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/janearc/libtheme-css/spaces/radiation"
	"github.com/janearc/nilgai"
	"github.com/janearc/nilgai/band"
	"github.com/janearc/nilgai/read"
)

// chandraRAAt and the constants with it are places in a Chandra row,
// counting from the name at zero.
const (
	// chandraRAAt is the place of the right ascension, the second cell.
	chandraRAAt = 1
	// chandraDecAt is the place of the declination, the third cell.
	chandraDecAt = 2
	// chandraR0At is the place of the ellipse's first radius, the fourth
	// cell.
	chandraR0At = 3
	// chandraR1At is the place of the ellipse's second radius, the fifth
	// cell.
	chandraR1At = 4
	// chandraAngleAt is the place of the ellipse's angle, the sixth cell.
	chandraAngleAt = 5
	// chandraFluxAt is the place of the broad-band flux, the seventh cell.
	chandraFluxAt = 6
	// chandraLowAt is the place of the flux's lower limit, the eighth
	// cell.
	chandraLowAt = 7
	// chandraHighAt is the place of the flux's upper limit, the ninth
	// cell.
	chandraHighAt = 8
)

// allwiseRAAt and the constants with it are places in an AllWISE row,
// counting from the designation at zero.
const (
	// allwiseRAAt is the place of the right ascension, the second cell.
	allwiseRAAt = 1
	// allwiseDecAt is the place of the declination, the third cell.
	allwiseDecAt = 2
	// allwiseSigraAt is the place of the error along RA, the fourth cell.
	allwiseSigraAt = 3
	// allwiseSigdecAt is the place of the error along Dec, the fifth cell.
	allwiseSigdecAt = 4
	// allwiseCosigmaAt is the place of the co-sigma, the sixth cell.
	allwiseCosigmaAt = 5
	// allwiseMagnitudeAt is the place of the W1 magnitude, the seventh
	// cell.
	allwiseMagnitudeAt = 6
)

// pair is a test case of two cells, with the name the failure is reported
// under.
type pair struct{ name, first, second string }

// triple is a test case of three cells, with the name the failure is
// reported under.
type triple struct{ name, first, second, third string }

// withCell returns the row with the cell at the place changed.
func withCell(row, separator string, place int, cell string) string {
	cells := strings.Split(row, separator)
	cells[place] = cell
	return strings.Join(cells, separator)
}

// withoutCell returns the row with the cell at the place left out.
func withoutCell(row, separator string, place int) string {
	cells := strings.Split(row, separator)
	return strings.Join(append(cells[:place:place], cells[place+one:]...),
		separator)
}

// namesAColumn says whether the read ended in one error that names the
// column and is not a row's, with no source read.
func namesAColumn(sources []nilgai.Source, errs []error, column string) bool {
	var row *nilgai.RowError
	return len(sources) == none && len(errs) == one &&
		!errors.As(errs[none], &row) &&
		strings.Contains(errs[none].Error(), column)
}

// namesACell says whether the read gave one row error naming the column,
// and no source.
func namesACell(sources []nilgai.Source, errs []error, column string) bool {
	var row *nilgai.RowError
	return len(sources) == none && len(errs) == one &&
		errors.As(errs[none], &row) && row.Column() == column
}

// TestChandraNamesEachColumnItLacks is an anti-test. A Chandra search that
// lacks any one of the columns it must select is refused, and the error
// names the column.
func TestChandraNamesEachColumnItLacks(test *testing.T) {
	header := strings.TrimSuffix(chandraHeader, "\n")
	for place, column := range strings.Split(header, "\t") {
		input := withoutCell(header, "\t", place) + "\n" +
			withoutCell(goodRow, "\t", place) + "\n"
		sources, errs := collect(chandraTable(test), strings.NewReader(input))
		if !namesAColumn(sources, errs, column) {
			test.Errorf("without %s: %d sources and %v", column,
				len(sources), errs)
		}
	}
}

// TestChandraNamesEachCellItCannotRead is an anti-test. A cell that the
// Chandra reader needs as a number, left empty or written as something
// that is not a number, is a row error that names its column. The name
// cell is text, and is read as it comes.
func TestChandraNamesEachCellItCannotRead(test *testing.T) {
	columns := strings.Split(strings.TrimSuffix(chandraHeader, "\n"), "\t")
	for place := chandraRAAt; place <= chandraHighAt; place++ {
		for _, cell := range []string{"", "north"} {
			input := chandraHeader + withCell(goodRow, "\t", place, cell) +
				"\n"
			sources, errs := collect(chandraTable(test),
				strings.NewReader(input))
			if !namesACell(sources, errs, columns[place]) {
				test.Errorf("%s as %q: %d sources and %v", columns[place],
					cell, len(sources), errs)
			}
		}
	}
}

// TestChandraSaysAnEmptyCellIsEmpty is an anti-test. An empty cell is
// refused as empty, because the catalogue has no value there, and not as
// the parse error that an empty string would give. A cell that is text is
// not called empty.
func TestChandraSaysAnEmptyCellIsEmpty(test *testing.T) {
	for place := chandraRAAt; place <= chandraHighAt; place++ {
		for cell, want := range map[string]bool{"": true, "north": false} {
			input := chandraHeader + withCell(goodRow, "\t", place, cell) +
				"\n"
			_, errs := collect(chandraTable(test), strings.NewReader(input))
			if len(errs) != one {
				test.Fatalf("cell %d as %q: %v", place, cell, errs)
			}
			said := strings.Contains(errs[none].Error(), "empty")
			if said != want {
				test.Errorf("cell %d as %q: %q, says empty is %v", place,
					cell, errs[none], said)
			}
		}
	}
}

// TestChandraRefusesEllipsesTheRecordRefuses is an anti-test. The Chandra
// reader refuses an ellipse that the nilgai record refuses, as a row error
// on the ellipse's column, and does not pass it on as it is.
func TestChandraRefusesEllipsesTheRecordRefuses(test *testing.T) {
	for _, each := range []triple{
		{"major NaN", "NaN", "0.7", "0"},
		{"major Inf", "Inf", "0.7", "0"},
		{"minor negative", "0.7", "-0.7", "0"},
		{"minor NaN", "0.7", "NaN", "0"},
		{"both radii none", "0", "0", "0"},
	} {
		row := withCell(goodRow, "\t", chandraR0At, each.first)
		row = withCell(row, "\t", chandraR1At, each.second)
		row = withCell(row, "\t", chandraAngleAt, each.third)
		sources, errs := collect(chandraTable(test),
			strings.NewReader(chandraHeader+row+"\n"))
		if !namesACell(sources, errs, "err_ellipse_r0") {
			test.Errorf("%s: %d sources and %v", each.name, len(sources),
				errs)
		}
	}
}

// TestChandraRefusesAnAngleOutsideTheCircleEitherWay is an anti-test. An
// angle outside [0, 360) is refused on err_ellipse_ang whichever radius is
// the larger. When r1 is the larger, the axes are swapped and the angle
// turned a quarter turn. That turn must not bring a negative angle into
// range, or bring an angle past a full turn back into it.
func TestChandraRefusesAnAngleOutsideTheCircleEitherWay(test *testing.T) {
	radii := []pair{
		{"r0 above r1", "0.7", "0.6"},
		{"r0 equal to r1", "0.7", "0.7"},
		{"r0 below r1", "0.6", "0.7"},
	}
	for _, angle := range []string{"-10", "-0.5", "360", "725", "NaN",
		"Inf", "-Inf"} {
		for _, each := range radii {
			row := withCell(goodRow, "\t", chandraR0At, each.first)
			row = withCell(row, "\t", chandraR1At, each.second)
			row = withCell(row, "\t", chandraAngleAt, angle)
			sources, errs := collect(chandraTable(test),
				strings.NewReader(chandraHeader+row+"\n"))
			if !namesACell(sources, errs, "err_ellipse_ang") {
				test.Errorf("angle %s, %s: %d sources and %v", angle,
					each.name, len(sources), errs)
			}
		}
	}
}

// TestChandraRefusesPositionsTheRecordRefuses is an anti-test. A Chandra
// position that the record refuses is a row error on ra, whichever of ra
// and dec is at fault.
func TestChandraRefusesPositionsTheRecordRefuses(test *testing.T) {
	for _, each := range []pair{
		{"ra a full circle", "360", "-29.0"},
		{"ra negative", "-1", "-29.0"},
		{"ra NaN", "NaN", "-29.0"},
		{"ra Inf", "Inf", "-29.0"},
		{"dec past north", "266.4", "90.5"},
		{"dec past south", "266.4", "-90.5"},
		{"dec NaN", "266.4", "NaN"},
		{"dec -Inf", "266.4", "-Inf"},
	} {
		row := withCell(goodRow, "\t", chandraRAAt, each.first)
		row = withCell(row, "\t", chandraDecAt, each.second)
		sources, errs := collect(chandraTable(test),
			strings.NewReader(chandraHeader+row+"\n"))
		if !namesACell(sources, errs, "ra") {
			test.Errorf("%s: %d sources and %v", each.name, len(sources),
				errs)
		}
	}
}

// TestChandraRefusesFluxesTheRecordRefuses is an anti-test. A Chandra flux
// that the record refuses is a row error on the broad-band flux's column.
// TestChandraRefusesFluxesTheRecordRefuses feeds fluxes below zero and
// fluxes that are not finite.
func TestChandraRefusesFluxesTheRecordRefuses(test *testing.T) {
	for _, each := range []triple{
		{"flux negative", "-1e-13", "0.9e-13", "1.1e-13"},
		{"flux NaN", "NaN", "0.9e-13", "1.1e-13"},
		{"flux Inf", "Inf", "0.9e-13", "1.1e-13"},
		{"upper limit Inf", "1e-13", "0.9e-13", "Inf"},
		{"lower limit NaN", "1e-13", "NaN", "1.1e-13"},
		{"limits crossed", "1e-13", "1.1e-13", "0.9e-13"},
	} {
		row := withCell(goodRow, "\t", chandraFluxAt, each.first)
		row = withCell(row, "\t", chandraLowAt, each.second)
		row = withCell(row, "\t", chandraHighAt, each.third)
		sources, errs := collect(chandraTable(test),
			strings.NewReader(chandraHeader+row+"\n"))
		if !namesACell(sources, errs, "flux_aper_b") {
			test.Errorf("%s: %d sources and %v", each.name, len(sources),
				errs)
		}
	}
}

// TestAllWISENamesEachColumnItLacks is an anti-test. An AllWISE search
// that lacks any one of the columns it must select is refused, and the
// error names the column.
func TestAllWISENamesEachColumnItLacks(test *testing.T) {
	fixture := loadAllwise(test)
	header := strings.TrimSuffix(fixture.Header, "\n")
	for place, column := range strings.Split(header, ",") {
		input := withoutCell(header, ",", place) + "\n" +
			withoutCell(fixture.Good, ",", place) + "\n"
		sources, errs := collect(allwiseTable(test), strings.NewReader(input))
		if !namesAColumn(sources, errs, column) {
			test.Errorf("without %s: %d sources and %v", column,
				len(sources), errs)
		}
	}
}

// TestAllWISENamesEachCellItCannotRead is an anti-test. A cell that the
// AllWISE reader needs as a number, left empty or written as something
// that is not a number, is a row error that names its column. The
// designation is text, and the flag columns are kept as they come.
func TestAllWISENamesEachCellItCannotRead(test *testing.T) {
	fixture := loadAllwise(test)
	columns := strings.Split(strings.TrimSuffix(fixture.Header, "\n"), ",")
	for place := allwiseRAAt; place <= allwiseMagnitudeAt; place++ {
		for _, cell := range []string{"", "north"} {
			row := withCell(fixture.Good, ",", place, cell)
			sources, errs := readAllwise(test, fixture, row)
			if !namesACell(sources, errs, columns[place]) {
				test.Errorf("%s as %q: %d sources and %v", columns[place],
					cell, len(sources), errs)
			}
		}
	}
}

// TestAllWISERefusesErrorsNoEllipseCouldHave is an anti-test. An ellipse
// that the record refuses, because the errors are not numbers, is a row
// error on sigra and not an ellipse of nothing.
func TestAllWISERefusesErrorsNoEllipseCouldHave(test *testing.T) {
	fixture := loadAllwise(test)
	for _, each := range []triple{
		{"sigra NaN", "NaN", "0.2", "0"},
		{"cosigma NaN", "0.2", "0.2", "NaN"},
		{"sigra Inf", "Inf", "0.2", "0"},
		{"cosigma Inf", "0.2", "0.2", "Inf"},
		{"cosigma -Inf", "0.2", "0.2", "-Inf"},
		{"both sigmas none", "0", "0", "0"},
		{"cosigma too big", "0.1", "0.1", "0.5"},
	} {
		row := withCell(fixture.Good, ",", allwiseSigraAt, each.first)
		row = withCell(row, ",", allwiseSigdecAt, each.second)
		row = withCell(row, ",", allwiseCosigmaAt, each.third)
		sources, errs := readAllwise(test, fixture, row)
		if !namesACell(sources, errs, "sigra") {
			test.Errorf("%s: %d sources and %v", each.name, len(sources),
				errs)
		}
	}
}

// TestAllWISERefusesAnErrorBelowNoneOnItsOwnColumn is an anti-test. sigra
// and sigdec are squared, so a negative one would read as its positive
// twin. Each is refused, and the row error names that column, as it does
// for w1sigmpro.
func TestAllWISERefusesAnErrorBelowNoneOnItsOwnColumn(test *testing.T) {
	fixture := loadAllwise(test)
	for _, bad := range []string{"-0.2", "-1e-9", "NaN", "Inf", "-Inf"} {
		for _, each := range []struct {
			column string
			place  int
		}{{"sigra", allwiseSigraAt}, {"sigdec", allwiseSigdecAt}} {
			row := withCell(fixture.Good, ",", each.place, bad)
			sources, errs := readAllwise(test, fixture, row)
			if !namesACell(sources, errs, each.column) {
				test.Errorf("%s as %s: %d sources and %v", each.column,
					bad, len(sources), errs)
			}
		}
	}
}

// TestAllWISERefusesPositionsTheRecordRefuses is an anti-test. An AllWISE
// position that the record refuses is a row error on ra.
func TestAllWISERefusesPositionsTheRecordRefuses(test *testing.T) {
	fixture := loadAllwise(test)
	for _, each := range []pair{
		{"ra a full circle", "360", "-29.0"},
		{"ra negative", "-1", "-29.0"},
		{"ra NaN", "NaN", "-29.0"},
		{"dec past north", "266.4", "90.5"},
		{"dec past south", "266.4", "-90.5"},
		{"dec Inf", "266.4", "Inf"},
	} {
		row := withCell(fixture.Good, ",", allwiseRAAt, each.first)
		row = withCell(row, ",", allwiseDecAt, each.second)
		sources, errs := readAllwise(test, fixture, row)
		if !namesACell(sources, errs, "ra") {
			test.Errorf("%s: %d sources and %v", each.name, len(sources),
				errs)
		}
	}
}

// TestFermiNamesEachColumnItLacks is an anti-test. A Fermi file that lacks
// any one of the columns it must have is refused, and the error names the
// column.
func TestFermiNamesEachColumnItLacks(test *testing.T) {
	fixture := loadFermi(test)
	for _, column := range fixture.Columns {
		file := fermiFile(test, fixture.Table,
			withoutColumn(fixture, column.Name), fixture.Good)
		sources, errs := readFermiFile(test, file)
		if !namesAColumn(sources, errs, column.Name) {
			test.Errorf("without %s: %d sources and %v", column.Name,
				len(sources), errs)
		}
	}
}

// onlyW1 is filters JSON with W1 and no other filter.
const onlyW1 = `{"filters": [{"name": "W1", "from": {"value": 2.8,
	"unit": "um"}, "to": {"value": 3.9, "unit": "um"}}]}`

// onlyFermi is filters JSON with the Fermi energy flux and no other
// filter.
const onlyFermi = `{"filters": [{"name": "Energy_Flux100", "from":
	{"value": 100, "unit": "MeV"}, "to": {"value": 100, "unit": "GeV"}}]}`

// TestReadersNeedTheirOwnFilter is an anti-test. A reader is made from
// filters that include its own filter, and not from filters that have only
// another reader's.
func TestReadersNeedTheirOwnFilter(test *testing.T) {
	if _, err := read.ReadFermi([]byte(onlyW1)); err == nil {
		test.Error("a Fermi reader was made with only W1")
	}
	if _, err := read.ReadAllWISE([]byte(onlyFermi)); err == nil {
		test.Error("an AllWISE reader was made with only the Fermi filter")
	}
	if _, err := read.ReadAllWISE([]byte(onlyW1)); err == nil {
		test.Error("an AllWISE reader was made with only W1")
	}
	if _, err := read.ReadFermi([]byte(onlyFermi)); err != nil {
		test.Error(err)
	}
}

// filtersNamed returns filters JSON with the names given, each filter one
// micron wide.
func filtersNamed(names ...string) []byte {
	var each []string
	for _, name := range names {
		each = append(each, fmt.Sprintf(`{"name": %q, "from": {"value": 1,
			"unit": "um"}, "to": {"value": 2, "unit": "um"}}`, name))
	}
	return []byte(`{"filters": [` + strings.Join(each, ",") + `]}`)
}

// TestAllWISENeedsEachBandsFilter is an anti-test. An AllWISE reader needs
// the filter of each of its four bands, and the error names the filter
// that is missing.
func TestAllWISENeedsEachBandsFilter(test *testing.T) {
	all := []string{"W1", "W2", "W3", "W4"}
	if _, err := read.ReadAllWISE(filtersNamed(all...)); err != nil {
		test.Fatal(err)
	}
	for place, missing := range all {
		without := slices.Delete(slices.Clone(all), place, place+one)
		_, err := read.ReadAllWISE(filtersNamed(without...))
		if err == nil || !strings.Contains(err.Error(), `"`+missing+`"`) {
			test.Errorf("without %s: %v", missing, err)
		}
	}
}

// TestNewChandraRefusesProfilesThatCannotGiveABand is an anti-test. A
// Chandra reader is not made from a chandra profile with no width, nor
// from profiles with no chandra profile among them.
func TestNewChandraRefusesProfilesThatCannotGiveABand(test *testing.T) {
	profiles, err := band.Profiles()
	if err != nil {
		test.Fatal(err)
	}
	var others, narrowed []radiation.Profile
	for _, each := range profiles {
		if each.Name != "chandra" {
			others = append(others, each)
			continue
		}
		each.Source.Hi = each.Source.Lo
		narrowed = append(narrowed, each)
	}
	if len(narrowed) != one {
		test.Fatalf("%d chandra profiles", len(narrowed))
	}
	if _, err := read.NewChandra(narrowed); err == nil {
		test.Error("a reader was made from a band of no width")
	}
	if _, err := read.NewChandra(others); err == nil {
		test.Error("a reader was made with no chandra profile")
	}
}

package read_test

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/astrogo/fitsio"
	"github.com/janearc/nilgai"
	"github.com/janearc/nilgai/band"
	"github.com/janearc/nilgai/internal/fitstest"
	"github.com/janearc/nilgai/read"
)

// columnSpec describes a column of the table as the fixture names it.
type columnSpec struct{ Name, Format string }

// fermiRowSpec describes a row as the fixture writes it.
type fermiRowSpec struct {
	Name                                          string
	RA, Dec, Major, Minor, Angle, Flux, FluxError Num
	Class, Assoc                                  string
	Flags                                         int16
}

// fermiCase pairs a row that must read with what it must read as.
type fermiCase struct {
	Name                                 string
	Row                                  fermiRowSpec
	ID                                   string
	Major, Minor, SigmaMajor, SigmaMinor Num
	Angle                                Num
	Quality                              string
	Kept                                 []string
}

// refusedCase describes a row that must be refused. It holds the column
// the row is refused at, and the words its error must and must not say.
type refusedCase struct {
	Name    string
	Row     fermiRowSpec
	Column  string
	Want    []string
	NotWant []string
}

// fermiFixture holds testdata/fermi.json.
type fermiFixture struct {
	Table   string
	Rel     Num
	Stated  Num
	Columns []columnSpec
	Filter  string
	Good    fermiRowSpec
	Sources []fermiCase
	Refused []refusedCase
}

// loadFermi reads the Fermi fixture.
func loadFermi(test *testing.T) fermiFixture {
	test.Helper()
	var read fermiFixture
	load(test, "fermi", &read)
	return read
}

// fermiFile returns a FITS file with one table, of the name and columns
// given, holding the rows. It has the shape of 4FGL's file and is built in
// memory.
func fermiFile(test *testing.T, name string, columns []columnSpec,
	rows ...fermiRowSpec) []byte {
	test.Helper()
	var out bytes.Buffer
	made := make([]fitsio.Column, len(columns))
	written := make([]any, len(rows))
	for place, each := range columns {
		made[place] = fitsio.Column{Name: each.Name, Format: each.Format}
	}
	for place, each := range rows {
		written[place] = fermiRow(each)
	}
	if err := fitstest.Write(&out, name, made, written...); err != nil {
		test.Fatal(err)
	}
	return out.Bytes()
}

// fermiRow returns the row to write to the table, by 4FGL's column names.
func fermiRow(each fermiRowSpec) any {
	return &struct {
		Name      string  `fits:"Source_Name"`
		RA        float32 `fits:"RAJ2000"`
		Dec       float32 `fits:"DEJ2000"`
		Major     float32 `fits:"Conf_95_SemiMajor"`
		Minor     float32 `fits:"Conf_95_SemiMinor"`
		Angle     float32 `fits:"Conf_95_PosAng"`
		Flux      float32 `fits:"Energy_Flux100"`
		FluxError float32 `fits:"Unc_Energy_Flux100"`
		Class     string  `fits:"CLASS1"`
		Assoc     string  `fits:"ASSOC1"`
		Flags     int16   `fits:"Flags"`
	}{each.Name, float32(each.RA), float32(each.Dec), float32(each.Major),
		float32(each.Minor), float32(each.Angle), float32(each.Flux),
		float32(each.FluxError), each.Class, each.Assoc, each.Flags}
}

// fermiTable returns a Fermi reader. fermiTable fails the test if the
// reader cannot be made.
func fermiTable(test *testing.T) *read.Fermi {
	test.Helper()
	table, err := read.NewFermi()
	if err != nil {
		test.Fatal(err)
	}
	return table
}

// readFermi reads a 4FGL-shaped file of the fixture's own columns.
func readFermi(test *testing.T, read fermiFixture,
	rows ...fermiRowSpec) ([]nilgai.Source, []error) {
	test.Helper()
	file := fermiFile(test, read.Table, read.Columns, rows...)
	return readFermiFile(test, file)
}

// readFermiFile reads a FITS file that may or may not be 4FGL's.
func readFermiFile(test *testing.T, file []byte) ([]nilgai.Source,
	[]error) {
	test.Helper()
	return collect(fermiTable(test), bytes.NewReader(file))
}

// withoutColumn returns the fixture's columns minus the one called name.
func withoutColumn(read fermiFixture, name string) []columnSpec {
	kept := make([]columnSpec, none)
	for _, each := range read.Columns {
		if each.Name != name {
			kept = append(kept, each)
		}
	}
	return kept
}

// oneFermi reads the row and returns the one source it holds. It fails the
// test if the row holds anything else.
func oneFermi(test *testing.T, read fermiFixture,
	row fermiRowSpec) nilgai.Source {
	test.Helper()
	sources, errs := readFermi(test, read, row)
	if len(sources) != one || len(errs) != none {
		test.Fatalf("%d sources and %v from %+v", len(sources), errs, row)
	}
	return sources[none]
}

// TestFermiReadsWhoAndWhere checks that a row of the point source table
// is a source of the 4FGL catalogue, at its position, under the name that
// 4FGL gives it.
func TestFermiReadsWhoAndWhere(test *testing.T) {
	read := loadFermi(test)
	for _, each := range read.Sources {
		source := oneFermi(test, read, each.Row)
		ra := Num(source.Position().RA())
		dec := Num(source.Position().Dec())
		if source.Catalogue() != nilgai.Fermi || source.ID() != each.ID ||
			!near(ra, each.Row.RA, read.Rel) ||
			!near(dec, each.Row.Dec, read.Rel) {
			test.Errorf("%s: read as %+v", each.Name, source)
		}
	}
}

// TestFermiReadsTheEllipse checks that the 95% ellipse in degrees is kept
// in arcseconds as stated, and at one sigma, with the angle east of north.
// testdata/fermi.json works these out by hand.
func TestFermiReadsTheEllipse(test *testing.T) {
	read := loadFermi(test)
	for _, each := range read.Sources {
		ellipse := oneFermi(test, read, each.Row).Error()
		stated := nilgai.Confidence(read.Stated)
		statedMajor := Num(ellipse.StatedSemiMajor())
		statedMinor := Num(ellipse.StatedSemiMinor())
		sigmaMajor, sigmaMinor := Num(ellipse.SemiMajor()),
			Num(ellipse.SemiMinor())
		apart := turnsApart(ellipse.PositionAngle(), float64(each.Angle))
		if ellipse.StatedAt() != stated ||
			!near(statedMajor, each.Major, read.Rel) ||
			!near(statedMinor, each.Minor, read.Rel) ||
			!near(sigmaMajor, each.SigmaMajor, read.Rel) ||
			!near(sigmaMinor, each.SigmaMinor, read.Rel) ||
			apart >= float64(read.Rel) {
			test.Errorf("%s: read as %+v", each.Name, ellipse)
		}
	}
}

// TestFermiReadsTheFluxAndKeepsTheFlags checks that the energy flux is a
// detection in erg/s/cm^2 through the Energy_Flux100 band, and that its
// flags, class and association are kept, in that order.
func TestFermiReadsTheFluxAndKeepsTheFlags(test *testing.T) {
	read := loadFermi(test)
	first, last, err := band.FilterBands("Energy_Flux100")
	if err != nil {
		test.Fatal(err)
	}
	for _, each := range read.Sources {
		flux := oneFermi(test, read, each.Row).Measurements()[none]
		gotFirst, gotLast := flux.Bands()
		value, valueIsEnergy := ergs(flux.Value())
		sigma, sigmaIsEnergy := ergs(flux.Error())
		if flux.Filter() != read.Filter || flux.Kind() != nilgai.Detection ||
			!valueIsEnergy || !sigmaIsEnergy ||
			!near(value, each.Row.Flux, read.Rel) ||
			!near(sigma, each.Row.FluxError, read.Rel) ||
			gotFirst != first || gotLast != last ||
			flux.Quality().String() != each.Quality {
			test.Errorf("%s: read as %+v", each.Name, flux)
		}
		names := []string{"Flags", "CLASS1", "ASSOC1"}
		flags := flux.Flags()
		for place, name := range names {
			if len(flags) != len(names) || flags[place].Name() != name ||
				flags[place].Value() != each.Kept[place] {
				test.Errorf("%s: flags %v, want %v", each.Name, flags,
					each.Kept)
			}
		}
	}
}

// TestFermiReadsFlagsAndSigmaApart is an anti-test. A source is not built
// from the stated 95% axes, and not given another source's flags. One
// sigma is under the stated axis, and a flagged source is doubtful where a
// clean one is good.
func TestFermiReadsFlagsAndSigmaApart(test *testing.T) {
	read := loadFermi(test)
	seen := map[string]bool{}
	for _, each := range read.Sources {
		sources, _ := readFermi(test, read, each.Row)
		ellipse := sources[none].Error()
		if ellipse.SemiMajor() >= ellipse.StatedSemiMajor() {
			test.Errorf("%s: one sigma %v is not under the stated %v",
				each.Name, ellipse.SemiMajor(), ellipse.StatedSemiMajor())
		}
		seen[sources[none].Measurements()[none].Quality().String()] = true
	}
	if len(seen) != halves {
		test.Errorf("qualities seen: %v, want good and doubtful", seen)
	}
}

// TestFermiRefusesRowsWithNoEllipse checks that a row with no ellipse is a
// RowError naming the ellipse's column and saying which kind of source has
// none, and that the row after it is read.
func TestFermiRefusesRowsWithNoEllipse(test *testing.T) {
	read := loadFermi(test)
	for _, each := range read.Refused {
		sources, errs := readFermi(test, read, each.Row, read.Good)
		var row *nilgai.RowError
		if len(errs) != one || !errors.As(errs[none], &row) ||
			row.Column() != each.Column || row.Row() != one ||
			len(sources) != one || sources[none].ID() != read.Good.Name {
			test.Errorf("%s: %d sources and %v, want column %q then a "+
				"good row", each.Name, len(sources), errs, each.Column)
			continue
		}
		for _, word := range each.Want {
			if !strings.Contains(errs[none].Error(), word) {
				test.Errorf("%s: %q does not say %q", each.Name,
					errs[none], word)
			}
		}
	}
}

// TestFermiSaysTheRightKindOfSource is an anti-test. An error does not
// blame the wrong kind of source, and does not invent an ellipse for the
// row.
func TestFermiSaysTheRightKindOfSource(test *testing.T) {
	read := loadFermi(test)
	for _, each := range read.Refused {
		sources, errs := readFermi(test, read, each.Row)
		if len(sources) != none || len(errs) != one {
			test.Errorf("%s: %d sources and %v, want none and one error",
				each.Name, len(sources), errs)
			continue
		}
		for _, word := range each.NotWant {
			if strings.Contains(errs[none].Error(), word) {
				test.Errorf("%s: %q says %q", each.Name, errs[none], word)
			}
		}
	}
}

// TestFermiRefusesFiles checks that input that is not a 4FGL file is
// yielded once, as an error that is not a row's. Three cases:
//   - not FITS at all
//   - FITS without the point source table
//   - the table without a column nilgai reads
func TestFermiRefusesFiles(test *testing.T) {
	read := loadFermi(test)
	files := map[string][]byte{
		"not fits": []byte("not a fits file"),
		"no table": fermiFile(test, "OTHER", read.Columns, read.Good),
		"no Flags": fermiFile(test, read.Table,
			withoutColumn(read, "Flags"), read.Good),
	}
	for name, file := range files {
		sources, errs := readFermiFile(test, file)
		var row *nilgai.RowError
		if len(errs) != one || errors.As(errs[none], &row) ||
			len(sources) != none {
			test.Errorf("%s: %d sources and %v, want one error that is "+
				"not a row's", name, len(sources), errs)
		}
	}
}

// TestFermiNamesWhatTheFileLacks is an anti-test. A column left out is
// named, and so is a table left out, so the caller is told what to give.
func TestFermiNamesWhatTheFileLacks(test *testing.T) {
	read := loadFermi(test)
	flagless := fermiFile(test, read.Table, withoutColumn(read, "Flags"),
		read.Good)
	_, errs := readFermiFile(test, flagless)
	if len(errs) != one || !strings.Contains(errs[none].Error(), "Flags") {
		test.Errorf("a file without Flags: %v", errs)
	}
	tableless := fermiFile(test, "OTHER", read.Columns, read.Good)
	_, errs = readFermiFile(test, tableless)
	if len(errs) != one || !strings.Contains(errs[none].Error(), read.Table) {
		test.Errorf("a file without the table: %v", errs)
	}
	if sources, errs := readFermi(test, read, read.Good); len(sources) != one ||
		len(errs) != none {
		test.Errorf("the whole file: %d sources and %v", len(sources), errs)
	}
}

// TestFermiReadsATableWithNoRows checks that a table with no rows yields
// nothing, and no error.
func TestFermiReadsATableWithNoRows(test *testing.T) {
	read := loadFermi(test)
	sources, errs := readFermi(test, read)
	if len(sources) != none || len(errs) != none {
		test.Errorf("%d sources and %v from a table of no rows",
			len(sources), errs)
	}
}

// TestFermiReadsEveryRowInOrder is an anti-test. A table of two rows
// yields two sources, not one and not none, in the order the file has
// them.
func TestFermiReadsEveryRowInOrder(test *testing.T) {
	read := loadFermi(test)
	first, second := read.Sources[none].Row, read.Sources[one].Row
	sources, errs := readFermi(test, read, first, second)
	if len(sources) != halves || len(errs) != none ||
		sources[none].ID() != strings.TrimSpace(first.Name) ||
		sources[one].ID() != strings.TrimSpace(second.Name) {
		test.Errorf("%d sources and %v from two rows", len(sources), errs)
	}
}

// TestFermiStopsWhenAsked checks that the reader does not yield again to
// a caller that stops early, whether at a source or at a row it cannot
// read.
func TestFermiStopsWhenAsked(test *testing.T) {
	read := loadFermi(test)
	for _, rows := range [][]fermiRowSpec{
		{read.Good, read.Good},
		{read.Refused[none].Row, read.Good},
	} {
		file := fermiFile(test, read.Table, read.Columns, rows...)
		seen := none
		for range fermiTable(test).Read("t", bytes.NewReader(file)) {
			seen++
			break
		}
		if seen != one {
			test.Errorf("yielded %d times after a stop", seen)
		}
	}
}

// TestFermiNeedsItsFilter checks that the reader's band comes from
// data/filters.json. Filters it cannot read give no reader, and the
// shipped filters give one.
func TestFermiNeedsItsFilter(test *testing.T) {
	for _, filters := range []string{"{", `{"filters": []}`} {
		_, err := read.ReadFermi([]byte(filters))
		if err == nil {
			test.Errorf("a reader was made with filters %q", filters)
		}
	}
	if _, err := read.ReadFermi([]byte(goodFilters)); err != nil {
		test.Error(err)
	}
	if _, err := read.NewFermi(); err != nil {
		test.Error(err)
	}
}

package read_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/janearc/nilgai"
)

// realFermi describes the real rows of 4FGL in testdata/4fgl-rows.fit,
// and what the reader must make of them.
type realFermi struct {
	File   string
	Rel    Num
	Centre struct {
		ID, Quality                     string
		RA, Dec                         Num
		StatedMajor, StatedMinor, Angle Num
		Flux, FluxError                 Num
		Kept                            []string
	}
	Unassociated struct {
		Limits        []float64
		GLATTolerance Num
		Fake          struct{ RA, Dec, Axis, Flux Num }
		Rows          []struct {
			ID, Class  string
			GLAT       Num
			SelectedAt []float64
		}
	}
	Refused []struct {
		Row    int
		Column string
		Want   []string
	}
}

// loadRealFermi returns the fixture's account of the real rows, and the
// rows themselves.
func loadRealFermi(test *testing.T) (realFermi, []byte) {
	test.Helper()
	var account struct{ Real realFermi }
	load(test, "fermi", &account)
	file, err := fixtures.ReadFile("testdata/" + account.Real.File)
	if err != nil {
		test.Fatal(err)
	}
	return account.Real, file
}

// Real 4FGL rows must retain file order. The Galactic Centre source keeps its
// position, ellipse, flux, and flags with Doubtful quality; rows without
// ellipses are reported as row errors.
func TestFermiReadsTheRealRows(test *testing.T) {
	account, file := loadRealFermi(test)
	sources, errs := readFermiFile(test, file)
	if len(sources) != one+len(account.Unassociated.Rows) ||
		len(errs) != len(account.Refused) {
		test.Fatalf("%d sources and %v from the real rows",
			len(sources), errs)
	}
	got, want := sources[none], account.Centre
	flux := got.Measurements()[none]
	ellipse := got.Error()
	position := got.Position()
	rel := account.Rel
	ra, dec := Num(position.RA()), Num(position.Dec())
	major, minor := Num(ellipse.StatedSemiMajor()), Num(ellipse.StatedSemiMinor())
	angle := Num(ellipse.PositionAngle())
	value, sigma := fluxNumber(flux.Value()), fluxNumber(flux.Error())
	if got.ID() != want.ID || got.Catalogue() != nilgai.Fermi ||
		!near(ra, want.RA, rel) || !near(dec, want.Dec, rel) ||
		!near(major, want.StatedMajor, rel) ||
		!near(minor, want.StatedMinor, rel) ||
		!near(angle, want.Angle, rel) || !near(value, want.Flux, rel) ||
		!near(sigma, want.FluxError, rel) ||
		flux.Quality().String() != want.Quality {
		test.Errorf("read as %+v", got)
	}
	var kept []string
	for _, each := range flux.Flags() {
		kept = append(kept, each.Value())
	}
	if strings.Join(kept, "|") != strings.Join(want.Kept, "|") {
		test.Errorf("flags kept as %q, want %q", kept, want.Kept)
	}
}

// Missing ellipses must produce row errors without ending the sequence. Each
// error must identify the correct reason, distinguishing extended sources from
// externally fixed positions.
func TestFermiRefusesTheRealRowsWithNoEllipse(test *testing.T) {
	account, file := loadRealFermi(test)
	_, errs := readFermiFile(test, file)
	for place, want := range account.Refused {
		var row *nilgai.RowError
		if place >= len(errs) || !errors.As(errs[place], &row) ||
			row.Row() != want.Row || row.Column() != want.Column {
			test.Errorf("refusal %d: %v, want row %d, column %s", place,
				errs, want.Row, want.Column)
			continue
		}
		for _, word := range want.Want {
			if !strings.Contains(errs[place].Error(), word) {
				test.Errorf("row %d: %q does not say %q", want.Row,
					errs[place], word)
			}
		}
	}
	other := errs[none].Error()
	if strings.Contains(other, "fixed from outside") {
		test.Errorf("the extended source is called a fixed one: %q", other)
	}
}

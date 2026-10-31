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

// TestFermiReadsTheRealRows checks the real rows in the order the file
// has them:
//   - the Galactic Centre source is a source of the 4FGL catalogue, at its
//     real position, with its real ellipse and flux, and with its flags
//     kept and read as doubtful
//   - the other two are named as rows that were lost
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
	value, valueIsEnergy := ergs(flux.Value())
	sigma, sigmaIsEnergy := ergs(flux.Error())
	if got.ID() != want.ID || got.Catalogue() != nilgai.Fermi ||
		!valueIsEnergy || !sigmaIsEnergy ||
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

// TestFermiRefusesTheRealRowsWithNoEllipse is an anti-test. The two real
// rows without an ellipse are refused as rows, and not as the end of the
// file. Each is refused for its own reason, and the reason is the right
// one: an extended source is not called a nova.
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

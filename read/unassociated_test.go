package read_test

import (
	"bytes"
	"iter"
	"math"
	"slices"
	"testing"

	"github.com/janearc/nilgai"
	"github.com/janearc/nilgai/band"
	"github.com/janearc/nilgai/read"
)

// selectedIDs returns the identifiers that Unassociated yields over the
// real rows at the limit, and the errors it passes on.
func selectedIDs(test *testing.T, file []byte, limit float64) (ids []string,
	errs []error) {
	test.Helper()
	rows := fermiTable(test).Read("real", bytes.NewReader(file))
	for source, err := range read.Unassociated(rows, limit) {
		if err != nil {
			errs = append(errs, err)
			continue
		}
		ids = append(ids, source.ID())
	}
	return ids, errs
}

// TestUnassociatedOverTheRealRows checks that, over the real rows, the
// unassociated source near the plane is selected. Three anti-cases are
// not:
//   - the associated pulsar at low latitude
//   - the source of unknown nature at low latitude, because class unk is
//     an association
//   - the unassociated source far from the plane
//
// Each row's latitude, as the reader's position gives it, is the file's
// own GLAT.
func TestUnassociatedOverTheRealRows(test *testing.T) {
	account, file := loadRealFermi(test)
	sources, _ := readFermiFile(test, file)
	rows := account.Unassociated.Rows
	for place, want := range rows {
		got := sources[one+place]
		class := got.Measurements()[none].Flags()[one].Value()
		latitude := got.Position().GalacticLatitude()
		if got.ID() != want.ID || class != want.Class ||
			math.Abs(latitude-float64(want.GLAT)) >
				float64(account.Unassociated.GLATTolerance) {
			test.Errorf("row %d: %s, class %q, b %v; want %s, %q, %v",
				place, got.ID(), class, latitude, want.ID, want.Class,
				want.GLAT)
		}
	}
	for _, limit := range account.Unassociated.Limits {
		var want []string
		for _, each := range rows {
			if slices.Contains(each.SelectedAt, limit) {
				want = append(want, each.ID)
			}
		}
		got, _ := selectedIDs(test, file, limit)
		if !slices.Equal(got, want) {
			test.Errorf("limit %v selects %q, want %q", limit, got, want)
		}
	}
}

// TestUnassociatedKeepsTheErrors checks that rows a reader refuses come
// through Unassociated as the errors they are, in their place, and are
// not dropped.
func TestUnassociatedKeepsTheErrors(test *testing.T) {
	account, file := loadRealFermi(test)
	limit := account.Unassociated.Limits[none]
	_, errs := selectedIDs(test, file, limit)
	if len(errs) != len(account.Refused) {
		test.Errorf("%d errors passed on, want %d: %v", len(errs),
			len(account.Refused), errs)
	}
}

// handed holds sources given as a read of them, and counts what it
// yields.
type handed struct {
	sources []nilgai.Source
	yielded int
}

// each yields the sources until the sources end or the caller stops.
func (given *handed) each(yield func(*nilgai.Source, error) bool) {
	for _, source := range given.sources {
		given.yielded++
		if !yield(&source, nil) {
			return
		}
	}
}

// TestUnassociatedStopsWhenTold checks that when a caller stops reading,
// Unassociated stops too. It asks for nothing after the yield that says
// no, whether that yield was an error or a source.
func TestUnassociatedStopsWhenTold(test *testing.T) {
	account, file := loadRealFermi(test)
	limit := account.Unassociated.Limits[one]
	rows := fermiTable(test).Read("real", bytes.NewReader(file))
	seen := none
	for range read.Unassociated(rows, limit) {
		seen++
		break
	}
	if seen != one {
		test.Errorf("%d yields after the caller broke out", seen)
	}
	given := &handed{sources: []nilgai.Source{
		blankFermi(test, account, nilgai.Fermi, true),
		blankFermi(test, account, nilgai.Fermi, true)}}
	for range read.Unassociated(given.each, limit) {
		break
	}
	if given.yielded != one {
		test.Errorf("%d sources asked for after the caller broke out",
			given.yielded)
	}
}

// blankFermi returns a source at the Galactic Centre, of the catalogue
// given. It has a blank CLASS1 flag when classed, and no flags at all when
// not.
func blankFermi(test *testing.T, account realFermi,
	catalogue nilgai.Catalogue, classed bool) nilgai.Source {
	test.Helper()
	fake := account.Unassociated.Fake
	first, last, err := band.FilterBands("Energy_Flux100")
	if err != nil {
		test.Fatal(err)
	}
	energy := nilgai.ErgPerSecPerCm2(float64(fake.Flux))
	flux, err := nilgai.NewDetection("f", first, last, energy, energy)
	if err != nil {
		test.Fatal(err)
	}
	if classed {
		flag, err := nilgai.NewFlag("CLASS1", "")
		if err != nil {
			test.Fatal(err)
		}
		flux, err = flux.WithReading(nilgai.Good, []nilgai.Flag{flag})
		if err != nil {
			test.Fatal(err)
		}
	}
	position, err := nilgai.NewPosition(float64(fake.RA), float64(fake.Dec))
	if err != nil {
		test.Fatal(err)
	}
	ellipse, err := nilgai.NewErrorEllipse(float64(fake.Axis),
		float64(fake.Axis), none, nilgai.OneSigma())
	if err != nil {
		test.Fatal(err)
	}
	var epoch nilgai.When
	source, err := nilgai.NewSource(catalogue, "x", position, ellipse,
		epoch, []nilgai.Measurement{*flux})
	if err != nil {
		test.Fatal(err)
	}
	return *source
}

// TestUnassociatedOnlyForFermiWithAClass checks that Unassociated selects
// only what it can see to be unassociated in 4FGL. A blank CLASS1 flag on
// a source of another catalogue is not 4FGL's blank, and a Fermi source
// with no CLASS1 flag at all has not said it is unassociated.
func TestUnassociatedOnlyForFermiWithAClass(test *testing.T) {
	account, _ := loadRealFermi(test)
	limit := account.Unassociated.Limits[one]
	for _, each := range []struct {
		name      string
		catalogue nilgai.Catalogue
		classed   bool
		want      int
	}{{"a Fermi source with a blank class", nilgai.Fermi, true, one},
		{"an AllWISE source with a blank class", nilgai.AllWISE, true, none},
		{"a Fermi source with no class", nilgai.Fermi, false, none}} {
		given := &handed{sources: []nilgai.Source{
			blankFermi(test, account, each.catalogue, each.classed)}}
		var only iter.Seq2[*nilgai.Source, error] = given.each
		got := none
		for range read.Unassociated(only, limit) {
			got++
		}
		if got != each.want {
			test.Errorf("%s: %d selected, want %d", each.name, got,
				each.want)
		}
	}
}

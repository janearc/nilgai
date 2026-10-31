package band_test

import (
	"cmp"
	"encoding/json"
	"math"
	"slices"
	"testing"

	"github.com/janearc/libtheme-css/primitives/swatch"
	"github.com/janearc/libtheme-css/spaces/radiation"
	"github.com/janearc/nilgai/band"
)

// profileFixture is testdata/profiles.json.
type profileFixture struct {
	SameTo Num
	Names  []string
	Fermi  struct{ Lo, Hi Num }
	Units  []struct {
		Value Num
		Unit  string
		NM    Num `json:"nm"`
	}
	UnitsRefused []struct {
		Value Num
		Unit  string
	}
	Bad []struct {
		Name string
		Doc  json.RawMessage
	}
}

// loadProfiles reads the profiles fixture.
func loadProfiles(test *testing.T) profileFixture {
	test.Helper()
	var read profileFixture
	load(test, "profiles", &read)
	return read
}

// Every unit nilgai reads comes to the wavelength the fixture works out.
func TestNanometres(test *testing.T) {
	read := loadProfiles(test)
	for _, each := range read.Units {
		got, err := band.Quantity{Value: float64(each.Value),
			Unit: each.Unit}.Nanometres()
		if err != nil || !near(Num(got), each.NM, read.SameTo) {
			test.Errorf("%v %s: %v nm, %v; want %v", each.Value,
				each.Unit, got, err, each.NM)
		}
	}
}

// TestNanometresRefuses is the anti-test. Nanometres refuses:
//   - a unit nilgai does not know
//   - a value that is not above zero
//   - a value that comes to no wavelength
func TestNanometresRefuses(test *testing.T) {
	for _, each := range loadProfiles(test).UnitsRefused {
		got, err := band.Quantity{Value: float64(each.Value),
			Unit: each.Unit}.Nanometres()
		if err == nil || got != none {
			test.Errorf("%v %q was read as %v nm", each.Value, each.Unit,
				got)
		}
	}
}

// TestProfilesReadTheirData reads the embedded profiles, one per
// instrument in the file's order, each band where its ends say: Fermi's
// short end is 1 TeV and its long end 50 MeV, as the fixture works out.
func TestProfilesReadTheirData(test *testing.T) {
	read := loadProfiles(test)
	got, err := band.Profiles()
	if err != nil {
		test.Fatal(err)
	}
	var names []string
	for _, each := range got {
		names = append(names, each.Name)
	}
	if !slices.Equal(names, read.Names) {
		test.Fatalf("profiles %v, want %v", names, read.Names)
	}
	fermi := got[none].Source
	if !near(Num(fermi.Lo), read.Fermi.Lo, read.SameTo) ||
		!near(Num(fermi.Hi), read.Fermi.Hi, read.SameTo) {
		test.Errorf("fermi runs %v to %v nm", fermi.Lo, fermi.Hi)
	}
}

// Every instrument shows a colour across its band: the middle of each
// band, taken evenly in the logarithm as libtheme spreads its ranges, is
// light and not dark.
func TestEachProfileShowsAColour(test *testing.T) {
	got, err := band.Profiles()
	if err != nil {
		test.Fatal(err)
	}
	for _, each := range got {
		middle := math.Sqrt(each.Source.Lo * each.Source.Hi)
		if each.Swatch(middle) == swatch.NoLight() {
			test.Errorf("%s shows nothing at %v nm", each.Name, middle)
		}
	}
}

// byShortestEnd orders profiles by the short end of their bands.
func byShortestEnd(first, second radiation.Profile) int {
	return cmp.Compare(first.Source.Lo, second.Source.Lo)
}

// TestProfilesKeepTheSpectrumsOrder: of any two instruments, the one at
// shorter wavelengths is shown bluer, so a shuffle in the file shows here.
func TestProfilesKeepTheSpectrumsOrder(test *testing.T) {
	got, err := band.Profiles()
	if err != nil {
		test.Fatal(err)
	}
	byBand := slices.Clone(got)
	slices.SortFunc(byBand, byShortestEnd)
	for place, bluer := range byBand[:len(byBand)-one] {
		redder := byBand[place+one]
		if bluer.Render.Hi > redder.Render.Lo {
			test.Errorf("%s is shown redder than %s", bluer.Name,
				redder.Name)
		}
	}
}

// Each document in the fixture's bad list is refused, and gives nothing.
func TestReadProfilesRefuses(test *testing.T) {
	for _, each := range loadProfiles(test).Bad {
		got, err := band.ReadProfiles(each.Doc)
		if err == nil || got != nil {
			test.Errorf("%s was accepted: %v", each.Name, got)
		}
	}
}

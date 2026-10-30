package nilgai_test

import (
	"embed"
	"encoding/json"
	"fmt"
	"math"
	"testing"

	"github.com/janearc/libtheme-css/primitives/bands"
	"github.com/janearc/nilgai"
)

//go:embed testdata/*.json
var fixtures embed.FS

// none is defined as a count of nothing.
const none = 0

// positive is defined: the sign math.Inf is asked for to get +Inf.
const positive = +1

// Num is a number in a fixture. JSON has no NaN or infinity, so a
// fixture writes them as the strings "NaN", "Inf" and "-Inf".
type Num float64

// UnmarshalJSON reads a number, or one of the three special strings.
func (num *Num) UnmarshalJSON(data []byte) error {
	var text string
	if json.Unmarshal(data, &text) == nil {
		return num.special(text)
	}
	var plain float64
	if err := json.Unmarshal(data, &plain); err != nil {
		return err
	}
	*num = Num(plain)
	return nil
}

// special reads NaN, Inf, -Inf or the word OneSigma.
func (num *Num) special(text string) error {
	switch text {
	case "NaN":
		*num = Num(math.NaN())
	case "Inf":
		*num = Num(math.Inf(positive))
	case "-Inf":
		*num = Num(math.Inf(-positive))
	case "OneSigma":
		*num = Num(nilgai.OneSigma())
	default:
		return fmt.Errorf("fixture number %q is not NaN, Inf, -Inf or "+
			"OneSigma", text)
	}
	return nil
}

// load reads testdata/name.json into into.
func load(test *testing.T, name string, into any) {
	test.Helper()
	data, err := fixtures.ReadFile("testdata/" + name + ".json")
	if err != nil {
		test.Fatal(err)
	}
	if err := json.Unmarshal(data, into); err != nil {
		test.Fatalf("%s: %v", name, err)
	}
}

// near says whether got is within rel of want, as a fraction of want;
// a want of nothing must be matched exactly.
func near(got, want, rel Num) bool {
	return math.Abs(float64(got-want)) <= float64(rel)*math.Abs(float64(want))
}

// bandSpec is a filter as a fixture writes it: its name and edges in nm.
type bandSpec struct {
	Name   string
	Lo, Hi Num
}

// filter is a catalogue's filter as nilgai holds it: a name and the
// first and last libtheme bands it covers.
type filter struct {
	name        string
	first, last bands.Band
}

// build is the filter from its edges, failing the test if refused.
func (spec bandSpec) build(test *testing.T) filter {
	test.Helper()
	first, last, err := nilgai.BandsBetween(float64(spec.Lo),
		float64(spec.Hi))
	if err != nil {
		test.Fatal(err)
	}
	return filter{name: spec.Name, first: first, last: last}
}

// fluxSpec is a flux as a fixture writes it.
type fluxSpec struct {
	Name  string
	Unit  string
	Value Num
}

// flux is the flux in the unit the fixture names: Jy, erg, or nil.
func (spec fluxSpec) flux() nilgai.Flux {
	switch spec.Unit {
	case "Jy":
		return nilgai.FluxDensity(spec.Value)
	case "erg":
		return nilgai.EnergyFlux(spec.Value)
	}
	return nil
}

// world is testdata/world.json.
type world struct {
	W1, W2      bandSpec
	Position    struct{ RA, Dec Num }
	BadPosition struct{ RA, Dec Num }
	Ellipse     struct{ Major, Minor, Angle, Level Num }
	Density     struct{ Value, Error Num }
	Energy      struct{ Value, Error Num }
	Looks       []struct {
		MJD  Num
		Band string
	}
	Span struct {
		First, Last Num
		W2Looks     int
	}
	RowError struct {
		File, Column, Text string
		Row                int
	}
	Filters []struct {
		Name        string
		Lo, Hi      Num
		First, Last int
		OnEdge      string
	}
	BadBands       []bandSpec
	WhenOk         []Num
	WhenRefused    []Num
	InvalidNumbers []int
	BadFluxes      []fluxSpec
}

// loadWorld reads the world fixture.
func loadWorld(test *testing.T) world {
	test.Helper()
	var read world
	load(test, "world", &read)
	return read
}

// position builds the fixture's position.
func (read world) position(test *testing.T) nilgai.Position {
	test.Helper()
	position, err := nilgai.NewPosition(float64(read.Position.RA),
		float64(read.Position.Dec))
	if err != nil {
		test.Fatal(err)
	}
	return position
}

// ellipse builds the fixture's error ellipse.
func (read world) ellipse(test *testing.T) nilgai.ErrorEllipse {
	test.Helper()
	spec := read.Ellipse
	ellipse, err := nilgai.NewErrorEllipse(float64(spec.Major),
		float64(spec.Minor), float64(spec.Angle),
		nilgai.Confidence(spec.Level))
	if err != nil {
		test.Fatal(err)
	}
	return ellipse
}

// density builds a flux-density detection through the filter.
func (read world) density(
	test *testing.T,
	through filter,
) nilgai.Measurement {
	test.Helper()
	measurement, err := nilgai.NewDetection(through.name, through.first,
		through.last, nilgai.FluxDensity(read.Density.Value),
		nilgai.FluxDensity(read.Density.Error))
	if err != nil {
		test.Fatal(err)
	}
	return measurement
}

// look builds a flux-density detection through the filter at the MJD
// given.
func (read world) look(test *testing.T, through filter,
	mjd Num) nilgai.Measurement {
	test.Helper()
	when, err := nilgai.NewWhen(float64(mjd))
	if err != nil {
		test.Fatal(err)
	}
	return read.density(test, through).WithTime(when)
}

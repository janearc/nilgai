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

// none is the count of nothing, zero.
const none = 0

// positive is the sign that math.Inf takes to return +Inf.
const positive = +1

// unit is one of a unit, the amount a conversion is defined by.
const unit = 1

// Num is a number in a fixture. JSON has no NaN or infinity, so a fixture
// writes them as the strings "NaN", "Inf" and "-Inf".
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

// special sets num from the text "NaN", "Inf", "-Inf" or "OneSigma".
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

// near says whether got is within rel of want, where rel is a fraction of
// want. A want of zero must be matched exactly.
func near(got, want, rel Num) bool {
	return math.Abs(float64(got-want)) <= float64(rel)*math.Abs(float64(want))
}

// bandSpec describes a filter as a fixture writes it: its name and edges
// in nm.
type bandSpec struct {
	Name   string
	Lo, Hi Num
}

// filter holds a catalogue's filter as nilgai holds it: a name and the
// first and last libtheme band it covers.
type filter struct {
	name        string
	first, last bands.Band
}

// build returns the filter for its edges, which is the libtheme bands the
// edges fall in. The fixture's edges are not on a band's edge, so the band
// that the high edge falls in is the filter's last.
func (spec bandSpec) build() filter {
	return filter{name: spec.Name, first: bands.Of(float64(spec.Lo)),
		last: bands.Of(float64(spec.Hi))}
}

// fluxSpec describes a flux as a fixture writes it.
type fluxSpec struct {
	Name  string
	Unit  string
	Value Num
}

// refusals tries the flux where NewDetection and NewUpperLimit take one,
// beside a good flux of its kind from the world, and says which refused
// it. The fixture names the kind as "Jy" or "erg".
func (spec fluxSpec) refusals(test *testing.T, read world,
	through filter) refused {
	test.Helper()
	bad := float64(spec.Value)
	switch spec.Unit {
	case "Jy":
		good := float64(read.Density.Value)
		return refusalsOf(through, nilgai.Jansky(bad), nilgai.Jansky(good))
	case "erg":
		good := float64(read.Energy.Value)
		return refusalsOf(through, nilgai.ErgPerSecPerCm2(bad),
			nilgai.ErgPerSecPerCm2(good))
	}
	test.Fatalf("%s: unit %q is not Jy or erg", spec.Name, spec.Unit)
	return refused{}
}

// refused says where a bad flux was refused: as a detection's value, as
// its error, and as an upper limit.
type refused struct{ value, err, limit bool }

// eitherFlux is a FluxDensity or an EnergyFlux as a type parameter, as
// NewDetection and NewUpperLimit take one.
type eitherFlux interface {
	nilgai.FluxDensity | nilgai.EnergyFlux
	nilgai.Flux
}

// refusalsOf tries bad as a detection's value and error beside good, and
// as an upper limit. A refusal counts only when it gives no measurement.
func refusalsOf[Of eitherFlux](through filter, bad, good Of) refused {
	value, asValue := nilgai.NewDetection(through.name, through.first,
		through.last, bad, good)
	sigma, asError := nilgai.NewDetection(through.name, through.first,
		through.last, good, bad)
	limit, asLimit := nilgai.NewUpperLimit(through.name, through.first,
		through.last, bad)
	return refused{value == nil && asValue != nil,
		sigma == nil && asError != nil, limit == nil && asLimit != nil}
}

// world holds testdata/world.json.
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
	WhenOk         []Num
	WhenRefused    []Num
	InvalidNumbers []int
	BadFluxes      []fluxSpec
	InSI           struct{ Jansky, ErgPerSecPerCm2, Rel Num }
}

// loadWorld reads the world fixture.
func loadWorld(test *testing.T) world {
	test.Helper()
	var read world
	load(test, "world", &read)
	return read
}

// position returns the fixture's position.
func (read world) position(test *testing.T) nilgai.Position {
	test.Helper()
	position, err := nilgai.NewPosition(float64(read.Position.RA),
		float64(read.Position.Dec))
	if err != nil {
		test.Fatal(err)
	}
	return position
}

// ellipse returns the fixture's error ellipse.
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

// density returns a flux-density detection through the filter.
func (read world) density(
	test *testing.T,
	through filter,
) nilgai.Measurement {
	test.Helper()
	value, sigma := float64(read.Density.Value), float64(read.Density.Error)
	measurement, err := nilgai.NewDetection(through.name, through.first,
		through.last, nilgai.Jansky(value), nilgai.Jansky(sigma))
	if err != nil {
		test.Fatal(err)
	}
	return *measurement
}

// look returns a flux-density detection through the filter at the MJD
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

package read_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/janearc/nilgai"
	"github.com/janearc/nilgai/read"
)

type wiseCase struct {
	Name          string
	Band          string
	Mag, Err      Num
	Flux, FluxErr Num
	Rel           Num
	Data          string
}

type wiseFixture struct {
	Flux, FluxRefused []wiseCase
	DifferenceStep    struct{ Mag, Err, Rel Num }
	BrokenZeroPoints  []wiseCase
	GoodZeroPoints    wiseCase
}

func loadWise(test *testing.T) wiseFixture {
	test.Helper()
	var fixture wiseFixture
	load(test, "wise", &fixture)
	return fixture
}

// bandNamed returns the WISE band that a fixture names. Any other name
// gives a band that does not exist.
func bandNamed(name string) read.WiseBand {
	switch name {
	case "W1":
		return read.W1
	case "W2":
		return read.W2
	case "W3":
		return read.W3
	case "W4":
		return read.W4
	}
	return read.WiseBand(-len(name))
}

// halved is the divisor that takes half of a number, two.
const halved = 2

func TestWiseFlux(test *testing.T) {
	for _, each := range loadWise(test).Flux {
		flux, fluxErr, err := read.WiseFlux(bandNamed(each.Band),
			float64(each.Mag), float64(each.Err))
		if err != nil || !near(inJanskys(flux), each.Flux, each.Rel) ||
			!near(inJanskys(fluxErr), each.FluxErr, each.Rel) {
			test.Errorf("%s: %v, %v, %v", each.Name, flux, fluxErr, err)
		}
	}
}

func TestWiseFluxRefuses(test *testing.T) {
	for _, each := range loadWise(test).FluxRefused {
		flux, fluxErr, err := read.WiseFlux(bandNamed(each.Band),
			float64(each.Mag), float64(each.Err))
		if err == nil || flux != none || fluxErr != none {
			test.Errorf("%s was accepted: %v %v", each.Name, flux, fluxErr)
		}
	}
}

// TestWiseFluxErrorMatchesTheDifference checks that, for a small error,
// the first-order error agrees with half the difference between the
// fluxes at the magnitude plus its error and minus its error.
func TestWiseFluxErrorMatchesTheDifference(test *testing.T) {
	step := loadWise(test).DifferenceStep
	_, propagated, err := read.WiseFlux(read.W1, float64(step.Mag),
		float64(step.Err))
	faint, _, faintErr := read.WiseFlux(read.W1,
		float64(step.Mag+step.Err), none)
	bright, _, brightErr := read.WiseFlux(read.W1,
		float64(step.Mag-step.Err), none)
	if err != nil || faintErr != nil || brightErr != nil {
		test.Fatal(err, faintErr, brightErr)
	}
	half := Num(bright-faint) / halved
	if !near(Num(propagated), half, step.Rel) {
		test.Errorf("error %v, half the difference %v", propagated, half)
	}
}

func TestZeroPoint(test *testing.T) {
	for _, each := range loadWise(test).Flux {
		if each.Mag != none {
			continue
		}
		got, err := bandNamed(each.Band).ZeroPoint()
		if err != nil || got != nilgai.Jansky(float64(each.Flux)) {
			test.Errorf("%s: %v, %v", each.Name, got, err)
		}
	}
}

func TestZeroPointRefusesOtherBands(test *testing.T) {
	for _, each := range loadWise(test).FluxRefused {
		if each.Band == "W1" {
			continue
		}
		if got, err := bandNamed(each.Band).ZeroPoint(); err == nil ||
			got != none {
			test.Errorf("%s: %v, %v", each.Name, got, err)
		}
	}
}

func TestReadWiseZeroPointsRefusesBrokenData(test *testing.T) {
	for _, each := range loadWise(test).BrokenZeroPoints {
		got, err := read.ReadWiseZeroPoints([]byte(each.Data))
		if err == nil || got != (read.WiseZeroPoints{}) {
			test.Errorf("%s was accepted: %v", each.Name, got)
			continue
		}
		if each.Band != "" && !strings.Contains(err.Error(), each.Band) {
			test.Errorf("%s: %q does not name %s", each.Name, err,
				each.Band)
		}
	}
}

func TestReadWiseZeroPointsReadsGoodData(test *testing.T) {
	good := loadWise(test).GoodZeroPoints
	points, err := read.ReadWiseZeroPoints([]byte(good.Data))
	if err != nil {
		test.Fatal(err)
	}
	var given map[string]Num
	if err := json.Unmarshal([]byte(good.Data), &given); err != nil {
		test.Fatal(err)
	}
	for name, want := range given {
		got, err := points.For(bandNamed(name))
		if err != nil || got != nilgai.Jansky(float64(want)) {
			test.Errorf("%s: %v, %v", name, got, err)
		}
	}
}

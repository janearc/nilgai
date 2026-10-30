package nilgai_test

import (
	"testing"

	"github.com/janearc/nilgai"
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

// bandNamed is the WISE band a fixture names; any other name is a band
// that does not exist.
func bandNamed(name string) nilgai.WiseBand {
	switch name {
	case "W1":
		return nilgai.W1
	case "W2":
		return nilgai.W2
	}
	return nilgai.WiseBand(len(name))
}

// halved is defined: half is a division by two.
const halved = 2

func TestWiseFlux(test *testing.T) {
	for _, each := range loadWise(test).Flux {
		flux, fluxErr, err := nilgai.WiseFlux(bandNamed(each.Band),
			float64(each.Mag), float64(each.Err))
		if err != nil || !near(Num(flux), each.Flux, each.Rel) ||
			!near(Num(fluxErr), each.FluxErr, each.Rel) {
			test.Errorf("%s: %v, %v, %v", each.Name, flux, fluxErr, err)
		}
	}
}

func TestWiseFluxRefuses(test *testing.T) {
	for _, each := range loadWise(test).FluxRefused {
		flux, fluxErr, err := nilgai.WiseFlux(bandNamed(each.Band),
			float64(each.Mag), float64(each.Err))
		if err == nil || flux != none || fluxErr != none {
			test.Errorf("%s was accepted: %v %v", each.Name, flux, fluxErr)
		}
	}
}

// The first-order error agrees with half the difference of the fluxes
// at the magnitude plus and minus its error, for a small error.
func TestWiseFluxErrorMatchesTheDifference(test *testing.T) {
	step := loadWise(test).DifferenceStep
	_, propagated, err := nilgai.WiseFlux(nilgai.W1, float64(step.Mag),
		float64(step.Err))
	faint, _, faintErr := nilgai.WiseFlux(nilgai.W1,
		float64(step.Mag+step.Err), none)
	bright, _, brightErr := nilgai.WiseFlux(nilgai.W1,
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
		if err != nil || Num(got) != each.Flux {
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
		if got, err := nilgai.ReadWiseZeroPoints([]byte(each.Data)); err == nil ||
			got != (nilgai.WiseZeroPoints{}) {
			test.Errorf("%s was accepted: %v", each.Name, got)
		}
	}
}

func TestReadWiseZeroPointsReadsGoodData(test *testing.T) {
	good := loadWise(test).GoodZeroPoints
	points, err := nilgai.ReadWiseZeroPoints([]byte(good.Data))
	if err != nil {
		test.Fatal(err)
	}
	got, err := points.For(nilgai.W2)
	if err != nil || Num(got) != good.Flux {
		test.Errorf("%v, %v", got, err)
	}
}

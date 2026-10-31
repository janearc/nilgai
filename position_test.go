package nilgai_test

import (
	"math"
	"testing"

	"github.com/janearc/nilgai"
)

type positionCase struct {
	Name    string
	RA, Dec Num
}

type positionFixture struct{ Ok, Refused []positionCase }

func TestNewPosition(test *testing.T) {
	var fixture positionFixture
	load(test, "position", &fixture)
	for _, each := range fixture.Ok {
		got, err := nilgai.NewPosition(float64(each.RA), float64(each.Dec))
		if err != nil || got.RA() != float64(each.RA) ||
			got.Dec() != float64(each.Dec) {
			test.Errorf("%s: %v, %v", each.Name, got, err)
		}
	}
}

// TestGalacticLatitude checks that the Galactic latitude of real 4FGL
// positions is the file's own GLAT, and that the latitude of the poles is
// a right angle.
func TestGalacticLatitude(test *testing.T) {
	var fixture struct {
		Galactic struct {
			Tol   Num
			Cases []struct {
				Name    string
				RA, Dec Num
				Lat     Num
			}
		}
	}
	load(test, "position", &fixture)
	for _, each := range fixture.Galactic.Cases {
		position, err := nilgai.NewPosition(float64(each.RA), float64(each.Dec))
		if err != nil {
			test.Fatal(err)
		}
		got := position.GalacticLatitude()
		if math.Abs(got-float64(each.Lat)) > float64(fixture.Galactic.Tol) {
			test.Errorf("%s: b is %v, want %v", each.Name, got, each.Lat)
		}
	}
}

func TestNewPositionRefuses(test *testing.T) {
	var fixture positionFixture
	load(test, "position", &fixture)
	for _, each := range fixture.Refused {
		got, err := nilgai.NewPosition(float64(each.RA), float64(each.Dec))
		if err == nil || got != (nilgai.Position{}) {
			test.Errorf("%s was accepted: %v", each.Name, got)
		}
	}
}

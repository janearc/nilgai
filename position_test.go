package nilgai_test

import (
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

package sky_test

import (
	"embed"
	"encoding/json"
	"fmt"
	"math"
	"testing"
)

//go:embed testdata/*.json
var fixtures embed.FS

// none is the count of nothing, zero.
const none = 0

// one is the count of a single thing.
const one = 1

// halves is the number of equal parts in a whole that a half is one of.
const halves = 2

// positive is the sign that math.Inf takes to return +Inf.
const positive = +1

// arcsecondsPerDegree is the arcseconds in a degree: sixty minutes of
// sixty seconds.
const arcsecondsPerDegree = 3600

// degreesInHalfACircle is the degrees in half a circle, which is pi
// radians.
const degreesInHalfACircle = 180

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

// special sets num from the text "NaN", "Inf" or "-Inf".
func (num *Num) special(text string) error {
	switch text {
	case "NaN":
		*num = Num(math.NaN())
	case "Inf":
		*num = Num(math.Inf(positive))
	case "-Inf":
		*num = Num(math.Inf(-positive))
	default:
		return fmt.Errorf("fixture number %q is not NaN, Inf or -Inf",
			text)
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

// turnsApart returns how far apart two angles in degrees are, measured
// the short way round the circle.
func turnsApart(got, want float64) float64 {
	turn := float64(halves * degreesInHalfACircle)
	apart := math.Mod(math.Abs(got-want), turn)
	return min(apart, turn-apart)
}

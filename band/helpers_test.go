package band_test

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

// positive is the sign that math.Inf takes to return +Inf.
const positive = +1

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

// near says whether got is within rel of want, where rel is a fraction of
// want. A want of zero must be matched exactly.
func near(got, want, rel Num) bool {
	return math.Abs(float64(got-want)) <= float64(rel)*math.Abs(float64(want))
}

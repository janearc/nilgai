package read_test

import (
	"embed"
	"encoding/json"
	"fmt"
	"math"
	"testing"

	"github.com/janearc/nilgai"
)

//go:embed testdata/*.json testdata/*.tsv testdata/*.csv testdata/*.fit
var fixtures embed.FS

// none is the count of nothing, zero.
const none = 0

// halves is the number of equal parts in a whole that a half is one of.
const halves = 2

// positive is the sign that math.Inf takes to return +Inf.
const positive = +1

// goodFilters is JSON that the readers can be made from. It is shaped
// like the shipped filters and holds the filters the readers need:
//   - W1 to W4, for AllWISE
//   - Energy_Flux100, for Fermi
const goodFilters = `{"filters": [
	{"name": "W1", "from": {"value": 2.8, "unit": "um"},
	 "to": {"value": 3.9, "unit": "um"}},
	{"name": "W2", "from": {"value": 4.0, "unit": "um"},
	 "to": {"value": 5.3, "unit": "um"}},
	{"name": "W3", "from": {"value": 7.5, "unit": "um"},
	 "to": {"value": 16.7, "unit": "um"}},
	{"name": "W4", "from": {"value": 19.7, "unit": "um"},
	 "to": {"value": 26.8, "unit": "um"}},
	{"name": "Energy_Flux100", "from": {"value": 100, "unit": "MeV"},
	 "to": {"value": 100, "unit": "GeV"}}]}`

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

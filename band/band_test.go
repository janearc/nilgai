package band_test

import (
	"testing"

	"github.com/janearc/libtheme-css/primitives/bands"
	"github.com/janearc/nilgai/band"
)

// betweenFixture holds testdata/between.json.
type betweenFixture struct {
	Filters []struct {
		Name        string
		Lo, Hi      Num
		First, Last int
		OnEdge      string
	}
	BadBands []struct {
		Name   string
		Lo, Hi Num
	}
}

// loadBetween reads the between fixture.
func loadBetween(test *testing.T) betweenFixture {
	test.Helper()
	var read betweenFixture
	load(test, "between", &read)
	return read
}

// TestBandsBetween checks that each fixture filter covers the libtheme
// bands the fixture lists. For example, W1 at 2800 nm to 3800 nm starts in
// band 274 and ends in band 285. Two edge cases:
//   - a high edge exactly on a band's low edge stops in the band below
//   - a low edge exactly on a band's low edge starts in that band
func TestBandsBetween(test *testing.T) {
	for _, each := range loadBetween(test).Filters {
		first, last, err := band.BandsBetween(float64(each.Lo),
			float64(each.Hi))
		if err != nil || first != bands.Band(each.First) ||
			last != bands.Band(each.Last) {
			test.Errorf("%s: %d to %d, %v; want %d to %d", each.Name,
				int(first), int(last), err, each.First, each.Last)
		}
	}
}

// TestBandsBetweenEdgesAreExact checks that each edge case has an edge
// exactly on a libtheme band edge. Without that, TestBandsBetween would be
// testing near the edge and not on it.
func TestBandsBetweenEdgesAreExact(test *testing.T) {
	for _, each := range loadBetween(test).Filters {
		if each.OnEdge == "" {
			continue
		}
		edge := map[string]Num{"lo": each.Lo, "hi": each.Hi}[each.OnEdge]
		start, _ := bands.Of(float64(edge)).Edges()
		if Num(start) != edge {
			test.Errorf("%s: %v nm is not a libtheme edge", each.Name, edge)
		}
	}
}

// TestBandsBetweenRefuses checks that each bad pair of edges in the
// fixture is refused and returns band zero for both ends.
func TestBandsBetweenRefuses(test *testing.T) {
	for _, each := range loadBetween(test).BadBands {
		first, last, err := band.BandsBetween(float64(each.Lo),
			float64(each.Hi))
		if err == nil || first != bands.Band(none) ||
			last != bands.Band(none) {
			test.Errorf("%q was accepted: %d to %d", each.Name,
				int(first), int(last))
		}
	}
}

package band

import (
	_ "embed"
	"encoding/json"
	"fmt"

	"github.com/janearc/libtheme-css/primitives/bands"
)

// filterData holds the filters nilgai labels measurements with. Each has
// edges in the units its paper uses, and a citation of that paper. Other
// people measured them, so they live in data and not in code.
//
//go:embed data/filters.json
var filterData []byte

// FilterBands returns the first and last libtheme band of the filter named
// in data/filters.json, such as "W1" or "Energy_Flux100".
func FilterBands(name string) (first, last bands.Band, err error) {
	return ReadFilterBands(filterData, name)
}

// ReadFilterBands finds the named filter in JSON shaped like data/filters.json
// and returns its first and last libtheme bands. Each entry records a name,
// citation, approximate flag, and two edges with values and units; the edges
// may be in either order.
//
// Invalid JSON, missing or duplicate names, unsupported units, or invalid
// bounds return an error. Bounds follow BandsBetween’s contract.
func ReadFilterBands(data []byte, name string) (first, last bands.Band,
	err error) {
	var read struct {
		Filters []struct {
			Name     string
			From, To struct {
				Value float64
				Unit  string
			}
		}
	}
	if err := json.Unmarshal(data, &read); err != nil {
		return first, last, fmt.Errorf("filters: %w", err)
	}
	found := none
	for _, each := range read.Filters {
		if each.Name != name {
			continue
		}
		found++
		first, last, err = filterEdges(each.From.Value, each.From.Unit,
			each.To.Value, each.To.Unit)
		if err != nil {
			return first, last, fmt.Errorf("filter %q: %w",
				name, err)
		}
	}
	if found != one {
		return first, last, fmt.Errorf("filter %q is in the filters "+
			"%d times, not once", name, found)
	}
	return first, last, nil
}

// filterEdges returns the bands between a filter's two edges, each given
// as a value and a unit. The edges may come in either order.
func filterEdges(fromValue float64, fromUnit string, toValue float64,
	toUnit string) (first, last bands.Band, err error) {
	from, err := Quantity{fromValue, fromUnit}.Nanometres()
	if err != nil {
		return first, last, err
	}
	to, err := Quantity{toValue, toUnit}.Nanometres()
	if err != nil {
		return first, last, err
	}
	return BandsBetween(min(from, to), max(from, to))
}

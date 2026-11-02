package read

import (
	"errors"

	"github.com/janearc/nilgai"
)

// cell holds one flag column of a row: its name and its value, verbatim.
type cell struct{ name, value string }

// kept attaches the original flag cells and the reader’s quality assessment to
// a measurement, returning the one-element slice NewSource expects.
//
// An empty flag name or undefined quality indicates a reader programming error:
// callers supply both in code. These validation errors are joined and returned
// with a nil slice.
func kept(measurement nilgai.Measurement, quality nilgai.Quality,
	cells ...cell) ([]nilgai.Measurement, error) {
	flags := make([]nilgai.Flag, len(cells))
	failed := make([]error, len(cells))
	for place, each := range cells {
		flags[place], failed[place] = nilgai.NewFlag(each.name,
			each.value)
	}
	made, err := measurement.WithReading(quality, flags)
	refused := errors.Join(append(failed, err)...)
	var list []nilgai.Measurement
	if refused == nil {
		list = []nilgai.Measurement{*made}
	}
	return list, refused
}
